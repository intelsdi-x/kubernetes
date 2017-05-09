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

package cm

import (
	"sync"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/v1"

	"fmt"
	"strings"
)

var CPUManagerSingleton CPUManager
var once sync.Once


type CPUManager interface {
	Start (getNodeAllocatable func() v1.ResourceList, activePods ActivePodsFunc) error

	GetCpusetCpus(podUID types.UID, containerName string, cpuRequest *resource.Quantity) (*string,error)

	AddCpusetCpusForGuaranteedPod(pod *v1.Pod) error

	GetSharedCores() *string
	//getExclusiveCores() (*string, error)
}


type cpuManagerImpl struct {
	sync.Mutex
	exclusiveCoreMap       map[string]*exclusiveCoreContainerInfo
	cachedSharedCores      *string
	activePods             ActivePodsFunc
	getNodeAllocatable     func() v1.ResourceList

}

type exclusiveCoreContainerInfo struct {
	podUID types.UID
	containerName string
}


var _ CPUManager = &cpuManagerImpl{}

func GetCPUManagerSingleton(enabled bool) CPUManager {
	once.Do(func() {
		if enabled {
			CPUManagerSingleton = &cpuManagerImpl{}
		} else {
			CPUManagerSingleton = &cpuManagerNoop{}
		}
	})
	return CPUManagerSingleton
}

func (cmg *cpuManagerImpl) Start(getNodeAllocatable func() v1.ResourceList, activePods ActivePodsFunc) error {
	glog.V(3).Infof("Starting CPU Manager")
	cmg.activePods = activePods
	cmg.getNodeAllocatable = getNodeAllocatable

	cmg.exclusiveCoreMap = make(map[string]*exclusiveCoreContainerInfo)
	coreNum := getNodeAllocatable().Cpu().Value()/1000

	for core:=int64(0); core<coreNum; core++ {
		cmg.exclusiveCoreMap[fmt.Sprintf("%d",core)] = nil
	}

	cmg.cachedSharedCores = cmg.calculateSharedCores()

	if cmg.cachedSharedCores == nil {
		return fmt.Errorf("could not find free shared cores during CPU Manager start")
	}

	return nil

}

// AddCpusetCpusForGuaranteedPod find cores and mark them as exclusive for containers in pod
func (cmg *cpuManagerImpl) AddCpusetCpusForGuaranteedPod(pod *v1.Pod) error {
	return nil
}

// GetCpusetCpus returns list of cpus for the containers in pod
func (cmg *cpuManagerImpl) GetCpusetCpus(podUID types.UID, containerName string, cpuRequest *resource.Quantity) (*string,error) {
	return nil, nil
}

// GetSharedCores return shared cores from cached value or recalculates
func (cmg *cpuManagerImpl) GetSharedCores() *string {
	return nil
}


func (cmg *cpuManagerImpl) calculateSharedCores() *string {
	sharedCores := []string{}
	for coreID, coreInfo := range cmg.exclusiveCoreMap {
		if coreInfo == nil {
			sharedCores = append(sharedCores, coreID)
		}
	}
	if len(sharedCores) == 0 {
		return nil
	}
	return &strings.Join(sharedCores, ",")
}




type cpuManagerNoop struct {
}

var _ CPUManager = &cpuManagerNoop{}

func (cmg *cpuManagerNoop ) Start() {
	return
}

func (cmg *cpuManagerNoop ) GetCpusetCpus(podUID types.UID, containerName string, cpuRequest *resource.Quantity) (*string,error) {
	return nil, nil
}
