package kube

import (
	kubeclient "github.com/maoqide/kubeutil/pkg/client"
	"github.com/maoqide/kubeutil/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

var deletePolicy = metav1.DeletePropagationForeground
var commonDeleteOpt = metav1.DeleteOptions{
	GracePeriodSeconds: utils.Int64Ptr(0),
	PropagationPolicy:  &deletePolicy,
}

// Client contains all kube resource client
type Client struct {
	*PodBox
	*EventBox
	*DeploymentBox
	*ServiceBox
	*StatefulSetBox
}

// GetClient get all kube resource client.
func GetClient() (*Client, error) {
	c := kubeclient.Clientset()
	cfg, err := kubeclient.Config()
	if err != nil {
		return nil, err
	}
	cli := Client{
		&PodBox{clientset: *c, config: cfg},
		&EventBox{clientset: *c},
		&DeploymentBox{clientset: *c},
		&ServiceBox{clientset: *c},
		&StatefulSetBox{clientset: *c},
	}
	return &cli, nil
}

// DecodeKubeObj decode kubernetes object from yaml
func DecodeKubeObj(yml []byte) (k8sruntime.Object, *schema.GroupVersionKind, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	return decode(yml, nil, nil)
}
