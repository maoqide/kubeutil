package kube

import (
	"log"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
)

// EventType event  type
type EventType int

const (
	// EVENTADD type add
	EVENTADD EventType = iota
	// EVENTUPDATE type update
	EVENTUPDATE
	// EVENTDELETE type delete
	EVENTDELETE
)

// PodHandler handles pod event
type PodHandler struct {
	queue workqueue.RateLimitingInterface
}

// NewPodHandler creates a new PodHandler
func NewPodHandler() *PodHandler {
	return &PodHandler{
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "podhandler"),
	}
}

// HandleAdd handle pods add event
func (handler *PodHandler) HandleAdd(etype EventType, obj interface{}) {
	pod := obj.(*v1.Pod)
	handler.queue.Add(obj)
	log.Printf("handle pod add event, %s %s", pod.Name, pod.Namespace)
	return
}

// HandleDel handle pods delete event
func (handler *PodHandler) HandleDel(etype EventType, obj interface{}) {
	pod := obj.(*v1.Pod)
	log.Printf("handle pod add event, %s %s", pod.Name, pod.Namespace)
	return
}

// HandleUpdate handle pods update event
func (handler *PodHandler) HandleUpdate(etype EventType, oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	log.Printf("handle pod update event, old: %s, new: %s, %s", oldPod.Name, newPod.Name, oldPod.Namespace)
	return
}
