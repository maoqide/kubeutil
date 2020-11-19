package cache

import (
	"testing"

	"github.com/maoqide/kubeutil/pkg/client"
	"k8s.io/apimachinery/pkg/labels"
)

func TestCache(t *testing.T) {
	cli := client.Clientset()
	BuildCacheFactory(cli)
	pods, err := Cache().PodLister().List(labels.Everything())
	if err != nil {
		t.FailNow()
	}
	t.Logf("%+v", pods)
}
