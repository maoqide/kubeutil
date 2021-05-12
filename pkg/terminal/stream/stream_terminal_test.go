package stream_test

import (
	"os"
	"testing"

	"github.com/maoqide/kubeutil/pkg/client"
	"github.com/maoqide/kubeutil/pkg/kube"
	"github.com/maoqide/kubeutil/pkg/terminal/stream"
)

func TestExec(t *testing.T) {
	client.BuildClientset()
	client, err := kube.GetClient()
	if err != nil {
		t.Fatalf("err1: %v", err)
	}

	session := stream.NewTerminalSession(
		stream.IOStreams{
			In:  os.Stdin,
			Out: os.Stdout,
		},
	)

	err = client.PodBox.Exec([]string{"ps", "-ef"},
		session, "default", "nginx-deployment-8d8d4dc86-sqfcx", "nginx")
	if err != nil {
		t.Fatalf("%+v", err)
	}
}
