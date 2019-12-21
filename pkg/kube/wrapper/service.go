package wrapper

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/maoqide/kubeutil/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

var serviceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  ports:
  - name: app
    port: {{.Port}}
    protocol: TCP
    targetPort: {{.Port}}
  selector:
    app : {{.Name}}
`

// ServiceWrapper wrapped kubernetes Service
type ServiceWrapper struct {
	service *corev1.Service
	err     error
}

// NewServiceWrapper create ServiceWrapper
func NewServiceWrapper() *ServiceWrapper {
	return &ServiceWrapper{err: errors.New("no created")}
}

// Create create Service from template
func (d *ServiceWrapper) Create(opts *Options) *ServiceWrapper {
	var yaml bytes.Buffer
	dtemplate, err := template.New("Service").Parse(serviceTemplate)
	if err != nil {
		d.err = err
		return nil
	}
	err = dtemplate.Execute(&yaml, opts)
	if err != nil {
		d.err = err
		return nil
	}
	svc, _, err := kube.DecodeKubeObj(yaml.Bytes())
	if err != nil {
		d.err = err
		return nil
	}
	d.service = svc.(*corev1.Service)
	d.err = nil
	return d
}

// Complete complete Service config
func (d *ServiceWrapper) Complete() (*corev1.Service, error) {
	// adding configs here
	return d.service, d.err
}

// Vaildate check if err nil
func (d *ServiceWrapper) Vaildate() bool {
	if d.err != nil {
		return false
	}
	if d.service == nil {
		d.err = errors.New("nil Service")
		return false
	}
	return true
}

// Err return d.err
func (d *ServiceWrapper) Err() error {
	return d.err
}
