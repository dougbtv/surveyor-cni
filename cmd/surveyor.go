package main

import (
	"fmt"
	"log"
	"os"
	"surveyor-cni/pkg/macvlan"
	"surveyor-cni/pkg/types"
	"surveyor-cni/pkg/version"

	"github.com/containernetworking/cni/pkg/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	cniVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
)

func main() {
	skel.PluginMain(
		cmdAdd,
		nil,
		cmdDel,
		cniVersion.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.4.0"),
		"Surveyor CNI "+version.Version)
}

func cmdAdd(args *skel.CmdArgs) error {

	conf, err := types.LoadNetConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("Error parsing CNI configuration \"%s\": %s", args.StdinData, err)
		return err
	}

	// Create a CNI result to use.
	// result := &current.Result{}
	result, err := macvlan.CmdAdd(args, conf)
	if err != nil {
		debugLogger(fmt.Sprintf("!bang add error: %s", err))
	}

	// surveyor.WriteToSocket(fmt.Sprintf("Config loaded: %+v", conf), conf)
	// !bang PUT THIS BACK.
	// surveyor.GetInterfaceMaps(args, conf)

	/*
		// Let's do the delegation
		binDirs := filepath.SplitList(os.Getenv("CNI_PATH"))
		cniNet := libcni.NewCNIConfig(binDirs, nil)
		// fmt.Printf("bindirs: %+v", binDirs)
		result, err := cniNet.AddNetwork(context.Background(), conf, rt)

		debugLogger(fmt.Sprintf("!bang bindirs: %+v", binDirs))
		debugLogger(fmt.Sprintf("!bang cniNet: %+v", cniNet))
	*/
	/*


		fmt.Printf("!bang cniNet: %+v\n", cniNet)
		fmt.Printf("!bang conf: %+v\n", conf)
	*/

	debugLogger("foo")

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

/*
func confAdd(rt *libcni.RuntimeConf, rawNetconf []byte, multusNetconf *types.NetConf, exec invoke.Exec) (cnitypes.Result, error) {
	logging.Debugf("confAdd: %v, %s", rt, string(rawNetconf))
	// In part, adapted from K8s pkg/kubelet/dockershim/network/cni/cni.go
	binDirs := filepath.SplitList(os.Getenv("CNI_PATH"))
	binDirs = append([]string{multusNetconf.BinDir}, binDirs...)
	cniNet := libcni.NewCNIConfigWithCacheDir(binDirs, multusNetconf.CNIDir, exec)

	conf, err := libcni.ConfFromBytes(rawNetconf)
	if err != nil {
		return nil, logging.Errorf("error in converting the raw bytes to conf: %v", err)
	}

	result, err := cniNet.AddNetwork(context.Background(), conf, rt)
	if err != nil {
		return nil, err
	}

	return result, nil
}
*/

func cmdDel(args *skel.CmdArgs) (err error) {
	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting netNS: %s\n", err)
	}
	defer netNS.Close()
	return nil
}
