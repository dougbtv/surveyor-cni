package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	cnitypes "github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/040"
	cniVersion "github.com/containernetworking/cni/pkg/version"
)

// NetConf is our definition for the CNI configuration
type NetConf struct {
	cnitypes.NetConf

	// Surveyor params
	Network       string `json:"network"`
	SocketEnabled bool   `json:"socket_enabled"`
	SocketPath    string `json:"socket_path"`
	Kubeconfig    string `json:"kubeconfig"`

	// Macvlan params
	Master        string `json:"master"`
	Mode          string `json:"mode"`
	MTU           int    `json:"mtu"`
	Mac           string `json:"mac,omitempty"`
	RuntimeConfig struct {
		Mac string `json:"mac,omitempty"`
	} `json:"runtimeConfig,omitempty"`
}

type K8sArgs struct {
	cnitypes.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               cnitypes.UnmarshallableString
	K8S_POD_NAMESPACE          cnitypes.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID cnitypes.UnmarshallableString
	K8S_POD_UID                cnitypes.UnmarshallableString
}

// LoadNetConf parses our cni configuration
func LoadNetConf(bytes []byte) (*NetConf, error) {

	// We switch out for the openshift-specific path if we need to.
	// TODO: This could probably be cleaner and more customizable.
	use_kubeconfig_path := "/etc/cni/net.d/surveyor.d/surveyor.kubeconfig"
	if _, err := os.Stat(use_kubeconfig_path); errors.Is(err, os.ErrNotExist) {
		use_kubeconfig_path = "/etc/kubernetes/cni/net.d/surveyor.d/surveyor.kubeconfig"
	}

	conf := NetConf{
		SocketEnabled: true,
		SocketPath:    "/var/run/surveyor-cni/surveyor.sock",
		Kubeconfig:    use_kubeconfig_path,
	}
	if err := json.Unmarshal(bytes, &conf); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %s", err)
	}

	// Parse previous result
	if conf.RawPrevResult != nil {
		resultBytes, err := json.Marshal(conf.RawPrevResult)
		if err != nil {
			return nil, fmt.Errorf("could not serialize prevResult: %v", err)
		}

		res, err := cniVersion.NewResult(conf.CNIVersion, resultBytes)

		if err != nil {
			return nil, fmt.Errorf("could not parse prevResult: %v", err)
		}

		conf.RawPrevResult = nil
		conf.PrevResult, err = current.NewResultFromResult(res)
		if err != nil {
			return nil, fmt.Errorf("could not convert result to current version: %v", err)
		}
	}

	return &conf, nil
}
