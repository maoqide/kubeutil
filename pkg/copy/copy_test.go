package copy

import (
	"testing"

	"github.com/maoqide/kubeutil/pkg/client"
	"github.com/maoqide/kubeutil/pkg/kube"
)

func TestCopy(t *testing.T) {
	client.BuildClientset()
	client, err := kube.GetClient()
	if err != nil {
		t.Fatalf("err1: %v", err)
	}
	cpOpt := Options{
		client:        client,
		podName:       "nginx-deployment-8d8d4dc86-sqfcx",
		namespace:     "default",
		containerName: "nginx",
	}
	_, _, err = cpOpt.CopyFromPod("root/ssss")
	if err != nil {
		t.Fatalf("%+v", err)
	}

}
