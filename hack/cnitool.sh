#!/bin/bash
export CNI_PATH=/home/doug/codebase/src/github.com/dougbtv/surveyor-cni/bin
export NETCONFPATH=/tmp/cniconfig/
mkdir -p /tmp/cniconfig

cat << EOF > /tmp/cniconfig/99-test-surveyor.conflist
{
	"cniVersion": "0.4.0",
	"name": "test-surveyor-chain",
	"plugins": [{
		"type": "surveyor",
		"foo": "bar",
		"master": "enp2s0",
		"mode": "bridge",
		"ipam": {
			"type": "static",
			"addresses": [{
				"address": "10.10.0.1/24"
			}]
		}
	}]
}
EOF

sudo ip netns add myplayground
sudo ip netns list | grep myplayground
echo "------------------ CNI ADD"
sudo NETCONFPATH=$(echo $NETCONFPATH) CNI_PATH=$(echo $CNI_PATH) $(which cnitool) add test-surveyor-chain /var/run/netns/myplayground
echo "------------------ CNI DEL"
sudo NETCONFPATH=$(echo $NETCONFPATH) CNI_PATH=$(echo $CNI_PATH) $(which cnitool) del test-surveyor-chain /var/run/netns/myplayground
echo "------------------ Inspection "
sudo ip netns exec myplayground ip a

sudo ip netns del myplayground
echo "----------------------"