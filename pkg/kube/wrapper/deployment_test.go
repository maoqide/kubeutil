package wrapper_test

import (
	"testing"

	kubewrapper "github.com/maoqide/kubeutil/pkg/kube/wrapper"
)

func TestDeployment(t *testing.T) {
	options := kubewrapper.Options{
		Name:      "test",
		Namespace: "default",
		Image:     "nginx",
		Port:      "80",
	}
	deployment, err := kubewrapper.NewDeploymentWrapper().Create(&options).Complete()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("d: %v", deployment)
}
