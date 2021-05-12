package copy

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/maoqide/kubeutil/pkg/kube"
	stream_terminal "github.com/maoqide/kubeutil/pkg/terminal/stream"
)

var (
	errFileSpecDoesntMatchFormat = errors.New("filespec must match the canonical format: [[namespace/]pod:]file/path")
	errFileCannotBeEmpty         = errors.New("filepath can not be empty")
)

// New Options
func New(client *kube.Client, namespace, podName, containerName string) Options {
	return Options{
		client:        client,
		podName:       podName,
		namespace:     namespace,
		containerName: containerName,
	}
}

// Options ...
type Options struct {
	client        *kube.Client
	podName       string
	namespace     string
	containerName string
}

// CopyFromPod ...
func (o *Options) CopyFromPod(file string) (io.Reader, string, error) {

	reader, outStream := io.Pipe()
	session := stream_terminal.NewTerminalSession(
		stream_terminal.IOStreams{
			In:  nil,
			Out: outStream,
		})

	fileDir, fileName := filepath.Split(file)
	go func() {
		defer outStream.Close()
		err := o.client.PodBox.Exec(
			[]string{"sh", "-c", fmt.Sprintf("cd %s && tar cf - %s", fileDir, fileName)},
			session, o.namespace, o.podName, o.containerName)
		if err != nil {
		}
	}()
	return reader, fileName, nil
}
