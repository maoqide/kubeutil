package kube

import (
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
func (s *ServiceBox) Get(name, namespace string) (*corev1.Service, error) {
	opt := metav1.GetOptions{}
	return s.clientset.CoreV1().Services(namespace).Get(name, opt)
}

// List list services in specified namespace.
func (s *ServiceBox) List(namespace string) (*corev1.ServiceList, error) {
	opt := metav1.ListOptions{}
	return s.clientset.CoreV1().Services(namespace).List(opt)
}

// Exists check if service exists.
func (s *ServiceBox) Exists(name, namespace string) (bool, error) {
	_, err := s.Get(name, namespace)
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

// Create creates a service
func (s *ServiceBox) Create(service *corev1.Service, namespace string) (*corev1.Service, error) {
	return s.clientset.CoreV1().Services(namespace).Create(service)
}

// Delete delete service
func (s *ServiceBox) Delete(name, namespace string) error {
	opt := commonDeleteOpt
	return s.clientset.CoreV1().Services(namespace).Delete(name, &opt)
}
