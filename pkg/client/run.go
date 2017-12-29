package client

import (
	"github.com/golang/glog"
	"github.com/wyatt88/k8swatch/pkg/controller"
	"github.com/wyatt88/k8swatch/pkg/handlers"
)

// Run is ok
func Run(kubeConfig string, master string, alertmanagerURL string) {

	var eventHandler handlers.AlertManager
	if err := eventHandler.Init(alertmanagerURL); err != nil {
		glog.Fatal(err)
	}

	controller.Start(kubeConfig, master, eventHandler)
}
