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

package cm

import (
	"errors"
	"net"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
)

// EventDispatcher manages a set of registered lifecycle event handlers and
// dispatches lifecycle events to them.
type EventDispatcher interface {
	// PreStartPod is invoked after the pod sandbox is created but before any
	// of a pod's containers are started.
	PreStartPod(cgroupPath string) error

	// PostStopPod is invoked after all of a pod's containers have permanently
	// stopped running, but before the pod sandbox is destroyed.
	PostStopPod(cgroupPath string) error

	// Start starts the dispatcher. After the dispatcher is started , handlers
	// can register themselves to receive lifecycle events.
	Start()
}

// Represents a registered event handler
type registeredHandler struct {
	// name by which the handler registered itself, unique
	name string
	// The client end of the event handler service.
	client lifecycle.LifecycleEventHandlerClient
}

type eventDispatcher struct {
	handlers []registeredHandler
}

func newEventDispatcher() *eventDispatcher {
	return &eventDispatcher{
		handlers: []registeredHandler{},
	}
}

func (ed *eventDispatcher) dispatchEvent(cgroupPath, eventName string) error {
	// construct an event
	ev := &lifecycle.Event{
		Name: eventName,
		CgroupInfo: &lifecycle.CgroupInfo{
			Kind: lifecycle.CgroupInfo_POD,
			Path: cgroupPath,
		},
	}

	for _, handler := range ed.handlers {
		// TODO(CD): Improve this by building a cancelable context
		ctx := context.Background()
		glog.Infof("Dispatching to event handler: %s", handler.name)
		reply, err := handler.client.Notify(ctx, ev)
		if err != nil {
			return err
		}
		if reply.Error != "" {
			return errors.New(reply.Error)
		}
	}

	return nil
}

func (ed *eventDispatcher) PreStartPod(cgroupPath string) error {
	return ed.dispatchEvent(cgroupPath, "PRE_START")
}

func (ed *eventDispatcher) PostStopPod(cgroupPath string) error {
	return ed.dispatchEvent(cgroupPath, "POST_STOP")
}

func (ed *eventDispatcher) Start(socketAddress string) {
	// Set up server bind address.
	lis, err := net.Listen("tcp", socketAddress)
	if err != nil {
		glog.Fatalf("failed to bind to socket address: %v", err)
	}

	// Create a grpc.Server.
	s := grpc.NewServer()

	// Register self as KubeletEventDispatcherServer.
	lifecycle.RegisterKubeletEventDispatcherServer(s, ed)

	// Start listen in a separate goroutine.
	go func() {
		if err := s.Serve(lis); err != nil {
			glog.Fatalf("failed to start event dispatcher server: %v", err)
		}
	}()
}

func (ed *eventDispatcher) Register(context.Context, *lifecycle.RegisterRequest) (*lifecycle.RegisterReply, error) {
	// TODO

	// Create a registeredHandler instance
	// Check registered name for uniqueness
	// Add registeredHandler to the slice of handlers
	// Construct and return a reply
	return &lifecycle.RegisterReply{}, nil
}
