package initialize

import "fmt"

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
}
