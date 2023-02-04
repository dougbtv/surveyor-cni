# surveyor-cni

<img src="https://github.com/dougbtv/surveyor-cni/blob/main/docs/surveyor_logo.png" width="420" height="420">

A CNI plugin for mapping interfaces to networks (instead of interface names!)

## But, why?

The gist is, if you've got a situation where you have a non-uniform k8s environment, let's say you have one node that has "eth1" and another box with an interface named "eth2", but they're both attached to the same switch, and you go to set a [master parameter on macvlan CNI](https://www.cni.dev/plugins/current/main/macvlan/#network-configuration-reference), you have to create two configurations.

This tool allows you give some meaning to networks. It creates Kubernetes CustomResources which you can use to associate interfaces with networks, and then attach pods to networks -- as opposed to interfaces. An abstraction layer, a way to give some semantics to network attachments.

Best paired with [Multus CNI](https://github.com/k8snetworkplumbingwg/multus-cni)

Currently pre-alpha!

## So, what's it do?

The tool currently create macvlan networks akin to the macvlan plugin (to be expanded later)

## Quickstart

### Requirements

* A running Kubernetes cluster (I'd recommend [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/) if you want to try one!)
* You have Multus CNI installed, you can [use the quickstart guide](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/quickstart.md) to install it.
* You have [the CNI reference plugins](https://github.com/containernetworking/plugins/) installed (semi-optional, but you'll have to update the examples to remove static IPAM CNI)
* Optional: You've tried macvlan + Multus before, which can give you some more context (The Multus quickstart guide above will help you there)

### Installation

### Give it a whirl!

## Trivia

I originally wanted to name the plugin after my favorite [Adirondack](https://en.wikipedia.org/wiki/Adirondack_Mountains) explorer, [Verplanck Colvin](https://en.wikipedia.org/wiki/Verplanck_Colvin), who was a topological engineer who through great adventures helped to survey and map to the Adirondack Mountains which for a long time remained largely unexplored. However, I couldn't figure out a way to make it roll off the tongue, so I stuck with "Surveyor CNI" as it gives a way to kind of survey the equipment you have, and then semantically map it. He also came up with a cool little invention which was basically a pinwheel he'd install on a mountain top so he could get a shine off it on a sunny day to help him sight where the mountain top was exactly.

The logo is created with the help of MidJourney. I realize that these imagery AI's are somewhat controversial, but I was trying to whip something up while travelling and had a low power laptop and no mouse, so I figured I'd give it a try, and tried to prompt MidJourney with what I thought Verplanck might look like while atop the mountain.

## Limitations / What could be done better

* It's a PoC right now! So beware.
* This uses a local fork of the macvlan CNI, it could use [a delegated call](https://pkg.go.dev/github.com/containernetworking/cni/libcni#CNIConfig.AddNetwork) instead.
* That also means, it's basically just macvlan CNI + mapping. It'd be cooler to be more generic.
* I'm unsure how well runtimeconfigs work.
* There's no testing (patches please!)
* There's no automated builds.
