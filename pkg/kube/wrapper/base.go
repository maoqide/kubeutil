package wrapper

import k8sruntime "k8s.io/apimachinery/pkg/runtime"

// Options template options for containers
type Options struct {
	Name      string
	Namespace string
	Image     string
	Port      string
}

// Wrapper  wrapper interface for kubernetes resource
type Wrapper interface {
	Create() *Wrapper
	Validate() bool
	Err() error
	Complete() (k8sruntime.Object, error)
}
