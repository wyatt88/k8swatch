package utils

import (
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("Can not get Kubernetes config: %v", err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Can not create Kubernetes client: %v", err)
	}
	return clientSet
}

// GetClientOutOfCluster returns a k8s clientSet to the request from outside of cluster
func GetClientOutOfCluster(kubeConfig string, master string) kubernetes.Interface {
	config, err := clientcmd.BuildConfigFromFlags(master, kubeConfig)
	if err != nil {
		glog.Errorf("Can not get kubernetes config from kubeconfig file: %v", err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Can not create kubernetes object because config file has error: %v", err)
	}
	return clientSet
}
