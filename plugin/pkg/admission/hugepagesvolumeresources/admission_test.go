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

package hugepagesvolumeresources

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/client-go/1.5/pkg/auth/user"
	"k8s.io/kubernetes/pkg/api"
)

func TestCalculatePagesCount(t *testing.T) {
	var tests = []struct {
		maxValue      string
		expectedValue int64
		shouldFail    bool
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
		if err != nil && test.shouldFail == false {
			t.Error(err)
		} else if err == nil && test.shouldFail == true {
			t.Error("Error expected for calculatePagesCount but received none")
		}
		if value != test.expectedValue {
			t.Errorf("Expecting value: %v got: %v\n", test.expectedValue, value)
		}
	}
}

func TestFillResourcesFor(t *testing.T) {

	volumeName := "test"

	pod := &api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mypod",
			Namespace: "namespace",
			Labels: map[string]string{
				"security": "S2",
			},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Name: "test",
					VolumeMounts: []api.VolumeMount{
						{
							Name:      volumeName,
							MountPath: "/mnt/huge",
						},
					},
				},
			},
			Volumes: []api.Volume{
				{
					Name: volumeName,
					VolumeSource: api.VolumeSource{
						HugePages: &api.HugePagesVolumeSource{
							MaxSize:  "100M",
							PageSize: "2M",
						},
					},
				},
			},
		},
	}

	fillResourcesFor(pod.Spec.Containers, "test", 50)

	hugePages := pod.Spec.Containers[0].Resources.Requests[hugepagesResource]

	if hugePages.Value() != 50 {
		t.Errorf("Request reseources for %s invalid should be 50 is %v", hugepagesResource, hugePages.Value())
	}

}

func TestAdmitPod(t *testing.T) {

	volumeName := "test"
	plugin := &hugePagesMounterPlugin{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}

	var tests = map[string]struct {
		pod        *api.Pod
		shouldFail bool
	}{
		"good one": {
			pod: &api.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mypod",
					Namespace: "namespace",
					Labels: map[string]string{
						"security": "S2",
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name: "test",
							VolumeMounts: []api.VolumeMount{
								{
									Name:      volumeName,
									MountPath: "/mnt/huge",
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: volumeName,
							VolumeSource: api.VolumeSource{
								HugePages: &api.HugePagesVolumeSource{
									MaxSize:  "100M",
									PageSize: "2M",
								},
							},
						},
					},
				},
			},
			shouldFail: false,
		},
		"incorrect quantity": {
			pod: &api.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mypod",
					Namespace: "namespace",
					Labels: map[string]string{
						"security": "S2",
					},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name: "test",
							VolumeMounts: []api.VolumeMount{
								{
									Name:      volumeName,
									MountPath: "/mnt/huge",
								},
							},
						},
					},
					Volumes: []api.Volume{
						{
							Name: volumeName,
							VolumeSource: api.VolumeSource{
								HugePages: &api.HugePagesVolumeSource{
									MaxSize:  "100shouldFailM",
									PageSize: "2M",
								},
							},
						},
					},
				},
			},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		err := admitPod(plugin, test.pod, volumeName)
		if err != nil && test.shouldFail == false {
			t.Error(err)
		} else if err == nil && test.shouldFail == true {
			t.Error("Error expected for admit pod but received none")
		}

	}
}

func admitPod(plugin *hugePagesMounterPlugin, pod *api.Pod, volumeName string) error {
	attrs := admission.NewAttributesRecord(
		pod,
		nil,
		api.Kind("Pod").WithVersion("version"),
		"namespace",
		"",
		api.Resource("pods").WithVersion("version"),
		"",
		admission.Create,
		&user.DefaultInfo{},
	)

	err := plugin.Admit(attrs)
	if err != nil {
		return err
	}

	value, found := pod.ObjectMeta.Annotations[fmt.Sprintf("%s/%s", annotationPrefix, volumeName)]
	if !found {
		return fmt.Errorf("Annotations hasn't been assigned")
	}

	if value != hugepagesResource {
		return fmt.Errorf("Annotation should be %s is %s instead", hugepagesResource, value)
	}
	return nil

}
