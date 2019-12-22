package kube_test

import (
	"testing"

	"github.com/maoqide/kubeutil/pkg/kube"
	kubewrapper "github.com/maoqide/kubeutil/pkg/kube/wrapper"
)

func TestDeployment(t *testing.T) {
	client, err := kube.GetClient()
	if err != nil {
		t.Fatalf("err1: %v", err)
	}
	options := kubewrapper.Options{
		Name:      "test",
		Namespace: "default",
		Image:     "nginx",
		Port:      "80",
	}
	deployment, err := kubewrapper.NewDeploymentWrapper().Create(&options).Complete()
	if err != nil {
		t.Fatalf("err2: %v", err)
	}
	_, err = client.DeploymentBox.Create(deployment, "default")
	if err != nil {
		t.Fatalf("err3: %v", err)
	}
	d, err := client.DeploymentBox.Get(deployment.Name, "default")
	if err != nil {
		t.Fatalf("err4: %v", err)
	}
	t.Logf("----- %v", d)
}
