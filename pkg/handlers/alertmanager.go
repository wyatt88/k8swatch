package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"net/http"

	"github.com/wyatt88/k8swatch/pkg/config"
	"github.com/wyatt88/k8swatch/pkg/event"
	kbEvent "github.com/wyatt88/k8swatch/pkg/event"

	"github.com/golang/glog"
)

const (
	alertPushEndpoint = "/api/v1/alerts"
	contentTypeJSON   = "application/json"
)

var podStatus = map[string]string{
	"Scheduled": "warning",
	"Killing":   "firing",
	"Started":   "good",
}

var nodeStatus = map[string]string{
	"NodeNotReady": "firing",
}

var alertLevel = map[string]map[string]string{
	"Pod":  podStatus,
	"Node": nodeStatus,
}

// Alert is ok
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
	url            string
	resourceconfig config.Resources
}

var alertManagerErrMsg = `
 %s
 
 You need to set alertmanager url for alert manager notify,
 using "--alertmanager http://yourserverip:9093":
 `

// Init is new returns a alert manager handler interface
func (a *AlertManager) Init(alertManagerURL string, resourceConf config.Resources) error {
	a.url = alertManagerURL
	a.resourceconfig = resourceConf
	if a.url == "" {
		return fmt.Errorf(alertManagerErrMsg, "Missing alertmanager url")
	}
	return nil
}

// ObjectCreated is ok
func (a *AlertManager) ObjectCreated(obj interface{}) {
	e := kbEvent.New(obj)
	nameList := a.resourceconfig.Resource[e.Kind]
	if nameList != nil || len(nameList) > 0 {
		r, _ := regexp.Compile(strings.Join(nameList, "|"))
		if r.MatchString(e.Name) {
			alerts := prepareMsg(e)
			a.fire(alerts)
		}
	} else {
		return
	}

}

func (a *AlertManager) fire(alerts Alerts) {
	url := a.url + alertPushEndpoint
	jsonBytes, err := json.Marshal(alerts)
	if err != nil {
		glog.Errorf("failed to marshal alerts %v to json: %v", alerts, err)
		return
	}

	resp, err := http.Post(url, contentTypeJSON, bytes.NewBuffer(jsonBytes))
	if err != nil {
		glog.Errorf("failed to post alerts to alertmanager: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		glog.Errorf("Non 200 HTTP response received - %v - %v", resp.StatusCode, resp.Status)
		return
	}
	glog.Infof("Message was successfully sent to alertmanager (%s)", url)
}

func prepareMsg(e *event.Event) Alerts {
	labels := map[string]string{
		"namespace": e.Namespace,
		"name":      e.Name,
		"reason":    e.Reason,
		"kind":      e.Kind,
		"message":   e.Message,
		"client":    "k8swatch",
	}
	if alertLevel[e.Kind] != nil && alertLevel[e.Kind][e.Reason] != "" {
		labels["alertstate"] = alertLevel[e.Kind][e.Reason]
	} else {
		labels["alertstate"] = "pending"
	}
	glog.V(3).Info(e.Reason)
	return Alerts{
		Alert{Labels: labels},
	}
}
