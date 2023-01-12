package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dougbtv/surveyor-cni/pkg/macvlan"
	"github.com/dougbtv/surveyor-cni/pkg/surveyor"
	"github.com/dougbtv/surveyor-cni/pkg/types"
	"github.com/dougbtv/surveyor-cni/pkg/version"

	"github.com/containernetworking/cni/pkg/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	// TODO: This should be replaced by the flag package.
	if len(os.Args) > 1 {
		if os.Args[1] == "i" || os.Args[1] == "introspect" {
			introspect()
			return
		}
	}

	skel.PluginMain(
		cmdAdd,
		nil,
		cmdDel,
		cniVersion.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.4.0"),
		"Surveyor CNI "+version.Version)

}

func introspect() {
	fmt.Printf("hi there\n")
	// TODO: This is static and it should be controlled with a flag, I think, since it can't read a CNI config, right?
	surveyor.CreateInterfaceMap("kube-system")
}

func cmdAdd(args *skel.CmdArgs) error {

	conf, err := types.LoadNetConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("Error parsing CNI configuration \"%s\": %s", args.StdinData, err)
		return err
	}

	if conf.Network == "" {
		return fmt.Errorf("surveyor: 'network' field cannot be empty.")
	}

	if conf.Master != "" {
		return fmt.Errorf("surveyor: 'master' field should not be filled in (saw: %s).", conf.Master)
	}

	// surveyor.WriteToSocket(fmt.Sprintf("Config loaded: %+v", conf), conf)
	// !bang PUT THIS BACK.
	mapping, err := surveyor.GetInterfaceMapping(args, conf)
	if err != nil {
		return err
	}
	// Set the mapping
	conf.Master = mapping
	surveyor.WriteToSocket(fmt.Sprintf("!bang config with mapping: %+v\n", conf), conf)

	// Make the call to macvlan to add this...
	result, err := macvlan.CmdAdd(args, conf)
	if err != nil {
		surveyor.WriteToSocket(fmt.Sprintf("surveyor - macvlan error: %+v\n", err), conf)
		return err
	}

	return cniTypes.PrintResult(result, conf.CNIVersion)
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

func cmdDel(args *skel.CmdArgs) (err error) {
	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting netNS: %s\n", err)
	}
	defer netNS.Close()
	return nil
}
