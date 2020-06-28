package kube

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
)

// EventBox provide functions for kubernetes event.
type EventBox struct {
	clientset clientset.Interface
}

// Search search events with labelselectors
func (b *EventBox) Search(namespace string, obj runtime.Object) (*corev1.EventList, error) {
	ref, err := reference.GetReference(scheme.Scheme, obj)
	if err != nil {
		return nil, err
	}
	ref.Kind = ""
	events, err := b.clientset.CoreV1().Events(namespace).Search(scheme.Scheme, ref)
	if err != nil {
		return nil, err
	}
	return events, nil
}
