package main

import (
	"flag"
	c "github.com/wyatt88/k8swatch/pkg/client"
)

func main()  {
	
	var kubeconfig string
	var master string
	//var podname string
	var alertmanagerurl string
	
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "http://127.0.0.1:8080", "master url")
	//flag.StringVar(&podname,"podname","","specific podname you want to watch")
	flag.StringVar(&alertmanagerurl,"alertmanager","http://127.0.0.1:9093","")
	flag.Parse()
	
	// creates the client
	c.Run(master,alertmanagerurl)
	
	
	
	
	
	
	//indexer.Add(&v1.Pod{
	//	ObjectMeta: meta_v1.ObjectMeta{
	//		//Name:      podname,
	//		Name: "kubernetes-dashboard-2843885014-51jp8",
	//		GenerateName: "kubernetes-dashboard-",
	//		Namespace: v1.NamespaceAll,
	//	},
	//})
	
}
