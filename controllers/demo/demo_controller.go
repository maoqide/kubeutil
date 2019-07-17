package demo

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/maoqide/kubeutil/controllers"

	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// DemoController handles pod event
type DemoController struct {
	name string

	queue      workqueue.RateLimitingInterface
	handleFunc func(obj interface{}) error

	// A store of pods, populated by the shared informer passed to NewReplicaSetController
	podLister corelisters.PodLister
	// podListerSynced returns true if the pod store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	podListerSynced cache.InformerSynced
}

// NewDemoController creates a new DemoController
func NewDemoController(podInformer coreinformers.PodInformer) *DemoController {
	dc := &DemoController{
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DemoController"),
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
	}
	dc.handleFunc = defaultHandleFunc
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			dc.HandleAdd(controllers.EVENTADD, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			dc.HandleUpdate(controllers.EVENTUPDATE, oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			dc.HandleDel(controllers.EVENTDELETE, obj)
		},
	})
	return dc
}

// Run begins watching and syncing.
func (dc *DemoController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer dc.queue.ShutDown()
	controllerName := strings.ToLower(dc.name)
	log.Printf("Starting %v controller", controllerName)
	defer log.Printf("Shutting down %v controller", controllerName)

	for i := 0; i < workers; i++ {
		go wait.Until(dc.worker, time.Second, stopCh)
	}
	<-stopCh
}

// HandleAdd handle pods add event
func (dc *DemoController) HandleAdd(etype controllers.EventType, obj interface{}) {
	pod := obj.(*v1.Pod)
	dc.queue.Add(pod)
	if len(pod.Status.ContainerStatuses) > 0 {
		log.Printf("add %v", pod.Status.ContainerStatuses[0].Ready)
	} else {
		log.Printf("add --")
	}
	log.Printf("handle pod add event, %s %s %v", pod.Name, pod.Namespace, pod.DeletionTimestamp)
	return
}

// HandleDel handle pods delete event
func (dc *DemoController) HandleDel(etype controllers.EventType, obj interface{}) {
	pod := obj.(*v1.Pod)
	dc.queue.Add(pod)
	if len(pod.Status.ContainerStatuses) > 0 {
		log.Printf("del %v", pod.Status.ContainerStatuses[0].Ready)
	} else {
		log.Printf("del --")
	}
	log.Printf("handle pod del event, %s %s %v", pod.Name, pod.Namespace, pod.DeletionTimestamp)
	return
}

// HandleUpdate handle pods update event
func (dc *DemoController) HandleUpdate(etype controllers.EventType, oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}
	dc.queue.Add(newPod)
	if len(newPod.Status.ContainerStatuses) > 0 {
		log.Printf("update %v", newPod.Status.ContainerStatuses[0].Ready)
	} else {
		log.Printf("update --")
	}
	log.Printf("handle pod update event, old: %s, new: %s, %s %v", oldPod.Name, newPod.Name, oldPod.Namespace, newPod.DeletionTimestamp)
	return
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (dc *DemoController) worker() {
	for dc.processNextWorkItem() {
	}
}

func (dc *DemoController) processNextWorkItem() bool {
	key, quit := dc.queue.Get()
	if quit {
		return false
	}
	defer dc.queue.Done(key)

	err := dc.handleFunc(key)
	if err == nil {
		dc.queue.Forget(key)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("handle %q failed with %v", key, err))
	dc.queue.AddRateLimited(key)
	return true
}

func defaultHandleFunc(obj interface{}) error {
	pod := obj.(*v1.Pod)
	fmt.Printf("handle %s \n", pod.Name)
	return nil
}
