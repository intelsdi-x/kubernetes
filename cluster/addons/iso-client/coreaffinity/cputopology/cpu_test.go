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

package cputopology

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"reflect"
)

var TOPOLOGY CPUTopology = CPUTopology{
	CPU: []CPU{
		{SocketID: 0, CoreID: 0, CPUID: 0},
		{SocketID: 0, CoreID: 1, CPUID: 1},
		{SocketID: 0, CoreID: 0, CPUID: 2},
		{SocketID: 0, CoreID: 1, CPUID: 3},
		{SocketID: 1, CoreID: 0, CPUID: 4},
		{SocketID: 1, CoreID: 1, CPUID: 5},
		{SocketID: 1, CoreID: 0, CPUID: 6},
		{SocketID: 1, CoreID: 1, CPUID: 7},
	},
}

func TestCPUTopology_GetNumSockets(t *testing.T) {
	testCases := []struct {
		topology               CPUTopology
		requestedSocketsNumber int
	}{
		{
			topology:               TOPOLOGY,
			requestedSocketsNumber: 2,
		},
		{
			topology: CPUTopology{
				CPU: []CPU{},
			},
			requestedSocketsNumber: 0,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.topology.GetNumSockets(), tc.requestedSocketsNumber)
	}
}

func TestCPUTopology_GetSocket(t *testing.T) {
	testCases := []struct {
		requestedCPUs   []CPU
		requestedSocket int
	}{
		{
			requestedCPUs: []CPU{
				{SocketID: 0, CoreID: 0, CPUID: 0},
				{SocketID: 0, CoreID: 1, CPUID: 1},
				{SocketID: 0, CoreID: 0, CPUID: 2},
				{SocketID: 0, CoreID: 1, CPUID: 3},
			},
			requestedSocket: 0,
		},
		{
			requestedCPUs: []CPU{
				{SocketID: 1, CoreID: 0, CPUID: 4},
				{SocketID: 1, CoreID: 1, CPUID: 5},
				{SocketID: 1, CoreID: 0, CPUID: 6},
				{SocketID: 1, CoreID: 1, CPUID: 7},
			},
			requestedSocket: 1,
		},
		{
			requestedCPUs:   nil,
			requestedSocket: 3,
		},
	}

	for _, tc := range testCases {
		cpus := TOPOLOGY.GetSocket(tc.requestedSocket)
		assert.Equal(t, cpus, tc.requestedCPUs)
	}
}

func TestCPUTopology_GetCore(t *testing.T) {
	testCases := []struct {
		requestedCPUs   []CPU
		requestedCore   int
		requestedSocket int
	}{
		{
			requestedCPUs: []CPU{
				{SocketID: 0, CoreID: 1, CPUID: 1},
				{SocketID: 0, CoreID: 1, CPUID: 3},
			},
			requestedSocket: 0,
			requestedCore:   1,
		},
		{
			requestedCPUs: []CPU{
				{SocketID: 1, CoreID: 0, CPUID: 4},
				{SocketID: 1, CoreID: 0, CPUID: 6},
			},
			requestedSocket: 1,
			requestedCore:   0,
		},
		{
			requestedCPUs:   nil,
			requestedCore:   0,
			requestedSocket: 2,
		},
	}

	for _, tc := range testCases {
		cpus := TOPOLOGY.GetCore(tc.requestedSocket, tc.requestedCore)
		assert.Equal(t, cpus, tc.requestedCPUs)
	}
}

func TestCPUTopology_GetCPU(t *testing.T) {
	testCases := []struct {
		requestedCPUs *CPU
		requestedCPU  int
	}{
		{
			requestedCPUs: &CPU{SocketID: 0, CoreID: 1, CPUID: 1},
			requestedCPU:  1,
		},
		{
			requestedCPUs: &CPU{SocketID: 1, CoreID: 0, CPUID: 6},
			requestedCPU:  6,
		},
		{
			requestedCPUs: nil,
			requestedCPU:  10,
		},
	}

	for _, tc := range testCases {
		cpus := TOPOLOGY.GetCPU(tc.requestedCPU)
		assert.Equal(t, cpus, tc.requestedCPUs)
	}
}

func TestCPUTopology_Reserve(t *testing.T) {
	topology := CPUTopology{
		CPU: []CPU{
			{SocketID: 0, CPUID: 0, CoreID: 0},
		},
	}
	err := topology.Reserve(0)
	assert.Nil(t, err)

	cpu := topology.GetCPU(0)
	assert.True(t, cpu.IsInUse)
	err = topology.Reserve(0)
	assert.NotNil(t, err)
}

func TestCPUTopology_Reclaim(t *testing.T) {
	topology := CPUTopology{
		CPU: []CPU{
			{SocketID: 0, CPUID: 0, CoreID: 0, IsInUse: true},
		},
	}
	cpu := topology.GetCPU(0)
	assert.True(t, cpu.IsInUse)

	topology.Reclaim(0)
	assert.False(t, cpu.IsInUse)
}

func TestCPUTopology_GetTotalCPUs(t *testing.T) {
	testCases := []struct {
		cpuTopology   *CPUTopology
		expectedTotal int
	}{
		{
			cpuTopology:   &CPUTopology{},
			expectedTotal: 0,
		},
		{
			cpuTopology:   &TOPOLOGY,
			expectedTotal: 8,
		},
	}
	for _, testCase := range testCases {
		if testCase.cpuTopology.GetTotalCPUs() != testCase.expectedTotal {
			t.Errorf("Expected number of CPUs (%q) isn't expected number (%q)",
				testCase.cpuTopology.GetTotalCPUs(), testCase.expectedTotal)
		}
	}

}

func TestCPUTopology_GetAvailableCPUs(t *testing.T) {
	testCases := []struct {
		cpuTopology   *CPUTopology
		availableCPUs []CPU
	}{
		{
			cpuTopology: &CPUTopology{
				CPU: []CPU{
					{SocketID: 0, CoreID: 0, CPUID: 0},
					{SocketID: 0, CoreID: 1, CPUID: 1, IsInUse: true},
					{SocketID: 0, CoreID: 0, CPUID: 2},
					{SocketID: 0, CoreID: 1, CPUID: 3, IsInUse: true},
					{SocketID: 1, CoreID: 0, CPUID: 4},
					{SocketID: 1, CoreID: 1, CPUID: 5, IsInUse: true},
					{SocketID: 1, CoreID: 0, CPUID: 6},
					{SocketID: 1, CoreID: 1, CPUID: 7, IsInUse: true},
				},
			},
			availableCPUs: []CPU{
				{SocketID: 0, CoreID: 0, CPUID: 0},
				{SocketID: 0, CoreID: 0, CPUID: 2},
				{SocketID: 1, CoreID: 0, CPUID: 4},
				{SocketID: 1, CoreID: 0, CPUID: 6},
			},
		},
		{
			cpuTopology: &CPUTopology{
				CPU: []CPU{
					{SocketID: 0, CoreID: 1, CPUID: 1, IsInUse: true},
					{SocketID: 0, CoreID: 1, CPUID: 3, IsInUse: true},
					{SocketID: 1, CoreID: 1, CPUID: 5, IsInUse: true},
					{SocketID: 1, CoreID: 1, CPUID: 7, IsInUse: true},
				},
			},
			availableCPUs: []CPU{},
		},
		{
			cpuTopology: &CPUTopology{
				CPU: []CPU{},
			},
			availableCPUs: []CPU{},
		},
	}

	for _, testCase := range testCases {
		if !reflect.DeepEqual(testCase.cpuTopology.GetAvailableCPUs(), testCase.availableCPUs) {
			t.Error("Requested number of CPUs doesn't match expected one")
		}
	}
}
