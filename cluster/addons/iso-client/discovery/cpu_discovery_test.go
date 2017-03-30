// +build linux

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

package discovery

import (
	"testing"

	"fmt"
	"github.com/stretchr/testify/assert"
)

func TestLscpuExecution(t *testing.T) {
	data, err := lscpuExecution()
	assert.Nil(t, err)
	assert.Contains(t, data, "0,0,0,0,,0,0,0,0")

}

func TestIsolcpusParsing(t *testing.T) {
	testCases := []struct {
		cmdline      string
		expectedCPUs []int
		isError      bool
	}{
		{
			cmdline:      `initrd=\initramfs-linux.img root=/dev/disk1 quiet rw idle=nomwait`,
			expectedCPUs: nil,
			isError:      false,
		},
		{
			cmdline:      `initrd=\initramfs-linux.img root=/dev/disk1 quiet rw idle=nomwait isolcpus=1,3,5`,
			expectedCPUs: []int{1, 3, 5},
			isError:      false,
		},
		{
			cmdline:      `initrd=\initramfs-linux.img root=/dev/disk1 quiet rw idle=nomwait isolcpus=1,x`,
			expectedCPUs: nil,
			isError:      true,
		},
	}

	for _, tc := range testCases {
		cpus, err := retriveIsolcpus(tc.cmdline)
		assert.Equal(t, cpus, tc.expectedCPUs)
		if tc.isError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}

}

func TestLscpuParse(t *testing.T) {
	lscpuRawOutput := `# The following is the parsable format, which can be fed to other
# programs. Each different item in every column has an unique ID
# starting from zero.
# CPU,Core,Socket,Node,,L1d,L1i,L2,L3
0,0,0,0,,0,0,0,0
1,1,0,0,,1,1,1,0
2,2,0,0,,2,2,2,0
3,3,0,0,,3,3,3,0
4,0,0,0,,0,0,0,0
5,1,0,0,,1,1,1,0
6,2,0,0,,2,2,2,0
7,3,0,0,,3,3,3,0
`

	testCases := []struct {
		isolatedCPUs       []int
		isError            bool
		cpuCounter         int
		additionalTestLine string
	}{
		{
			isolatedCPUs:       []int{},
			isError:            true,
			cpuCounter:         0,
			additionalTestLine: "x,y;z",
		},
		{
			isolatedCPUs:       []int{0},
			isError:            false,
			cpuCounter:         8,
			additionalTestLine: "",
		},
		{
			isolatedCPUs:       []int{},
			isError:            false,
			cpuCounter:         8,
			additionalTestLine: "",
		},
	}

	for _, tc := range testCases {
		lscpuOutput := fmt.Sprintf("%s%s", lscpuRawOutput, tc.additionalTestLine)
		topo, err := cpuParse(lscpuOutput, tc.isolatedCPUs)
		if tc.isError {
			assert.NotNil(t, err)
			assert.Nil(t, topo)
			continue
		}
		assert.Nil(t, err)
		assert.NotNil(t, topo)
		assert.Equal(t, len(topo.CPU), tc.cpuCounter)

		for _, cpuid := range tc.isolatedCPUs {
			cpu := topo.GetCPU(cpuid)
			assert.True(t, cpu.IsIsolated)
		}

	}
}

func TestIsolcpusRetriving(t *testing.T) {
	_, err := isolcpus()
	assert.Nil(t, err)
}

func TestDiscoverTopology(t *testing.T) {
	topo, err := DiscoverTopology()
	assert.Nil(t, err)
	assert.True(t, topo.GetNumSockets() > 0)
	assert.NotNil(t, topo.GetCPU(0))
}
