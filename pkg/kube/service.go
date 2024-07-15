package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// ServiceBox provide functions for kubernetes service.
type ServiceBox struct {
	clientset clientset.Interface
}

// Get get specified service in specified namespace.
func (s *ServiceBox) Get(ctx context.Context, name, namespace string) (*corev1.Service, error) {
	opt := metav1.GetOptions{}
	return s.clientset.CoreV1().Services(namespace).Get(ctx, name, opt)
}

// List list services in specified namespace.
func (s *ServiceBox) List(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	opt := metav1.ListOptions{}
	return s.clientset.CoreV1().Services(namespace).List(ctx, opt)
}

// Exists check if service exists.
func (s *ServiceBox) Exists(ctx context.Context, name, namespace string) (bool, error) {
	_, err := s.Get(ctx, name, namespace)
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// Create creates a service
func (s *ServiceBox) Create(ctx context.Context, service *corev1.Service, namespace string) (*corev1.Service, error) {
	return s.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
}

// Delete delete service
func (s *ServiceBox) Delete(ctx context.Context, name, namespace string) error {
	opt := commonDeleteOpt
	return s.clientset.CoreV1().Services(namespace).Delete(ctx, name, opt)
}
