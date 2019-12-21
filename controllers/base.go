package controllers

import (
	"fmt"
	"log"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

var (
	KeyFunc = cache.DeletionHandlingMetaNamespaceKeyFunc
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

// WaitForCacheSync is a wrapper around cache.WaitForCacheSync that generates log messages
// indicating that the controller identified by controllerName is waiting for syncs, followed by
// either a successful or failed sync.
func WaitForCacheSync(controllerName string, stopCh <-chan struct{}, cacheSyncs ...cache.InformerSynced) bool {
	log.Printf("Waiting for caches to sync for %s controller", controllerName)
	if !cache.WaitForCacheSync(stopCh, cacheSyncs...) {
		utilruntime.HandleError(fmt.Errorf("unable to sync caches for %s controller", controllerName))
		return false
	}
	log.Printf("Caches are synced for %s controller", controllerName)
	return true
}
