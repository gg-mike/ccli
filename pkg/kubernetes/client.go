package kubernetes

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func NewInnerClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return &Client{}, err
	}
	return newClient(config)
}

func NewDefaultOuterClient() (*Client, error) {
	return NewOuterClient(filepath.Join(homedir.HomeDir(), ".kube", "config"))
}

func NewOuterClient(kubeconfig string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return &Client{}, err
	}
	return newClient(config)
}

func newClient(config *rest.Config) (*Client, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return &Client{}, err
	}
	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}
