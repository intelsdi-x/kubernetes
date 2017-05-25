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
	"strconv"
	"strings"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/kubernetes/pkg/api"
	kubeapiserveradmission "k8s.io/kubernetes/pkg/kubeapiserver/admission"
)

const (
	annotationPrefix = "hugepagesmounter.admission.kubernetes.io"
	pluginName       = "HugePagesMounter"
	volumeNamePrefix = "hugepagesmounter-"
)

func init() {
	kubeapiserveradmission.Plugins.Register(pluginName, func(config io.Reader) (admission.Interface, error) {
		return NewPlugin(), nil
	})
}

// hugePagesMounterPlugin is an implementation of admission.Interface.
type hugePagesMounterPlugin struct {
	*admission.Handler
	//client internalclientset.Interface
}

// NewPlugin creates a new hugePagesMounter admission plugin.
func NewPlugin() *hugePagesMounterPlugin {
	return &hugePagesMounterPlugin{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
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

	for _, volume := range pod.Spec.Volumes {
		if volume.HugePages != nil {
			hugepagesCount, err := calculatePagesCount(volume.HugePages.MaxSize)
			if err != nil {
				return fmt.Errorf("Cannot calculate hugePages size for %v : %v", volume.Name, err)
			}
			modifyContainers(hugepagesCount, pod.Spec.Containers, volume.Name)

		}
	}

	return nil
}

func modifyContainers(hugepagesCount int64, containers []api.Container, volumeName string) {
	for containerID := range containers {
		if containerHasHugepagesVolume(volumeName, containers[containerID]) {
			requests := containers[containerID].Resources.Requests
			if requests == nil {
				requests = make(api.ResourceList)
			}
			hugePagesResourceName := api.ResourceName("alpha.kubernetes.io/hugepages-2048kB")
			hugePage, found := requests[hugePagesResourceName]
			newValue := hugepagesCount
			if found {
				newValue += hugePage.Value()
			}

			requests[hugePagesResourceName] = *resource.NewQuantity(newValue, resource.DecimalSI)
			// make sure we store new request in case it was empty

			glog.Infof("Request is : %#v\n", requests)
			containers[containerID].Resources.Requests = requests
		}
	}
}

func containerHasHugepagesVolume(volumeName string, container api.Container) bool {
	for _, containerVolume := range container.VolumeMounts {
		if containerVolume.Name == volumeName {
			return true
		}
	}
	return false
}

func calculatePagesCount(maxSize string) (int64, error) {
	// TODO(pprokop): add support for other sizes
	glog.Infof("-----------Last char: %s\n", string(maxSize[len(maxSize)-1]))
	if !strings.Contains("mM", string(maxSize[len(maxSize)-1])) {
		return 0, fmt.Errorf("Not supported size of HugePages not: %v", maxSize[len(maxSize)-1])
	}
	glog.Infof("-------------------------------maxSize: %s", maxSize)

	size, err := strconv.Atoi(maxSize[:len(maxSize)-1])

	if err != nil {
		return int64(size), fmt.Errorf("Only available sizes are m and M not: %v", maxSize)
	}

	glog.Infof("SIze is : %v\n", size)

	size1 := convertMaxSizeTo2MHugePages(size)
	glog.Infof("SIze1 is : %v\n", size1)
	return size1, nil

}

func convertMaxSizeTo2MHugePages(maxSize int) int64 {
	if maxSize%2 != 0 {
		return int64((maxSize + 1) / 2)
	}
	return int64(maxSize / 2)
}
