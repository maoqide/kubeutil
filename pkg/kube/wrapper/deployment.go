package wrapper

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/maoqide/kubeutil/pkg/kube"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var deployTemplate = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.Name}}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
      - name: app
        image: {{.Image}}
        ports:
        - containerPort: {{.Port}}
`

// DeploymentWrapper wrapped kubernetes Deployment
type DeploymentWrapper struct {
	deployment *appsv1.Deployment
	err        error
}

// NewDeploymentWrapper create DeploymentWrapper
func NewDeploymentWrapper() *DeploymentWrapper {
	return &DeploymentWrapper{err: errors.New("no created")}
}

// Create create Deployment from template
func (d *DeploymentWrapper) Create(opts *Options) *DeploymentWrapper {
	var deployYaml bytes.Buffer
	tpl := deployTemplate
	dtemplate, err := template.New("deployment").Parse(tpl)
	if err != nil {
		d.err = err
		return nil
	}
	err = dtemplate.Execute(&deployYaml, opts)
	if err != nil {
		d.err = err
		return nil
	}
	deploy, _, err := kube.DecodeKubeObj(deployYaml.Bytes())
	if err != nil {
		d.err = err
		return nil
	}
	d.deployment = deploy.(*appsv1.Deployment)
	d.err = nil
	return d
}

// SetResource set resource limit to deployment
func (d *DeploymentWrapper) SetResource(resources *corev1.ResourceRequirements) *DeploymentWrapper {
	if !d.Vaildate() {
		return d
	}
	d.deployment.Spec.Template.Spec.Containers[0].Resources = *resources
	return d
}

// AddProbe add probes to deployment
func (d *DeploymentWrapper) AddProbe(probe *corev1.Probe, probeType string) *DeploymentWrapper {
	if !d.Vaildate() {
		return d
	}
	switch probeType {
	case "readiness":
		d.deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = probe
	case "liveness":
		d.deployment.Spec.Template.Spec.Containers[0].LivenessProbe = probe
	default:
		d.err = errors.New("invalid probe type")
	}
	return d
}

// AddCommand add command to deployment
func (d *DeploymentWrapper) AddCommand(cmd []string) *DeploymentWrapper {
	if !d.Vaildate() {
		return d
	}
	d.deployment.Spec.Template.Spec.Containers[0].Command = cmd
	return d
}

// AddArgs add args to deployment
func (d *DeploymentWrapper) AddArgs(args []string) *DeploymentWrapper {
	if !d.Vaildate() {
		return d
	}
	d.deployment.Spec.Template.Spec.Containers[0].Args = args
	return d
}

// AddPersistentVolume add persistent volume to deployment
func (d *DeploymentWrapper) AddPersistentVolume(pvcName, mountPath string) *DeploymentWrapper {
	if !d.Vaildate() {
		return d
	}
	if pvcName == "" || mountPath == "" {
		d.err = errors.New("invalid pvcName or mountPath when AddPersistentVolume")
		return d
	}
	volumeName := fmt.Sprintf("%s-volume", pvcName)
	d.deployment.Spec.Template.Spec.Volumes = append(d.deployment.Spec.Template.Spec.Volumes,
		*genVolume(volumeName, pvcName))
	d.deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		d.deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  false,
		})
	return d
}

// Complete complete Deployment config
func (d *DeploymentWrapper) Complete() (*appsv1.Deployment, error) {
	// adding configs here
	return d.deployment, d.err
}

// Vaildate check if err nil
func (d *DeploymentWrapper) Vaildate() bool {
	if d.err != nil {
		return false
	}
	if d.deployment == nil {
		d.err = errors.New("nil deployment")
		return false
	}
	return true
}

// Err return d.err
func (d *DeploymentWrapper) Err() error {
	return d.err
}

func genVolume(volumeName, claimName string) *corev1.Volume {
	return &corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: claimName,
			},
		},
	}
}
