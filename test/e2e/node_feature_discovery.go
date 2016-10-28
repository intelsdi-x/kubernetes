/*
Copyright 2016 The Kubernetes Authors.

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

package e2e

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/system"
	"k8s.io/kubernetes/pkg/util/uuid"
	"k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = framework.KubeDescribe("Node Feature Discovery [Feature:NodeFeatureDiscovery]", func() {
	const (
		labelPrefix       = "node.alpha.intel.com/nfd"
		imageTag          = "e2e-tests"
		fakeFeatureSource = "fake"
	)
	f := framework.NewDefaultFramework("node-feature-discovery")
	var node *api.Node
	var ns, name, image string
	fakeFeatureLabels := map[string]string{
		fmt.Sprintf("%s-%s-fakefeature1", labelPrefix, fakeFeatureSource): "true",
		fmt.Sprintf("%s-%s-fakefeature2", labelPrefix, fakeFeatureSource): "true",
		fmt.Sprintf("%s-%s-fakefeature3", labelPrefix, fakeFeatureSource): "true",
	}

	BeforeEach(func() {
		ns = f.Namespace.Name
		name = "node-feature-discovery-" + string(uuid.NewUUID())
		image = fmt.Sprintf("quay.io/kubernetes_incubator/node-feature-discovery:%s", imageTag)
		By("Selecting a non-master node")
		if node == nil {
			// Select the first non-master node.
			nodes, err := f.ClientSet.Core().Nodes().List(api.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			for _, n := range nodes.Items {
				if !system.IsMasterNode(&n) {
					node = &n
					break
				}
			}
		}
		Expect(node).NotTo(BeNil())
	})

	It("Should decorate the selected node with the fake feature labels", func() {
		By("Creating a node-feature-discovery pod on the selected node")
		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name: name,
			},
			Spec: api.PodSpec{
				NodeName: node.Name,
				Containers: []api.Container{
					{
						Name:    name,
						Image:   image,
						Command: []string{"/go/bin/node-feature-discovery", "--source=fake"},
						Env: []api.EnvVar{
							{
								Name: "POD_NAME",
								ValueFrom: &api.EnvVarSource{
									FieldRef: &api.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name: "POD_NAMESPACE",
								ValueFrom: &api.EnvVarSource{
									FieldRef: &api.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
						},
					},
				},
				RestartPolicy: api.RestartPolicyNever,
			},
		}

		_, err := f.ClientSet.Core().Pods(ns).Create(pod)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for the node-feature-discovery pod to succeed")
		Expect(framework.WaitForPodSuccessInNamespace(f.ClientSet, name, ns))

		By("Making sure the selected node was decorated with the fake feature labels")
		options := api.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set(fakeFeatureLabels))}
		matchedNodes, err := f.ClientSet.Core().Nodes().List(options)
		Expect(len(matchedNodes.Items)).To(Equal(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(matchedNodes.Items[0].Name).To(Equal(node.Name))

		By("Removing the fake feature and version labels advertised by the node-feature-discovery pod")
		for key := range fakeFeatureLabels {
			framework.RemoveLabelOffNode(f.ClientSet, node.Name, key)
		}
		framework.RemoveLabelOffNode(f.ClientSet, node.Name, "node.alpha.intel.com/node-feature-discovery.version")
	})
})
