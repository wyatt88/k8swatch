package main

import (
	"flag"

	c "github.com/wyatt88/k8swatch/pkg/client"
)

func main() {

	var kubeConfig string
	var master string
	var alertmanagerURL string

	flag.StringVar(&kubeConfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "http://127.0.0.1:8080", "master url")
	flag.StringVar(&alertmanagerURL, "alertmanager", "http://127.0.0.1:9093", "alertmanager url")
	flag.Parse()

	// creates the client
	c.Run(kubeConfig, alertmanagerURL)
}
