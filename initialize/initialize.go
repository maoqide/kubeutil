package initialize

import (
	"fmt"

	"github.com/maoqide/kubeutil/pkg/client"
	"github.com/maoqide/kubeutil/pkg/kube/cache"
)

var (
	// GitCommit git commit id
	GitCommit = "Unknown"
	// BuildTime build time
	BuildTime = "Unknown"
	// Version v1.0
	Version = "v1.0"
)

func init() {
	fmt.Println("initializing...")
	client.BuildClientset()
	cache.BuildCacheFactory(client.Clientset())
}
