package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"k8swatch/pkg/event"
	kbEvent "k8swatch/pkg/event"
	
	"github.com/golang/glog"
)

var alertLevel = map[string]string{
	"Scheduled": "pending",
	"Killing": "firing",
}
type Alert struct {
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	GeneratorURL string            `json:"generatorURL"`
}

// Alerts is the end data structure that is converted into JSON and posted to alert managers /api/v1/alerts
// endpoint.
type Alerts []Alert

// AlertManager is the underlying struct used by the alert manager handler receivers
type AlertManager struct {
	url string
}

var alertManagerErrMsg = `
 %s
 
 You need to set alertmanager url for alert manager notify,
 using "--alertmanager http://yourserverip:9093":
 `

// New returns a alert manager handler interface
func (a *AlertManager) Init(alertmanagerurl string) error {
	url := alertmanagerurl
	a.url = url
	return checkMissingAlertManagerVars(a)
}

func (a *AlertManager) ObjectCreated(obj interface{}) {
	e := kbEvent.New(obj)
	if e.Kind == "Pod" {
		if alertLevel[e.Reason] != "" {
			alerts := prepareMsg(e)
			notifyAlertManager(a, alerts)
		}
	}
}

func notifyAlertManager(a *AlertManager, alerts Alerts) {
	
	url := fmt.Sprintf("%v/api/v1/alerts", a.url)
	
	jsonBytes, err := json.Marshal(alerts)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		glog.Fatal(err)
		return
	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		glog.Fatal(err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		glog.Fatalf("Non 200 HTTP response received - %v - %v", resp.StatusCode, resp.Status)
		return
	}
	
	glog.Infof("Message was successfully sent to alertmanager (%s)", url)
}

func checkMissingAlertManagerVars(a *AlertManager) error {
	if a.url == "" {
		return fmt.Errorf(alertManagerErrMsg, "Missing alertmanager url")
	}
	
	return nil
}

func prepareMsg(e event.Event)  Alerts {
	
	labels := make(map[string]string)
	annotations := make(map[string]string)
	labels["namespace"] = e.Namespace
	labels["name"] = e.Name
	labels["reason"] = e.Reason
	labels["kind"] = e.Kind
	labels["message"] = e.Message
	labels["client"] = "k8swatch"
	labels["alertstate"] = alertLevel[e.Reason]
	
	alert := Alert{
		Labels:      labels,
		Annotations: annotations,
	}
	
	alerts := Alerts{alert}
	return alerts
	
}
