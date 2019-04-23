package options

// KubeOptions is options for kubeutil.
type KubeOptions struct {
	Version bool
}

// NewkubeOptions creates a new KubeOptions with default config.
func NewkubeOptions() (*KubeOptions, error) {
	opt := KubeOptions{}
	return &opt, nil
}
