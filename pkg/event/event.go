package event

import (
	"k8s.io/api/core/v1"
	//"k8s.io/api/apps/v1beta2"
	//batch "k8s.io/api/batch/v1"
)

type Event struct {
	Namespace string
	Kind      string
	Reason    string
	Name      string
	Message   string
}

// New create new Kubewatch Event
func New(obj interface{}) Event {
	var namespace, kind, reason, name, message string
	if apiService, ok := obj.(*v1.Event); ok {
		namespace = apiService.Namespace
		name = apiService.Name
		kind = apiService.InvolvedObject.Kind
		reason = apiService.Reason
		message = apiService.Message
	}
	
	kbEvent := Event{
		Namespace: namespace,
		Kind:      kind,
		Reason:    reason,
		Name:      name,
		Message: message,
	}
	
	return kbEvent
}
