package main

import (
	"flag"

	"github.com/golang/glog"
	"k8s.io/kubernetes/cluster/addons/iso-client/decorator"
	"k8s.io/kubernetes/cluster/addons/iso-client/isolator"
)

const (
	// Address of event dispatcher in Kubelet
	remoteAddress = ":5433"
	// Address of this client
	localAddress = ":5445"
)

// TODO: split it to smaller functions
func main() {
	// Enable logging to STDERR
	flag.Set("logtostderr", "true")
	flag.Parse()

	decorator, err := decorator.New("env-decorator", map[string]string{"FOO": "BAR"})
	if err != nil {
		glog.Fatalf("unable to create isolator: %q", err)
	}

	// Start isolator server
	err = isolator.StartIsolatorServer(decorator, localAddress, remoteAddress)
	if err != nil {
		glog.Fatalf("unable to initialize isolator: %v", err)
	}
}
