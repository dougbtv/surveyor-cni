package surveyor

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	cnitypes "github.com/containernetworking/cni/pkg/types"

	// current "github.com/containernetworking/cni/pkg/types/040"
	// cniVersion "github.com/containernetworking/cni/pkg/version"

	crdtypes "github.com/dougbtv/surveyor-cni/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/dougbtv/surveyor-cni/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dougbtv/surveyor-cni/pkg/types"

	"github.com/containernetworking/cni/pkg/skel"

	// v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/tools/record"
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

		fmt.Fprintf(os.Stderr, "%s\n", output)

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

func CreateInterfaceMap(namespace string) error {
	ifmapname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Printf("debug - ifmapname: %+v\n", ifmapname)

	ifmap := &crdtypes.InterfaceMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ifmapname,
			Namespace: namespace,
		},
	}

	// Actually let's try to make a list of the interfaces....
	bash_command := `ip a | grep -P "^\d" | grep -vi "veth" | awk '{print $2}' | sed -e 's/:$//'`
	rawbytes, err := exec.Command("/bin/bash", "-c", bash_command).Output()
	if err != nil {
		fmt.Printf("error executing introspection command, dude: %s", err)
		os.Exit(1)
	}
	bashout := string(rawbytes[:])

	// fmt.Printf("debug - bashout: %+v\n", bashout)

	if err != nil {
		log.Fatal(err)
	}

	// Iterate the lines
	for _, line := range strings.Split(strings.TrimSuffix(bashout, "\n"), "\n") {
		newdev := &crdtypes.InterfaceMapSpec{
			Interface: line,
			Network:   "",
		}
		ifmap.Spec = append(ifmap.Spec, *newdev)
	}

	// Create the client.
	client, err := GetK8sClient("")
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.K8sV1().InterfaceMaps(namespace).Create(context.TODO(), ifmap, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("!bang data: %+v\n", result)
	fmt.Printf("!bang ifmap: %+v\n", ifmap)

	return nil
}

func GetK8sClient(kubeconfig string) (*versioned.Clientset, error) {
	var config *rest.Config
	var err error

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

	// Create the client.
	client, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetInterfaceMaps(args *skel.CmdArgs, conf *types.NetConf) (string, error) {

	WriteToSocket(fmt.Sprintf("!bang kubeconfig: %+v\n", conf.Kubeconfig), conf)
	client, err := GetK8sClient(conf.Kubeconfig)
	if err != nil {
		WriteToSocket(fmt.Sprintf("error get client: %+v\n", err), conf)
		return "", err
	}

	// Define the custom resource.
	ifmap := &crdtypes.InterfaceMap{}

	// Set the custom resource name, which is the host we're on.
	ifmapname, err := os.Hostname()
	if err != nil {
		WriteToSocket(fmt.Sprintf("error getting hostname: %+v\n", err), conf)
		return "", err
	}

	// Get the interface map.
	ifmap, err = client.K8sV1().InterfaceMaps(conf.CRDNamespace).Get(context.TODO(), ifmapname, metav1.GetOptions{})
	if err != nil {
		WriteToSocket(fmt.Sprintf("error get cr: %+v\n", err), conf)
		return "", err
	}

	// Print the custom resource.

	WriteToSocket(fmt.Sprintf("!bang Custom Resource: %+v\n", ifmap), conf)
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
