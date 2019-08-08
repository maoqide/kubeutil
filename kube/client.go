package kube

import (
	"fmt"
	"log"
	"time"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var defaultDuration = time.Duration(time.Second * 5)

// NewKubeInClusterClient creates an in cluster kubernetes clientset interface
func NewKubeInClusterClient() (clientset.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: %v", err)
	}
	c, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}
	return c, nil
}

// NewKubeOutClusterClient creates a out cluster kubernetes clientset interface
func NewKubeOutClusterClient(config []byte) (clientset.Interface, error) {
	cfg, err := LoadKubeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize inclusterconfig: %v", err)
	}
	clientset, err := clientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}
	return clientset, nil
}

// LoadKubeConfig return *rest.Config from bytes.
func LoadKubeConfig(config []byte) (*rest.Config, error) {
	c, err := clientcmd.Load(config)
	if err != nil {
		log.Fatalf("unable to load config: %v", err)
		return nil, err
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*c, &clientcmd.ConfigOverrides{})
	return clientConfig.ClientConfig()
}

// NewSharedInformerFactory creates a new SharedInformerFactory
func NewSharedInformerFactory(clientset clientset.Interface) (informers.SharedInformerFactory, error) {
	sharedInformers := informers.NewSharedInformerFactory(clientset, defaultDuration)
	return sharedInformers, nil
}
