package client

import (
	"k8swatch/pkg/handlers"
	"k8swatch/pkg/controller"
	"github.com/golang/glog"
)

func Run(master string,alertmanagerurl string) {
	
	eventHandler := new(handlers.AlertManager)
	if err := eventHandler.Init(alertmanagerurl); err != nil {
		glog.Fatal(err)
	}
	
	controller.Start(master, *eventHandler)
}

