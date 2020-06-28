package kube

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
)

// StatefulSetBox provide functions for kubernetes statefulset.
type StatefulSetBox struct {
	clientset clientset.Interface
}

//NewStatefulSetBoxWithClient creates a statefulsetBox
func NewStatefulSetBoxWithClient(c *clientset.Interface) *StatefulSetBox {
	return &StatefulSetBox{clientset: *c}
}

// Get get specified statefulset in specified namespace.
func (b *StatefulSetBox) Get(name, namespace string) (*appsv1.StatefulSet, error) {
	opt := metav1.GetOptions{}
	return b.clientset.AppsV1().StatefulSets(namespace).Get(name, opt)
}

// List list statefulsets in specified namespace.
func (b *StatefulSetBox) List(namespace string) (*appsv1.StatefulSetList, error) {
	opt := metav1.ListOptions{}
	l, err := b.clientset.AppsV1().StatefulSets(namespace).List(opt)
	return l, err
}

// ListWithSelector list statefulsets in specified namespace.
func (b *StatefulSetBox) ListWithSelector(namespace, labelSelector string) (*appsv1.StatefulSetList, error) {
	opt := metav1.ListOptions{LabelSelector: labelSelector}
	l, err := b.clientset.AppsV1().StatefulSets(namespace).List(opt)
	return l, err
}

// Exists check if statefulset exists.
func (b *StatefulSetBox) Exists(name, namespace string) (bool, error) {
	_, err := b.Get(name, namespace)
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// Create creates a sts
func (b *StatefulSetBox) Create(statefulset *appsv1.StatefulSet, namespace string) (*appsv1.StatefulSet, error) {
	return b.clientset.AppsV1().StatefulSets(namespace).Create(statefulset)
}

// Watch watch sts in specified namespace with timeoutSeconds
func (b *StatefulSetBox) Watch(namespace, labelSelector string, timeoutSeconds *int64) (watch.Interface, error) {
	// labelSelector: example "app", "app=test-app"
	opt := metav1.ListOptions{TimeoutSeconds: timeoutSeconds, LabelSelector: labelSelector}
	w, err := b.clientset.AppsV1().StatefulSets(namespace).Watch(opt)
	return w, err
}

// WatchStatefulSetBox watch specified sts in specified namespace with timeoutSeconds
func (b *StatefulSetBox) WatchStatefulSetBox(namespace, stsName string, timeoutSeconds *int64) (watch.Interface, error) {
	sts, err := b.Get(stsName, namespace)
	if err != nil {
		return nil, err
	}
	opt := metav1.ListOptions{
		TimeoutSeconds:  timeoutSeconds,
		FieldSelector:   fmt.Sprintf("metadata.name=%s", stsName),
		ResourceVersion: sts.ResourceVersion,
	}
	w, err := b.clientset.AppsV1().StatefulSets(namespace).Watch(opt)
	return w, err
}

// Delete delete sts
func (b *StatefulSetBox) Delete(name, namespace string) error {
	opt := commonDeleteOpt
	return b.clientset.AppsV1().StatefulSets(namespace).Delete(name, &opt)
}

// Patch patch sts
func (b *StatefulSetBox) Patch(name, namespace string, data []byte) (*appsv1.StatefulSet, error) {
	return b.clientset.AppsV1().StatefulSets(namespace).Patch(name, patchtypes.StrategicMergePatchType, data)
}

// GetLatestReplicaSet get latest replicaSet of sts
func (b *StatefulSetBox) GetLatestReplicaSet(name, namespace string) (*appsv1.StatefulSet, string, error) {
	sts, err := b.Get(name, namespace)
	if err != nil {
		return nil, "", err
	}
	revision := sts.Annotations["statefulset.kubernetes.io/revision"]
	labelSelector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, "", err
	}

	opt := metav1.ListOptions{LabelSelector: labelSelector.String()}
	replicasets, err := b.clientset.AppsV1().ReplicaSets(namespace).List(opt)
	if err != nil {
		return nil, "", err
	}
	for _, rs := range replicasets.Items {
		if rs.Annotations["statefulset.kubernetes.io/revision"] == revision {
			return sts, rs.Name, nil
		}
	}
	return nil, "", fmt.Errorf("lastest replicaset(that revision corresponding to sts) hasn't been created yet")
}
