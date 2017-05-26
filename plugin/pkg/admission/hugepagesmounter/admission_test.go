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
	"testing"
)

func TestCalculatePagesCount(t *testing.T) {
	var tests = []struct {
		maxValue      string
		expectedValue int64
		errorExpected bool
	}{
		{
			"200M",
			100,
			false,
		},
		{
			"201M",
			101,
			false,
		},
		{
			"2G",
			1000,
			false,
		},
		{
			"23aM",
			0,
			true,
		},
	}

	for _, test := range tests {
		value, err := calculatePagesCount(test.maxValue)
		fmt.Printf("Value is %#v error is : %v \n", value, err)
		if err != nil && test.errorExpected == false {
			t.Error(err)
		} else if err == nil && test.errorExpected == true {
			t.Error("Error expected for calculatePagesCount but received none")
		}
		if value != test.expectedValue {
			t.Errorf("Expecting value: %v got: %v\n", test.expectedValue, value)
		}
	}
}
