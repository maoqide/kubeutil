package kube

import (
	"bufio"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	terminal "github.com/maoqide/kubeutil/pkg/terminal"
	"github.com/maoqide/kubeutil/utils"
)

// PodBox provide functions for kubernetes pod.
type PodBox struct {
	clientset clientset.Interface
	config    *restclient.Config
}

//NewPodBoxWithClient creates a PodBox
func NewPodBoxWithClient(c *clientset.Interface) *PodBox {
	return &PodBox{clientset: *c}
}

// Get get specified pod in specified namespace.
func (b *PodBox) Get(name, namespace string) (*corev1.Pod, error) {
	opt := metav1.GetOptions{}
	return b.clientset.CoreV1().Pods(namespace).Get(name, opt)
}

// List list pods in specified namespace.
func (b *PodBox) List(namespace, labelSelector string) (*corev1.PodList, error) {
	opt := metav1.ListOptions{LabelSelector: labelSelector}
	return b.clientset.CoreV1().Pods(namespace).List(opt)
}

// Exists check if pod exists.
func (b *PodBox) Exists(name, namespace string) (bool, error) {
	_, err := b.Get(name, namespace)
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// Create creates a pod
func (b *PodBox) Create(pod *corev1.Pod, namespace string) (*corev1.Pod, error) {
	return b.clientset.CoreV1().Pods(namespace).Create(pod)
}

// Watch watch pod in specified namespace with timeoutSeconds
func (b *PodBox) Watch(namespace string, timeoutSeconds *int64, labelSelector string) (watch.Interface, error) {
	opt := metav1.ListOptions{TimeoutSeconds: timeoutSeconds, LabelSelector: labelSelector}
	return b.clientset.CoreV1().Pods(namespace).Watch(opt)
}

// WatchPod watch specified pod in specified namespace with timeoutSeconds
func (b *PodBox) WatchPod(namespace, podName string, timeoutSeconds *int64) (watch.Interface, error) {
	pod, err := b.Get(podName, namespace)
	if err != nil {
		return nil, err
	}
	opt := metav1.ListOptions{
		TimeoutSeconds:  timeoutSeconds,
		FieldSelector:   fmt.Sprintf("metadata.name=%s", podName),
		ResourceVersion: pod.ResourceVersion,
	}
	w, err := b.clientset.CoreV1().Pods(namespace).Watch(opt)
	return w, err
}

// Exec exec into a pod
func (b *PodBox) Exec(cmd []string, ptyHandler terminal.PtyHandler, namespace, podName, containerName string) error {
	defer func() {
		ptyHandler.Done()
	}()

	req := b.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     !(ptyHandler.Stdin() == nil),
		Stdout:    !(ptyHandler.Stdout() == nil),
		Stderr:    !(ptyHandler.Stderr() == nil),
		TTY:       ptyHandler.Tty(),
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(b.config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler.Stdin(),
		Stdout:            ptyHandler.Stdout(),
		Stderr:            ptyHandler.Stderr(),
		TerminalSizeQueue: ptyHandler,
		Tty:               ptyHandler.Tty(),
	})
	return err
}

// Logs get logs of specified pod in specified namespace.
func (b *PodBox) Logs(name, namespace string, opts *corev1.PodLogOptions) *restclient.Request {
	return b.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
}

// LogStream get logs of specified pod in specified namespace and copy to writer.
func (b *PodBox) LogStream(name, namespace string, opts *corev1.PodLogOptions, writer io.Writer) error {
	req := b.Logs(name, namespace, opts)
	r, err := req.Stream()
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(writer, r)
	return err
}

// LogStreamLine get logs of specified pod in specified namespace and copy to writer.
func (b *PodBox) LogStreamLine(name, namespace string, opts *corev1.PodLogOptions, writer io.Writer) error {
	req := b.Logs(name, namespace, opts)
	r, err := req.Stream()
	if err != nil {
		return err
	}
	defer r.Close()
	bufReader := bufio.NewReaderSize(r, 256)
	// bufReader := bufio.NewReader(r)
	for {
		line, _, err := bufReader.ReadLine()
		// line = []byte(fmt.Sprintf("%s", string(line)))
		line = utils.ToValidUTF8(line, []byte(""))
		if err != nil {
			if err == io.EOF {
				_, err = writer.Write(line)
			}
			return err
		}
		// line = append(line, []byte("\r\n")...)
		// line = append(bytes.Trim(line, " "), []byte("\r\n")...)
		_, err = writer.Write(line)
		if err != nil {
			return err
		}
	}
}

// Delete delete pod
func (b *PodBox) Delete(name, namespace string) error {
	opt := commonDeleteOpt
	return b.clientset.CoreV1().Pods(namespace).Delete(name, &opt)
}
