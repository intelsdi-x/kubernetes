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

package cpumanager

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/api/v1"
)

type cpuNoopManager struct {
	driver cpuDriver
}

func (cnm *cpuNoopManager) AddPod(pod *v1.Pod, qosClass v1.PodQOSClass) error {
	return nil
}

func (cnm *cpuNoopManager) DeletePod(pod *v1.Pod) error {
	return nil
}

func (cnm *cpuNoopManager) GetQoSClassCpuset(qosClass v1.PodQOSClass) cpuList {
	return ""
}

func (cnm *cpuNoopManager) GetPodCpuset(podUID types.UID) cpuList {
	return ""
}

func (cnm *cpuNoopManager) GetContainerCpuSet(podUID types.UID, containerName string) cpuList {
	return ""
}
