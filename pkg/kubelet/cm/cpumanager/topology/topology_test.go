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

package topology

import (
	"reflect"
	"testing"

	cadvisorapi "github.com/google/cadvisor/info/v1"
)

func Test_Discover(t *testing.T) {

	tests := []struct {
		name    string
		args    *cadvisorapi.MachineInfo
		want    *CPUTopology
		wantErr bool
	}{
		{
			name: "FailNumCores",
			args: &cadvisorapi.MachineInfo{
				NumCores: 0,
			},
			want:    &CPUTopology{},
			wantErr: true,
		},
		{
			name: "OneSocketHT",
			args: &cadvisorapi.MachineInfo{
				NumCores: 8,
				Topology: []cadvisorapi.Node{
					{Id: 0,
						Cores: []cadvisorapi.Core{
							{Id: 0, Threads: []int{0, 4}},
							{Id: 1, Threads: []int{1, 5}},
							{Id: 2, Threads: []int{2, 6}},
							{Id: 3, Threads: []int{3, 7}},
						},
					},
				},
			},
			want: &CPUTopology{
				NumCPUs:    8,
				NumSockets: 1,
				NumCores:   4,
				CPUDetails: map[int]CPUInfo{
					0: {CoreID: 0, SocketID: 0},
					1: {CoreID: 1, SocketID: 0},
					2: {CoreID: 2, SocketID: 0},
					3: {CoreID: 3, SocketID: 0},
					4: {CoreID: 0, SocketID: 0},
					5: {CoreID: 1, SocketID: 0},
					6: {CoreID: 2, SocketID: 0},
					7: {CoreID: 3, SocketID: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "DualSocketNoHT",
			args: &cadvisorapi.MachineInfo{
				NumCores: 4,
				Topology: []cadvisorapi.Node{
					{Id: 0,
						Cores: []cadvisorapi.Core{
							{Id: 0, Threads: []int{0}},
							{Id: 1, Threads: []int{1}},
						},
					},
					{Id: 1,
						Cores: []cadvisorapi.Core{
							{Id: 0, Threads: []int{2}},
							{Id: 1, Threads: []int{3}},
						},
					},
				},
			},
			want: &CPUTopology{
				NumCPUs:    4,
				NumSockets: 2,
				NumCores:   4,
				CPUDetails: map[int]CPUInfo{
					0: {CoreID: 0, SocketID: 0},
					1: {CoreID: 1, SocketID: 0},
					2: {CoreID: 0, SocketID: 1},
					3: {CoreID: 1, SocketID: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "DualSocketHT",
			args: &cadvisorapi.MachineInfo{
				NumCores: 48,
				Topology: []cadvisorapi.Node{
					{Id: 0,
						Cores: []cadvisorapi.Core{
							{Id: 0, Threads: []int{0,24}},
							{Id: 1, Threads: []int{1,25}},
							{Id: 2, Threads: []int{2,26}},
							{Id: 3, Threads: []int{3,27}},
							{Id: 4, Threads: []int{4,28}},
							{Id: 5, Threads: []int{5,29}},
							{Id: 8, Threads: []int{6,30}},
							{Id: 9, Threads: []int{7,31}},
							{Id: 10, Threads: []int{8,32}},
							{Id: 11, Threads: []int{9,33}},
							{Id: 12, Threads: []int{10,34}},
							{Id: 13, Threads: []int{11,35}},
						},
					},
					{Id: 1,
						Cores: []cadvisorapi.Core{
							{Id: 0, Threads: []int{12,36}},
							{Id: 1, Threads: []int{13,37}},
							{Id: 2, Threads: []int{14,38}},
							{Id: 3, Threads: []int{15,39}},
							{Id: 4, Threads: []int{16,40}},
							{Id: 5, Threads: []int{17,41}},
							{Id: 8, Threads: []int{18,42}},
							{Id: 9, Threads: []int{19,43}},
							{Id: 10, Threads: []int{20,44}},
							{Id: 11, Threads: []int{21,45}},
							{Id: 12, Threads: []int{22,46}},
							{Id: 13, Threads: []int{23,47}},
						},
					},
				},
			},
			want: &CPUTopology{
				NumCPUs:    48,
				NumSockets: 2,
				NumCores:   24,
				CPUDetails: map[int]CPUInfo{
					0: {CoreID: 0, SocketID: 0},
					1: {CoreID: 1, SocketID: 0},
					2: {CoreID: 2, SocketID: 0},
					3: {CoreID: 3, SocketID: 0},
					4: {CoreID: 4, SocketID: 0},
					5: {CoreID: 5, SocketID: 0},
					6: {CoreID: 8, SocketID: 0},
					7: {CoreID: 9, SocketID: 0},
					8: {CoreID: 10, SocketID: 0},
					9: {CoreID: 11, SocketID: 0},
					10: {CoreID: 12, SocketID: 0},
					11: {CoreID: 13, SocketID: 0},
					12: {CoreID: 0, SocketID: 1},
					13: {CoreID: 1, SocketID: 1},
					14: {CoreID: 2, SocketID: 1},
					15: {CoreID: 3, SocketID: 1},
					16: {CoreID: 4, SocketID: 1},
					17: {CoreID: 5, SocketID: 1},
					18: {CoreID: 8, SocketID: 1},
					19: {CoreID: 9, SocketID: 1},
					20: {CoreID: 10, SocketID: 1},
					21: {CoreID: 11, SocketID: 1},
					22: {CoreID: 12, SocketID: 1},
					23: {CoreID: 13, SocketID: 1},
					24: {CoreID: 0, SocketID: 0},
					25: {CoreID: 1, SocketID: 0},
					26: {CoreID: 2, SocketID: 0},
					27: {CoreID: 3, SocketID: 0},
					28: {CoreID: 4, SocketID: 0},
					29: {CoreID: 5, SocketID: 0},
					30: {CoreID: 8, SocketID: 0},
					31: {CoreID: 9, SocketID: 0},
					32: {CoreID: 10, SocketID: 0},
					33: {CoreID: 11, SocketID: 0},
					34: {CoreID: 12, SocketID: 0},
					35: {CoreID: 13, SocketID: 0},
					36: {CoreID: 0, SocketID: 1},
					37: {CoreID: 1, SocketID: 1},
					38: {CoreID: 2, SocketID: 1},
					39: {CoreID: 3, SocketID: 1},
					40: {CoreID: 4, SocketID: 1},
					41: {CoreID: 5, SocketID: 1},
					42: {CoreID: 8, SocketID: 1},
					43: {CoreID: 9, SocketID: 1},
					44: {CoreID: 10, SocketID: 1},
					45: {CoreID: 11, SocketID: 1},
					46: {CoreID: 12, SocketID: 1},
					47: {CoreID: 13, SocketID: 1},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Discover(tt.args)
			if err != nil {
				if tt.wantErr {
					t.Logf("Discover() expected error = %v", err)
				} else {
					t.Errorf("Discover() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discover() = %v, want %v", got, tt.want)
			}
		})
	}
}