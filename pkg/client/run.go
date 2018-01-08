package client

import (
	"github.com/golang/glog"
	"github.com/wyatt88/k8swatch/pkg/config"
	"github.com/wyatt88/k8swatch/pkg/controller"
	"github.com/wyatt88/k8swatch/pkg/handlers"
)

// Run is ok
func Run(kubeConfig string, master string, alertmanagerURL string, resourceConfig config.Resources) {

	var eventHandler handlers.AlertManager
	if err := eventHandler.Init(alertmanagerURL, resourceConfig); err != nil {
		glog.Fatal(err)
	}

	controller.Start(kubeConfig, master, eventHandler)
}
