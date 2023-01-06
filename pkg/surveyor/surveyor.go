package surveyor

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	cnitypes "github.com/containernetworking/cni/pkg/types"

	// current "github.com/containernetworking/cni/pkg/types/040"
	// cniVersion "github.com/containernetworking/cni/pkg/version"
	crdtypes "surveyor-cni/pkg/apis/k8s.cni.cncf.io/v1"
	"surveyor-cni/pkg/types"
	"time"

	"github.com/containernetworking/cni/pkg/skel"

	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/kubernetes/scheme"
	// v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// "k8s.io/client-go/tools/record"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	chainsawAnnotation = "k8s.v1.cni.cncf.io/chainsaw"
)

// WriteToSocket writes to our socketfile, for logging.
func WriteToSocket(output string, conf *types.NetConf) error {
	if conf.SocketEnabled {

		filestat, err := os.Stat(conf.SocketPath)
		if err != nil {
			return fmt.Errorf("socket file stat failed: %v", err)
		}

		if !filestat.IsDir() {
			if filestat.Mode()&os.ModeSocket == 0 {
				return fmt.Errorf("is not a socket file: %v", err)
			}
		}

		fmt.Fprintf(os.Stderr, "!bang output: %s\n", output)

		handler, err := net.Dial("unix", conf.SocketPath)
		if err != nil {
			return fmt.Errorf("can't open unix socket %v: %v", conf.SocketPath, err)
		}
		defer handler.Close()

		_, err = handler.Write([]byte(output + "\n"))
		if err != nil {
			return fmt.Errorf("socket write error: %v", err)
		}
	}
	return nil
}

// ParseAnnotation parses JSON out of the annotation
func ParseAnnotation(rawannotation string) ([]string, error) {

	var commands []string

	// Parse it if we have JSON.
	if strings.Contains(rawannotation, "[") {
		if err := json.Unmarshal([]byte(rawannotation), &commands); err != nil {
			return nil, fmt.Errorf("failed to parse JSON annotation: %s", err)
		}
	} else {
		// Just parse it as a command.
		commands = append(commands, rawannotation)
	}

	// Cycle through each command and make sure it's legit.
	//
	validationrx, _ := regexp.Compile("^[^\\.\\/][\\w\\s\\.\\:_\\-\\d\\/]+$")
	replaceiprx, _ := regexp.Compile("^\\s*?ip\\s+")
	// r.MatchString("peach")
	for idx, v := range commands {
		if !validationrx.MatchString(v) {
			return nil, fmt.Errorf("We cannot validate the value: '%s' (it's validated like this: https://regex101.com/r/vPKuZC/1)", v)
		}

		// You can use the "ip" name optionally, but we don't want to use the user input.
		if replaceiprx.MatchString(v) {
			commands[idx] = replaceiprx.ReplaceAllString(v, "")
		}
	}

	return commands, nil

}

func GetInterfaceMaps(args *skel.CmdArgs, conf *types.NetConf) (string, error) {

	WriteToSocket(fmt.Sprintf("!bang kubeconfig: %+v\n", conf.Kubeconfig), conf)
	kubeClient, err := GetK8sClient(conf.Kubeconfig, nil)

	// Define the custom resource.
	customResource := &crdtypes.InterfaceMap{}

	// Set the custom resource namespace and name.
	namespace := "default"
	name := "hostfoo"

	// Get the custom resource.
	err = kubeClient.RestClient.Get().
		Namespace(namespace).
		Resource("interfacemaps").
		Name(name).
		Do(context.TODO()).
		Into(customResource)
	if err != nil {
		WriteToSocket(fmt.Sprintf("error get cr: %+v\n", err), conf)
		return "", err
	}

	// Print the custom resource.

	WriteToSocket(fmt.Sprintf("!bang Custom Resource: %+v\n", customResource), conf)
	return "hello", nil
}

// GetK8sArgs gets k8s related args from CNI args
func GetK8sArgs(args *skel.CmdArgs) (*types.K8sArgs, error) {
	k8sArgs := &types.K8sArgs{}

	err := cnitypes.LoadArgs(args.Args, k8sArgs)
	if err != nil {
		return nil, err
	}

	return k8sArgs, nil
}

// ClientInfo contains information given from k8s client
type ClientInfo struct {
	Client     kubernetes.Interface
	RestClient rest.Interface
	// NetClient        netclient.K8sCniCncfIoV1Interface
	// EventBroadcaster record.EventBroadcaster
	// EventRecorder    record.EventRecorder
}

// GetK8sClient gets client info from kubeconfig
func GetK8sClient(kubeconfig string, kubeClient *ClientInfo) (*ClientInfo, error) {
	// logging.Debugf("GetK8sClient: %s, %v", kubeconfig, kubeClient)
	// If we get a valid kubeClient (eg from testcases) just return that
	// one.
	if kubeClient != nil {
		return kubeClient, nil
	}

	var err error
	var config *rest.Config

	// Otherwise try to create a kubeClient from a given kubeConfig
	if kubeconfig != "" {
		// uses the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("GetK8sClient: failed to get context for the kubeconfig %v: %v", kubeconfig, err)
		}
	} else if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		// Try in-cluster config where multus might be running in a kubernetes pod
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("GetK8sClient: failed to get context for in-cluster kube config: %v", err)
		}
	} else {
		// No kubernetes config; assume we shouldn't talk to Kube at all
		return nil, nil
	}

	// Specify that we use gRPC
	config.AcceptContentTypes = "application/vnd.kubernetes.protobuf,application/json"
	config.ContentType = "application/vnd.kubernetes.protobuf"
	// Set the config timeout to one minute.
	config.Timeout = time.Minute

	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	restconfig := *config
	restconfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: crdtypes.GroupName, Version: crdtypes.GroupVersion}
	restconfig.APIPath = "/apis"
	restconfig.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	restconfig.UserAgent = rest.DefaultKubernetesUserAgent()

	rclient, err := rest.UnversionedRESTClientFor(&restconfig)

	return &ClientInfo{
		Client:     client,
		RestClient: rclient,
	}, nil
}
