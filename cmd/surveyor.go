package main

import (
	"fmt"
	"os"
	"surveyor-cni/pkg/surveyor"
	"surveyor-cni/pkg/types"
	"surveyor-cni/pkg/version"

	"github.com/containernetworking/cni/pkg/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/040"
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

	// We try to do as little as possible to get the annotation, and only do more if it has it.
	conf, err := types.LoadNetConf(args.StdinData)
	if err != nil {
		err = fmt.Errorf("Error parsing CNI configuration \"%s\": %s", args.StdinData, err)
		return err
	}

	//

	// var result cniTypes.Result
	// cniresult, err := current.NewResultFromResult(result)
	// if err != nil {
	// 	return err
	// }

	result := &current.Result{}

	// fmt.Printf("foo! !bang")

	// surveyor.WriteToSocket("Check 1212", conf)
	surveyor.GetInterfaceMaps(args, conf)

	/*
		anno, err := surveyor.GetAnnotation(args, conf)
		if err != nil {
			return err
		}

		// We only do the rest if we have an annotation...
		if anno != "" {

			// Figure out the current interface name.
			// We get the last one in the list that has a sandbox
			// surveyor.WriteToSocket(fmt.Sprintf("!bang cniresult: %+v", cniresult.Interfaces), conf)
			currentInterface := ""
			for _, v := range cniresult.Interfaces {
				if v.Sandbox != "" {
					currentInterface = v.Name
				}
			}

			surveyor.WriteToSocket(fmt.Sprintf("!bang =========== ifname: %s / netns: %s", currentInterface, args.Netns), conf)
			// surveyor.WriteToSocket(fmt.Sprintf("!bang anno: %+v", anno), conf)
			commands, err := surveyor.ParseAnnotation(anno)
			if err != nil {
				surveyor.WriteToSocket(fmt.Sprintf("Error parsing command: %v", err), conf)
				return err
			}
			surveyor.WriteToSocket(fmt.Sprintf("Detected commands: %v", commands), conf)
			err = surveyor.ProcessCommands(args.Netns, commands, currentInterface, conf)
			if err != nil {
				return err
			}
		}
	*/

	return cniTypes.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) (err error) {
	netNS, err := ns.GetNS(args.Netns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting netNS: %s\n", err)
	}
	defer netNS.Close()
	return nil
}
