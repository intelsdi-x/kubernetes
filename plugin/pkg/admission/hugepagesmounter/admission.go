/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hugepagesmounter

import (
	"fmt"
	"io"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	informers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	kubeapiserveradmission "k8s.io/kubernetes/pkg/kubeapiserver/admission"
)

const (
	annotationPrefix = "hugepagesmounter.admission.kubernetes.io"
	pluginName       = "HugePagesMounter"
)

func init() {
	kubeapiserveradmission.Plugins.Register(pluginName, func(config io.Reader) (admission.Interface, error) {
		return NewPlugin(), nil
	})
}

// hugePagesMounterPlugin is an implementation of admission.Interface.
type hugePagesMounterPlugin struct {
	*admission.Handler
	client internalclientset.Interface
}

var _ = kubeapiserveradmission.WantsInternalKubeInformerFactory(&hugePagesMounterPlugin{})
var _ = kubeapiserveradmission.WantsInternalKubeClientSet(&hugePagesMounterPlugin{})

// NewPlugin creates a new pod preset admission plugin.
func NewPlugin() *hugePagesMounterPlugin {
	return &hugePagesMounterPlugin{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

func (plugin *hugePagesMounterPlugin) Validate() error {
	if plugin.client == nil {
		return fmt.Errorf("%s requires a client", pluginName)
	}
	return nil
}

func (a *hugePagesMounterPlugin) SetInternalKubeClientSet(client internalclientset.Interface) {
	a.client = client
}

func (a *hugePagesMounterPlugin) SetInternalKubeInformerFactory(f informers.SharedInformerFactory) {
}

// Admit injects a pod with the specific fields for each pod preset it matches.
func (c *hugePagesMounterPlugin) Admit(a admission.Attributes) error {
	// Ignore all calls to subresources or resources other than pods.
	// Ignore all operations other than CREATE.
	if len(a.GetSubresource()) != 0 || a.GetResource().GroupResource() != api.Resource("pods") || a.GetOperation() != admission.Create {
		return nil
	}

	pod, ok := a.GetObject().(*api.Pod)
	if !ok {
		return errors.NewBadRequest("Resource was marked with kind Pod but was unable to be converted")
	}

	for index, container := range pod.Spec.Containers {
		if hugePage, found := container.Resources.Requests[api.ResourceName("alpha.kubernetes.io/hugepages-2048kB")]; found {
			pod.Spec.Volumes = append(pod.Spec.Volumes, api.Volume{
				Name: "hugepagesmounterplugin-" + fmt.Sprintf("%d", index),
				VolumeSource: api.VolumeSource{
					HugePages: &api.HugePagesVolumeSource{
						PageSize: "2M",
						MaxSize:  fmt.Sprintf("%dM", hugePage.Value()*2),
						MinSize:  "2M",
					},
				},
			},
			)

			pod.Spec.Containers[index].VolumeMounts = append(pod.Spec.Containers[index].VolumeMounts, api.VolumeMount{
				Name:      "hugepagesmounterplugin-" + fmt.Sprintf("%d", index),
				MountPath: "/hugepages",
			})
			glog.V(4).Info("Added hugePagesMount according to requests")
		}

	}

	// add annotation
	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = map[string]string{}
	}
	pod.ObjectMeta.Annotations[fmt.Sprintf("%s", annotationPrefix)] = "hugePageMounter"

	return nil
}
