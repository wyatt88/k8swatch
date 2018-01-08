package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	c "github.com/wyatt88/k8swatch/pkg/client"
	config "github.com/wyatt88/k8swatch/pkg/config"
)

func main() {

	var kubeConfig string
	var master string
	var alertmanagerURL string
	var resourceConfigPath string

	flag.StringVar(&kubeConfig, "kubeconfig", "/Users/wenwen/.kube/config", "absolute path to the kubeconfig file")
	// flag.StringVar(&kubeConfig, "kubeconfig", "$HOME/.kube/config", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "http://127.0.0.1:8080", "master url")
	flag.StringVar(&alertmanagerURL, "alertmanager", "http://127.0.0.1:9093", "alertmanager url")
	flag.StringVar(&resourceConfigPath, "resourceconfig", "resources.yml", "path to watching resources config file")
	flag.Parse()

	// creates the client
	if _, err := os.Stat(resourceConfigPath); os.IsNotExist(err) {
		glog.Error("Resource config file doesn't exist")
	}
	resourceConf := config.GetResourceConfig(resourceConfigPath)
	c.Run(kubeConfig, master, alertmanagerURL, resourceConf)
}
