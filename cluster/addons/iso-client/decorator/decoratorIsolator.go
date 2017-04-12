package decorator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kubernetes/cluster/addons/iso-client/coreaffinity/cputopology"
	"k8s.io/kubernetes/cluster/addons/iso-client/coreaffinity/discovery"
	"k8s.io/kubernetes/cluster/addons/iso-client/opaque"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
)

type decoratorIsolator struct {
	name string
	env  map[string]string
}

// Implement "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (c *decoratorIsolator) Init() error {
	// nothing to do
	return nil
}

// Constructor for Isolator
func New(name string, env map[string]string) (*decoratorIsolator, error) {
	return &decoratorIsolator{
		name: name,
		env:  env,
	}, nil
}

// Implement "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (c *decoratorIsolator) PreStartPod(pod *v1.Pod, resource *lifecycle.CgroupInfo) ([]*lifecycle.IsolationControl, error) {
	controls := []*lifecycle.IsolationControl{
		{
			MapValue: c.env,
			Kind:     lifecycle.IsolationControl_ENVIRONMENT_VARIABLES,
		},
	}
	return controls, nil
}

// Implement "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (c *decoratorIsolator) PostStopPod(cgroupInfo *lifecycle.CgroupInfo) error {
	// nothing to do
	return nil
}

// Implement "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (c *decoratorIsolator) ShutDown() {
	// nothing to do
	return
}

// Implement "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (c *decoratorIsolator) Name() string {
	return c.name
}
