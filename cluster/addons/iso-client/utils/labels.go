package utils

import (
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeapi "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func generateIsolatorName(name string) string {
	return fmt.Sprintf("external-isolator-%s", name)
}

func applyLabelChange(isolatorName string, add bool) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("cannot get in-cluster config: %q", err)
	}

	nodename, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("cannot gather hostname: %q", err)
	}

	clientset, err := kubeapi.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("cannot achieve client instance for Kubernetes: %s", err)
	}

	client := clientset.CoreV1()
	node, err := client.Nodes().Get(nodename, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot get node %q: %q", nodename, err)
	}

	labels := node.GetLabels()
	if add {
		labels[isolatorName] = ""
	} else {
		delete(labels, isolatorName)
	}

	node.SetLabels(labels)
	_, err = client.Nodes().Update(node)
	return err
}

func SetLabel(isolatorName string) error {
	return applyLabelChange(generateIsolatorName(isolatorName), true)
}

func UnsetLabel(isolatorName string) error {
	return applyLabelChange(generateIsolatorName(isolatorName), false)
}
