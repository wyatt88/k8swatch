package event

import (
	"k8s.io/api/core/v1"
)

// Event is a interface
type Event struct {
	Namespace string
	Kind      string
	Reason    string
	Name      string
	Message   string
}

// New create new k8swatch Event
func New(obj interface{}) *Event {
	if apiService, ok := obj.(*v1.Event); ok {
		return &Event{
			Namespace: apiService.Namespace,
			Name:      apiService.InvolvedObject.Name,
			Kind:      apiService.InvolvedObject.Kind,
			Reason:    apiService.Reason,
			Message:   apiService.Message,
		}
	} else {
		return nil
	}

}
