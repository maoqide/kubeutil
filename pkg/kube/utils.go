package kube

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	ContainerStateWaiting    string = "Waiting"
	ContainerStateRunning    string = "Running"
	ContainerStateTerminated string = "Terminated"
	ContainerStateUnknown    string = "Unknown"
)

// GetContainerState parse corev1.ContainerState to string
func GetContainerState(state corev1.ContainerState) string {
	if state.Waiting != nil {
		return ContainerStateWaiting
	}
	if state.Running != nil {
		return ContainerStateRunning
	}
	if state.Terminated != nil {
		return ContainerStateTerminated
	}
	return ContainerStateUnknown
}
