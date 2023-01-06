# surveyor-cni

A CNI plugin for mapping interfaces to networks (instead of interface names!)

## But, why?

The gist is, if you've got a situation where you have a non-uniform k8s environment, let's say you have one node that has "eth1" and another box with an interface named "eth2", but they're both attached to the same switch, and you go to set a [master parameter on macvlan CNI](https://www.cni.dev/plugins/current/main/macvlan/#network-configuration-reference), you have to create two configurations.

This tool allows you give some meaning to networks. It creates Kubernetes CustomResources which you can use to associate interfaces with networks, and then attach pods to networks -- as opposed to interfaces. An abstraction layer, a way to give some semantics to network attachments.

Best paired with [Multus CNI](https://github.com/k8snetworkplumbingwg/multus-cni)

Currently pre-alpha!

## So, what's it do?

The tool currently create macvlan networks akin to the macvlan plugin (to be expanded later).

...TODO

## Quickstart

...

## What could be done better

...

## Limitations

* This uses a local fork of the macvlan CNI, it could use [a delegated call](https://pkg.go.dev/github.com/containernetworking/cni/libcni#CNIConfig.AddNetwork) instead.
* That also means, it's basically just macvlan CNI + mapping. It'd be cooler to be more generic.
* I'm unsure how well runtimeconfigs work.
* There's no testing (patches please!)