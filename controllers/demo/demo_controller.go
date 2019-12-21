package demo

import (
	"fmt"
	"log"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/maoqide/kubeutil/controllers"
)

type eventObj struct {
	ObjType string
	Key     string
}

const (
	objectTypePod         = "Pod"
	objectTypeDeployment  = "Deployment"
	objectTypeStatefulSet = "StatefulSet"
)

// Controller handles pod event
type Controller struct {
	name string

	queue      workqueue.RateLimitingInterface
	handleFunc func(obj interface{}) error
	enqueueObj func(obj interface{})

	podIndexer         cache.Indexer
	deploymentIndexer  cache.Indexer
	statefulsetIndexer cache.Indexer
	// A store of pods, populated by the shared informer passed to NewReplicaSetController
	podLister corelisters.PodLister
	// podListerSynced returns true if the pod store has been synced at least once.
	// Added as a member to the struct to allow injection for testing.
	podListerSynced         cache.InformerSynced
	deploymentListerSynced  cache.InformerSynced
	statefulsetListerSynced cache.InformerSynced
}

// NewDemoController creates a new DemoController
func NewDemoController(podInformer coreinformers.PodInformer, deployInformer appsinformers.DeploymentInformer, statefulsetInformer appsinformers.StatefulSetInformer) *Controller {
	dc := &Controller{
		name:                    "DemoController",
		queue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "DemoController"),
		podLister:               podInformer.Lister(),
		podListerSynced:         podInformer.Informer().HasSynced,
		deploymentListerSynced:  deployInformer.Informer().HasSynced,
		statefulsetListerSynced: statefulsetInformer.Informer().HasSynced,
		podIndexer:              podInformer.Informer().GetIndexer(),
		deploymentIndexer:       deployInformer.Informer().GetIndexer(),
		statefulsetIndexer:      statefulsetInformer.Informer().GetIndexer(),
	}

	dc.handleFunc = dc.defaultHandleFunc
	dc.enqueueObj = dc.enqueue
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			dc.addPod(
				controllers.EVENTADD, obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			dc.updatePod(controllers.EVENTUPDATE, oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			dc.deletePod(controllers.EVENTDELETE, obj)
		},
	})
	deployInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			dc.enqueueObj(deployment)
			log.Printf("receive Deployment delele event, %s %s", deployment.Name, deployment.Namespace)
		},
	})
	statefulsetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			sts := obj.(*appsv1.StatefulSet)
			dc.enqueueObj(sts)
			log.Printf("receive StatefulSet delete event, %s %s", sts.Name, sts.Namespace)
		},
	})

	return dc
}

// Run begins watching and syncing.
func (dc *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer dc.queue.ShutDown()
	controllerName := strings.ToLower(dc.name)
	log.Printf("Starting %v controller", controllerName)
	defer log.Printf("Shutting down %s controller", controllerName)

	if !controllers.WaitForCacheSync(dc.name, stopCh, dc.podListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(dc.worker, time.Second, stopCh)
	}
	<-stopCh
}

// addPod handle pods add event
func (dc *Controller) addPod(etype controllers.EventType, obj interface{}) {
	pod := obj.(*v1.Pod)
	dc.enqueueObj(pod)
	log.Printf("receive Pod add event, %s %s", pod.Name, pod.Namespace)
	return
}

// deletePod pods delete event
func (dc *Controller) deletePod(etype controllers.EventType, obj interface{}) {
	pod := obj.(*v1.Pod)
	dc.enqueueObj(pod)
	log.Printf("receive Pod del event, %s %s", pod.Name, pod.Namespace)
	return
}

// updatePod handle pods update event
func (dc *Controller) updatePod(etype controllers.EventType, oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		return
	}
	dc.enqueueObj(newPod)
	log.Printf("receive Pod update event, old: %s, new: %s, %s", oldPod.Name, newPod.Name, oldPod.Namespace)
	return
}

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (dc *Controller) worker() {
	for dc.processNextWorkItem() {
	}
}

func (dc *Controller) processNextWorkItem() bool {
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

func (dc *Controller) defaultHandleFunc(key interface{}) error {
	// namespace, name, err := cache.SplitMetaNamespaceKey(key.(string))
	// if err != nil {
	// 	return err
	// }
	event := key.(eventObj)
	switch event.ObjType {
	case objectTypePod:
		return dc.handlePod(event.Key)
	case objectTypeDeployment:
		return dc.handleDeployment(event.Key)
	case objectTypeStatefulSet:
		return dc.handleStatefulset(event.Key)
	default:
		log.Printf("receive unexcepted object type: %s!", event.ObjType)
	}
	return nil
}

func (dc *Controller) enqueue(obj interface{}) {
	objType := ""
	switch obj.(type) {
	case *corev1.Pod:
		objType = objectTypePod
	case *appsv1.Deployment:
		objType = objectTypeDeployment
	case *appsv1.DaemonSet:
		objType = objectTypeStatefulSet
	}
	key, err := controllers.KeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Couldn't get key for object %#v: %v", obj, err))
		return
	}
	event := eventObj{
		ObjType: objType,
		Key:     key,
	}
	dc.queue.Add(event)
}

func (dc *Controller) handlePod(key string) error {
	obj, exists, err := dc.podIndexer.GetByKey(key)
	if err != nil || !exists {
		return err
	}
	pod := obj.(*corev1.Pod)
	if pod.DeletionTimestamp != nil {
		log.Printf("pod [%s] has been deteled, ignored.", pod.Name)
		return nil
	}
	return nil
}

func (dc *Controller) handleDeployment(key string) error {
	obj, exists, err := dc.deploymentIndexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		// ...
		return err
	}
	deployment := obj.(*appsv1.Deployment)
	if deployment.DeletionTimestamp != nil {
		log.Printf("deployment [%s] is being deleted now, namespace: %s.", deployment.Name, deployment.Namespace)
		// ...
		return nil
	}
	// ...
	return nil
}

func (dc *Controller) handleStatefulset(key string) error {
	obj, exists, err := dc.statefulsetIndexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return err
		}
		log.Printf("statefulSet [%s] not exists now, namespace: %s.", name, namespace)
		// ...
		return nil
	}
	sts := obj.(*appsv1.StatefulSet)
	if sts.DeletionTimestamp != nil {
		log.Printf("statefulSet [%s] is being deleted now, namespace: %s.", sts.Name, sts.Namespace)
		// ...
		return nil
	}
	return nil
}
