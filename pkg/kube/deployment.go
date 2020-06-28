package kube

import (
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// DeploymentBox provide functions for kubernetes deployment.
type DeploymentBox struct {
	clientset clientset.Interface
}

//NewDeploymentBoxWithClient creates a DeploymentBox
func NewDeploymentBoxWithClient(c *clientset.Interface) *DeploymentBox {
	return &DeploymentBox{clientset: *c}
}

// Get get specified deployment in specified namespace.
func (b *DeploymentBox) Get(name, namespace string) (*appsv1.Deployment, error) {
	opt := metav1.GetOptions{}
	return b.clientset.AppsV1().Deployments(namespace).Get(name, opt)
}

// List list deployments in specified namespace.
func (b *DeploymentBox) List(namespace string) (*appsv1.DeploymentList, error) {
	opt := metav1.ListOptions{}
	l, err := b.clientset.AppsV1().Deployments(namespace).List(opt)
	return l, err
}

// Exists check if deployment exists.
func (b *DeploymentBox) Exists(name, namespace string) (bool, error) {
	_, err := b.Get(name, namespace)
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// Create creates a deployment
func (b *DeploymentBox) Create(deployment *appsv1.Deployment, namespace string) (*appsv1.Deployment, error) {
	return b.clientset.AppsV1().Deployments(namespace).Create(deployment)
}

// Watch watch deployment in specified namespace with timeoutSeconds
func (b *DeploymentBox) Watch(namespace, labelSelector string, timeoutSeconds *int64) (watch.Interface, error) {
	// labelSelector: example "app", "app=test-app"
	opt := metav1.ListOptions{TimeoutSeconds: timeoutSeconds, LabelSelector: labelSelector}
	w, err := b.clientset.AppsV1().Deployments(namespace).Watch(opt)
	if apierrors.IsNotFound(err) {
		// for kubernetes cluster with low version.
		w, err = b.clientset.ExtensionsV1beta1().Deployments(namespace).Watch(opt)
	}

	return w, err
}

// WatchDeployment watch specified deployment in specified namespace with timeoutSeconds
func (b *DeploymentBox) WatchDeployment(namespace, deploymentName string, timeoutSeconds *int64) (watch.Interface, error) {
	deploy, err := b.Get(deploymentName, namespace)
	if err != nil {
		return nil, err
	}
	opt := metav1.ListOptions{
		TimeoutSeconds:  timeoutSeconds,
		FieldSelector:   fmt.Sprintf("metadata.name=%s", deploymentName),
		ResourceVersion: deploy.ResourceVersion,
	}
	w, err := b.clientset.AppsV1().Deployments(namespace).Watch(opt)
	if apierrors.IsNotFound(err) {
		// for kubernetes cluster with low version.
		w, err = b.clientset.ExtensionsV1beta1().Deployments(namespace).Watch(opt)
	}

	return w, err
}

// Delete delete deployment
func (b *DeploymentBox) Delete(name, namespace string) error {
	opt := commonDeleteOpt
	return b.clientset.AppsV1().Deployments(namespace).Delete(name, &opt)
}

// Patch patch deployment
func (b *DeploymentBox) Patch(name, namespace string, data []byte) (*appsv1.Deployment, error) {
	return b.clientset.AppsV1().Deployments(namespace).Patch(name, patchtypes.StrategicMergePatchType, data)
}

// Scale scale deployment replicas
func (b *DeploymentBox) Scale(name, namespace string, replicas int32) error {
	scale, err := b.clientset.AppsV1().Deployments(namespace).GetScale(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas

	// retry in case of OptimisticLockErrorMsg when resourceversion changed before scale
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		scale, err := b.clientset.AppsV1().Deployments(namespace).GetScale(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		scale.Spec.Replicas = replicas
		_, err = b.clientset.AppsV1().Deployments(namespace).UpdateScale(name, scale)
		return err
	}); err != nil {
		return err
	}
	// _, err = b.clientset.AppsV1().Deployments(namespace).UpdateScale(name, scale)
	return err
}

// GetLatestReplicaSet get latest replicaSet of deployment
func (b *DeploymentBox) GetLatestReplicaSet(name, namespace string) (*appsv1.Deployment, string, error) {
	deployment, err := b.Get(name, namespace)
	if err != nil {
		return nil, "", err
	}
	revision := deployment.Annotations["deployment.kubernetes.io/revision"]
	labelSelector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, "", err
	}

	opt := metav1.ListOptions{LabelSelector: labelSelector.String()}
	replicasets, err := b.clientset.AppsV1().ReplicaSets(namespace).List(opt)
	if err != nil {
		return nil, "", err
	}
	for _, rs := range replicasets.Items {
		if rs.Annotations["deployment.kubernetes.io/revision"] == revision {
			return deployment, rs.Name, nil
		}
	}
	return nil, "", fmt.Errorf("lastest replicaset(that revision corresponding to deployment) hasn't been created yet")
}

// GetPods get pods of deployment
func (b *DeploymentBox) GetPods(name, namespace string) (*corev1.PodList, error) {
	deployment, err := b.Get(name, namespace)
	if err != nil {
		return nil, err
	}
	labelSelector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	opt := metav1.ListOptions{LabelSelector: labelSelector.String()}
	podList, err := b.clientset.CoreV1().Pods(namespace).List(opt)
	return podList, nil
}

// PatchImage reutn bytes for a StrategicMergePatch of deployment
func PatchImage(deployment *appsv1.Deployment, image string) ([]byte, error) {
	curJSON, err := json.Marshal(deployment)
	if err != nil {
		return []byte{}, err
	}
	modDeployment := *deployment
	if image != "" {
		modDeployment.Spec.Template.Spec.Containers[0].Image = image
	}
	modDeployment.Spec.Template.Labels["last_update"] = fmt.Sprintf("%d", time.Now().Unix())
	var UpdateTimeEnvExist = false
	for i, e := range modDeployment.Spec.Template.Spec.Containers[0].Env {
		if e.Name == "INSTANCE_LAST_UPDATE" {
			modDeployment.Spec.Template.Spec.Containers[0].Env[i] =
				corev1.EnvVar{Name: "INSTANCE_LAST_UPDATE", Value: fmt.Sprintf("%d", time.Now().Unix())}
			UpdateTimeEnvExist = true
		}
	}
	if !UpdateTimeEnvExist {
		modDeployment.Spec.Template.Spec.Containers[0].Env = append(
			modDeployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "INSTANCE_LAST_UPDATE", Value: fmt.Sprintf("%d", time.Now().Unix())})
	}
	modJSON, err := json.Marshal(modDeployment)
	if err != nil {
		return []byte{}, err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(curJSON, modJSON, appsv1.Deployment{})
	return patchBytes, err
}
