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
	cadvisorapi "github.com/google/cadvisor/info/v1"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
)

type cpuDriver struct {
	cadvisor cadvisor.Interface
	topology []cadvisorapi.Node
}

func NewCPUDriver(cadvisor cadvisor.Interface) *cpuDriver {
	var info *cadvisorapi.MachineInfo
	var err error
	cpud := &cpuDriver{
		cadvisor: cadvisor,
	}

	info, err = cadvisor.MachineInfo()
	if err != nil {
		return nil
	}
	cpud.topology = info.Topology
	return cpud
}
