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
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/api/v1"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
)

type cpuStaticManager struct {
	sync.Mutex
	containerRuntime kubecontainer.Runtime
	driver           *cpuDriver
}

func (csm *cpuStaticManager) AddPod(pod *v1.Pod, qosClass v1.PodQOSClass) error {
	csm.Lock()
	defer csm.Unlock()
	//TODO(SCc) Add pod to list
	// check if new pod or already on the list - if new then :
	// -> if pod needs exclusive core find free and start eviction
	// -> if not mark as using shared pool
	// if not new, already on the list exit
	return nil
}

func (csm *cpuStaticManager) DeletePod(pod *v1.Pod) error {
	csm.Lock()
	defer csm.Unlock()
	return nil
}

func (csm *cpuStaticManager) GetQoSClassCpuset(qosClass v1.PodQOSClass) cpuList {
	//TODO (SSc) Return string of available CPUs for QoS class
	return ""
}

func (csm *cpuStaticManager) GetPodCpuset(podUID types.UID) cpuList {
	return ""
}

func (csm *cpuStaticManager) GetContainerCpuSet(podUID types.UID, containerName string) cpuList {
	return ""
}
