// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package macvlan

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	surveyortypes "surveyor-cni/pkg/types"

	"github.com/vishvananda/netlink"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
)

/*
type NetConf struct {
	types.NetConf
	Master string `json:"master"`
	Mode   string `json:"mode"`
	MTU    int    `json:"mtu"`
	Mac    string `json:"mac,omitempty"`

	RuntimeConfig struct {
		Mac string `json:"mac,omitempty"`
	} `json:"runtimeConfig,omitempty"`
}
*/

// MacEnvArgs represents CNI_ARG
type MacEnvArgs struct {
	types.CommonArgs
	MAC types.UnmarshallableString `json:"mac,omitempty"`
}

func debugLogger(s string) {
	f, err := os.OpenFile("/tmp/debug.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s + "\n"); err != nil {
		log.Println(err)
	}
}

/*
func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}
*/

func getDefaultRouteInterfaceName() (string, error) {
	routeToDstIP, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return "", err
	}

	for _, v := range routeToDstIP {
		if v.Dst == nil {
			l, err := netlink.LinkByIndex(v.LinkIndex)
			if err != nil {
				return "", err
			}
			return l.Attrs().Name, nil
		}
	}

	return "", fmt.Errorf("no default route interface found")
}

func loadConf(bytes []byte, envArgs string) (*surveyortypes.NetConf, string, error) {
	n := &surveyortypes.NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	if n.Master == "" {
		defaultRouteInterface, err := getDefaultRouteInterfaceName()
		if err != nil {
			return nil, "", err
		}
		n.Master = defaultRouteInterface
	}

	// check existing and MTU of master interface
	masterMTU, err := getMTUByName(n.Master)
	if err != nil {
		return nil, "", err
	}
	if n.MTU < 0 || n.MTU > masterMTU {
		return nil, "", fmt.Errorf("invalid MTU %d, must be [0, master MTU(%d)]", n.MTU, masterMTU)
	}

	if envArgs != "" {
		e := MacEnvArgs{}
		err := types.LoadArgs(envArgs, &e)
		if err != nil {
			return nil, "", err
		}

		if e.MAC != "" {
			n.Mac = string(e.MAC)
		}
	}

	if n.RuntimeConfig.Mac != "" {
		n.Mac = n.RuntimeConfig.Mac
	}

	return n, n.CNIVersion, nil
}

func validateConf(n *surveyortypes.NetConf) (*surveyortypes.NetConf, error) {

	if n.Master == "" {
		defaultRouteInterface, err := getDefaultRouteInterfaceName()
		if err != nil {
			return n, err
		}
		n.Master = defaultRouteInterface
	}

	// check existing and MTU of master interface
	masterMTU, err := getMTUByName(n.Master)
	if err != nil {
		return n, err
	}
	if n.MTU < 0 || n.MTU > masterMTU {
		return n, fmt.Errorf("invalid MTU %d, must be [0, master MTU(%d)]", n.MTU, masterMTU)
	}

	if n.RuntimeConfig.Mac != "" {
		n.Mac = n.RuntimeConfig.Mac
	}

	return n, nil
}

func getMTUByName(ifName string) (int, error) {
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return 0, err
	}
	return link.Attrs().MTU, nil
}

func modeFromString(s string) (netlink.MacvlanMode, error) {
	switch s {
	case "", "bridge":
		return netlink.MACVLAN_MODE_BRIDGE, nil
	case "private":
		return netlink.MACVLAN_MODE_PRIVATE, nil
	case "vepa":
		return netlink.MACVLAN_MODE_VEPA, nil
	case "passthru":
		return netlink.MACVLAN_MODE_PASSTHRU, nil
	default:
		return 0, fmt.Errorf("unknown macvlan mode: %q", s)
	}
}

func modeToString(mode netlink.MacvlanMode) (string, error) {
	switch mode {
	case netlink.MACVLAN_MODE_BRIDGE:
		return "bridge", nil
	case netlink.MACVLAN_MODE_PRIVATE:
		return "private", nil
	case netlink.MACVLAN_MODE_VEPA:
		return "vepa", nil
	case netlink.MACVLAN_MODE_PASSTHRU:
		return "passthru", nil
	default:
		return "", fmt.Errorf("unknown macvlan mode: %q", mode)
	}
}

func createMacvlan(conf *surveyortypes.NetConf, ifName string, netns ns.NetNS) (*current.Interface, error) {
	macvlan := &current.Interface{}

	mode, err := modeFromString(conf.Mode)
	if err != nil {
		return nil, err
	}

	m, err := netlink.LinkByName(conf.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", conf.Master, err)
	}

	// due to kernel bug we have to create with tmpName or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	linkAttrs := netlink.LinkAttrs{
		MTU:         conf.MTU,
		Name:        tmpName,
		ParentIndex: m.Attrs().Index,
		Namespace:   netlink.NsFd(int(netns.Fd())),
	}

	if conf.Mac != "" {
		addr, err := net.ParseMAC(conf.Mac)
		if err != nil {
			return nil, fmt.Errorf("invalid args %v for MAC addr: %v", conf.Mac, err)
		}
		linkAttrs.HardwareAddr = addr
	}

	mv := &netlink.Macvlan{
		LinkAttrs: linkAttrs,
		Mode:      mode,
	}

	if err := netlink.LinkAdd(mv); err != nil {
		return nil, fmt.Errorf("failed to create macvlan: %v", err)
	}

	err = netns.Do(func(_ ns.NetNS) error {
		err := ip.RenameLink(tmpName, ifName)
		if err != nil {
			_ = netlink.LinkDel(mv)
			return fmt.Errorf("failed to rename macvlan to %q: %v", ifName, err)
		}
		macvlan.Name = ifName

		// Re-fetch macvlan to get all properties/attributes
		contMacvlan, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch macvlan %q: %v", ifName, err)
		}
		macvlan.Mac = contMacvlan.Attrs().HardwareAddr.String()
		macvlan.Sandbox = netns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return macvlan, nil
}

func CmdAdd(args *skel.CmdArgs, n *surveyortypes.NetConf) (*current.Result, error) {
	debugLogger("!trace a")

	result := &current.Result{}

	/*
		n, _, err := loadConf(args.StdinData, args.Args)
		if err != nil {
			return result, err
		}
	*/

	n, err := validateConf(n)
	if err != nil {
		return result, fmt.Errorf("config validation failed: %s", err)
	}

	debugLogger("!trace b")

	isLayer3 := n.IPAM.Type != ""

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return result, fmt.Errorf("failed to open netns %q: %v", netns, err)
	}
	defer netns.Close()

	macvlanInterface, err := createMacvlan(n, args.IfName, netns)
	if err != nil {
		return result, err
	}

	// Delete link if err to avoid link leak in this ns
	defer func() {
		if err != nil {
			netns.Do(func(_ ns.NetNS) error {
				return ip.DelLinkByName(args.IfName)
			})
		}
	}()

	debugLogger("!trace c")

	// Assume L2 interface only
	result = &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{macvlanInterface},
	}

	if isLayer3 {
		// run the IPAM plugin and get back the config to apply
		r, err := ipam.ExecAdd(n.IPAM.Type, args.StdinData)
		if err != nil {
			return result, err
		}

		debugLogger("!trace d")

		// Invoke ipam del if err to avoid ip leak
		defer func() {
			if err != nil {
				ipam.ExecDel(n.IPAM.Type, args.StdinData)
			}
		}()

		// Convert whatever the IPAM result was into the current Result type
		ipamResult, err := current.NewResultFromResult(r)
		if err != nil {
			return result, err
		}

		if len(ipamResult.IPs) == 0 {
			return result, errors.New("IPAM plugin returned missing IP config")
		}

		result.IPs = ipamResult.IPs
		result.Routes = ipamResult.Routes

		for _, ipc := range result.IPs {
			// All addresses apply to the container macvlan interface
			ipc.Interface = current.Int(0)
		}

		err = netns.Do(func(_ ns.NetNS) error {
			_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv4/conf/%s/arp_notify", args.IfName), "1")

			if err := ipam.ConfigureIface(args.IfName, result); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return result, err
		}
	} else {
		// For L2 just change interface status to up
		err = netns.Do(func(_ ns.NetNS) error {
			macvlanInterfaceLink, err := netlink.LinkByName(args.IfName)
			if err != nil {
				return fmt.Errorf("failed to find interface name %q: %v", macvlanInterface.Name, err)
			}

			if err := netlink.LinkSetUp(macvlanInterfaceLink); err != nil {
				return fmt.Errorf("failed to set %q UP: %v", args.IfName, err)
			}

			return nil
		})
		if err != nil {
			return result, err
		}
	}

	debugLogger("!trace e")

	result.DNS = n.DNS

	return result, nil
}

func cmdDel(args *skel.CmdArgs) error {
	n, _, err := loadConf(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	isLayer3 := n.IPAM.Type != ""

	if isLayer3 {
		err = ipam.ExecDel(n.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}
	}

	if args.Netns == "" {
		return nil
	}

	// There is a netns so try to clean up. Delete can be called multiple times
	// so don't return an error if the device is already removed.
	err = ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		if err := ip.DelLinkByName(args.IfName); err != nil {
			if err != ip.ErrLinkNotFound {
				return err
			}
		}
		return nil
	})

	if err != nil {
		//  if NetNs is passed down by the Cloud Orchestration Engine, or if it called multiple times
		// so don't return an error if the device is already removed.
		// https://github.com/kubernetes/kubernetes/issues/43014#issuecomment-287164444
		_, ok := err.(ns.NSPathNotExistErr)
		if ok {
			return nil
		}
		return err
	}

	return err
}

/*
func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("macvlan"))
}
*/

func cmdCheck(args *skel.CmdArgs) error {

	n, _, err := loadConf(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	isLayer3 := n.IPAM.Type != ""

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	if isLayer3 {
		// run the IPAM plugin and get back the config to apply
		err = ipam.ExecCheck(n.IPAM.Type, args.StdinData)
		if err != nil {
			return err
		}
	}

	// Parse previous result.
	if n.NetConf.RawPrevResult == nil {
		return fmt.Errorf("Required prevResult missing")
	}

	if err := version.ParsePrevResult(&n.NetConf); err != nil {
		return err
	}

	result, err := current.NewResultFromResult(n.PrevResult)
	if err != nil {
		return err
	}

	var contMap current.Interface
	// Find interfaces for names whe know, macvlan device name inside container
	for _, intf := range result.Interfaces {
		if args.IfName == intf.Name {
			if args.Netns == intf.Sandbox {
				contMap = *intf
				continue
			}
		}
	}

	// The namespace must be the same as what was configured
	if args.Netns != contMap.Sandbox {
		return fmt.Errorf("Sandbox in prevResult %s doesn't match configured netns: %s",
			contMap.Sandbox, args.Netns)
	}

	m, err := netlink.LinkByName(n.Master)
	if err != nil {
		return fmt.Errorf("failed to lookup master %q: %v", n.Master, err)
	}

	// Check prevResults for ips, routes and dns against values found in the container
	if err := netns.Do(func(_ ns.NetNS) error {

		// Check interface against values found in the container
		err := validateCniContainerInterface(contMap, m.Attrs().Index, n.Mode)
		if err != nil {
			return err
		}

		err = ip.ValidateExpectedInterfaceIPs(args.IfName, result.IPs)
		if err != nil {
			return err
		}

		err = ip.ValidateExpectedRoute(result.Routes)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func validateCniContainerInterface(intf current.Interface, parentIndex int, modeExpected string) error {

	var link netlink.Link
	var err error

	if intf.Name == "" {
		return fmt.Errorf("Container interface name missing in prevResult: %v", intf.Name)
	}
	link, err = netlink.LinkByName(intf.Name)
	if err != nil {
		return fmt.Errorf("Container Interface name in prevResult: %s not found", intf.Name)
	}
	if intf.Sandbox == "" {
		return fmt.Errorf("Error: Container interface %s should not be in host namespace", link.Attrs().Name)
	}

	macv, isMacvlan := link.(*netlink.Macvlan)
	if !isMacvlan {
		return fmt.Errorf("Error: Container interface %s not of type macvlan", link.Attrs().Name)
	}

	mode, err := modeFromString(modeExpected)
	if macv.Mode != mode {
		currString, err := modeToString(macv.Mode)
		if err != nil {
			return err
		}
		confString, err := modeToString(mode)
		if err != nil {
			return err
		}
		return fmt.Errorf("Container macvlan mode %s does not match expected value: %s", currString, confString)
	}

	if intf.Mac != "" {
		if intf.Mac != link.Attrs().HardwareAddr.String() {
			return fmt.Errorf("Interface %s Mac %s doesn't match container Mac: %s", intf.Name, intf.Mac, link.Attrs().HardwareAddr)
		}
	}

	return nil
}
