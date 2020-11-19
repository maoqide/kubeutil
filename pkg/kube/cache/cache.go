package cache

import (
	"sync"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listerappsv1 "k8s.io/client-go/listers/apps/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// full resyc cache resource time
const defaultResyncPeriod = 30 * time.Second

var defaultCachedResources = []string{"Pod"}

var cacheFactory *CacheFactory
var once sync.Once

// CacheFactory provide lister from cache
type CacheFactory struct {
	stopChan              chan struct{}
	sharedInformerFactory informers.SharedInformerFactory
	cachedResources       []string
}

// Cache return instance of CacheFactory
func Cache() *CacheFactory {
	return cacheFactory
}

// BuildCacheFactory build cache factory and start informers
func BuildCacheFactory(client *kubernetes.Interface) {
	once.Do(func() {
		if err := buildCacheFactory(client); err != nil {
			panic(err)
		}
	})
}

func buildCacheFactory(client *kubernetes.Interface) error {
	stop := make(chan struct{})
	sharedInformerFactory := informers.NewSharedInformerFactory(*client, defaultResyncPeriod)
	cf := &CacheFactory{
		stopChan:              stop,
		sharedInformerFactory: sharedInformerFactory,
		cachedResources:       defaultCachedResources,
	}
	cf.initialize()
	sharedInformerFactory.Start(stop)

	if !cache.WaitForCacheSync(stop, sharedInformerFactory.Core().V1().Pods().Informer().HasSynced) {
		panic("WaitForCacheSync failed")
	}
	klog.Infof("cache hasSyncd")

	cacheFactory = cf
	return nil
}

// initialize call Informer() for cachedResources
// so that the informer could run when sharedInformerFactory start
func (c *CacheFactory) initialize() {
	for _, v := range c.cachedResources {
		switch v {
		case "Pod":
			c.sharedInformerFactory.Core().V1().Pods().Informer()
		case "Deployment":
			c.sharedInformerFactory.Apps().V1().Deployments().Informer()
		case "Event":
			c.sharedInformerFactory.Core().V1().Events().Informer()
		case "Node":
			c.sharedInformerFactory.Core().V1().Nodes().Informer()
		default:

		}
	}
}

// PodLister cache
func (c *CacheFactory) PodLister() listercorev1.PodLister {
	return c.sharedInformerFactory.Core().V1().Pods().Lister()
}

// PodIndexer cache
func (c *CacheFactory) PodIndexer() cache.Indexer {
	return c.sharedInformerFactory.Core().V1().Pods().Informer().GetIndexer()
}

// EventLister cache
func (c *CacheFactory) EventLister() listercorev1.EventLister {
	return c.sharedInformerFactory.Core().V1().Events().Lister()
}

// DeploymentLister cache
func (c *CacheFactory) DeploymentLister() listerappsv1.DeploymentLister {
	return c.sharedInformerFactory.Apps().V1().Deployments().Lister()
}

// NodeLister cache
func (c *CacheFactory) NodeLister() listercorev1.NodeLister {
	return c.sharedInformerFactory.Core().V1().Nodes().Lister()
}
