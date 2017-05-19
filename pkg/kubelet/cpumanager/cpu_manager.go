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
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
)

type cpuManagerName string
type cpuList string

var (
	cpuManagerSingleton CPUManager
	oncePrepare         sync.Once
	cpuManagerType      cpuManagerName = CPUManagerNoop
)

const (
	CPUManagerNoop    cpuManagerName = "noop"
	CPUManagerStatic  cpuManagerName = "static"
	CPUManagerDynamic cpuManagerName = "dynamic"
)

type CPUManager interface {
	AddPod(pod *v1.Pod, qosClass v1.PodQOSClass) error
	DeletePod(pod *v1.Pod) error

	GetQoSClassCpuset(qosClass v1.PodQOSClass) cpuList
	GetPodCpuset(podUID types.UID) cpuList
	GetContainerCpuSet(podUID types.UID, containerName string) cpuList
}

func NewCPUManager(name cpuManagerName, runtime kubecontainer.Runtime, cadvisor cadvisor.Interface) (CPUManager, error) {
	oncePrepare.Do(func() {
		cpuManagerType = name
		switch cpuManagerType {
		case CPUManagerNoop:
			cpuManagerSingleton = &cpuNoopManager{}
		case CPUManagerStatic:
			cpuManagerSingleton = &cpuStaticManager{
				containerRuntime: runtime,
				driver:           NewCPUDriver(cadvisor),
			}
		}
	})
	return cpuManagerSingleton, nil
}

func GetCPUManagerSingleton() CPUManager {
	return cpuManagerSingleton
}

func (c cpuList) String() string {
	return string(c)
}
