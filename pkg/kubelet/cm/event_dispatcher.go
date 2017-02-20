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
	// client end of the event handler service.
	client lifecycle.LifecycleEventHandlerClient
	// token to identify this registration
	token string
}

type eventDispatcher struct {
	handlers []registeredHandler
}

func newEventDispatcher() *eventDispatcher {
	return &eventDispatcher{
		handlers: map[string]registeredHandler{},
	}
}

func (ed *eventDispatcher) dispatchEvent(cgroupPath, kind lifecycle.Event_Kind) error {
	// construct an event
	ev := &lifecycle.Event{
		Kind: kind,
		CgroupInfo: &lifecycle.CgroupInfo{
			Kind: lifecycle.CgroupInfo_POD,
			Path: cgroupPath,
		},
	}

	// TODO(CD): Re-evaluate nondeterministic delegation order arising
	//           from Go map iteration.
	for name, handler := range ed.handlers {
		// TODO(CD): Improve this by building a cancelable context
		ctx := context.Background()
		glog.Infof("Dispatching to event handler: %s", name)
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
	return ed.dispatchEvent(cgroupPath, lifecycle.Event_POD_PRE_START)
}

func (ed *eventDispatcher) PostStopPod(cgroupPath string) error {
	return ed.dispatchEvent(cgroupPath, lifecycle.Event_POD_POST_STOP)
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

func (ed *eventDispatcher) Register(ctx context.Context, request *lifecycle.RegisterRequest) (*lifecycle.RegisterReply, error) {
	// Create a gRPC client connection
	cxn := nil // TODO(CD): Create cxn

	// Create a registeredHandler instance
	h := &registeredHandler{
		name:   request.Name,
		client: lifecycle.NewLifecycleEventHandlerClient(cxn),
		token:  newToken(),
	}

	log.Info("attempting to register event handler [%s]", h.name)

	// Check registered name for uniqueness
	reg := handler(h.name)
	if reg != nil {
		if reg.token != request.token {
			msg := fmt.Sprintf("registration failed: an event handler named [%s] is already registered and the supplied registration token does not match.", reg.name)
			log.Warning(msg)
			return &lifecycle.RegisterReply{Error: msg}, nil
		}
		log.Info("re-registering event handler [%s]", h.name)
	}

	// Save registeredHandler
	ed.handlers[h.name] = h

	return &lifecycle.RegisterReply{Token: h.token}, nil
}

func (ed *eventDispatcher) Unregister(ctx context.Context, request *lifecycle.UnregisterRequest) (*lifecycle.UnregisterReply, error) {
	reg := handler(request.name)
	if reg == nil {
		msg = fmt.Sprintf("unregistration failed: no handler named [%s] is currently registered.", request.name)
		log.Warning(msg)
		return &lifecycle.UnregisterReply{Error: msg}, nil
	}
	if reg.token != request.token {
		msg = fmt.Sprintf("unregistration failed: token mismatch for handler [%s].", request.name)
		log.Warning(msg)
		return &lifecycle.UnregisterReply{Error: msg}, nil
	}
	delete(ed.handlers, request.name)
	return &lifecycle.UnregisterReply{}, nil
}

func (ed *eventDispatcher) handler(name string) *registeredHandler {
	for _, h := range ed.registeredHandlers {
		if h.name == name {
			return h
		}
	}
	return nil
}

func newToken() string {
	// TODO(CD): Generate and return a new UUIDv4 here
	return "foobar"
}

///////////////////////////////////////////////////////////////////////////////
// TODO(CD)
///////////////////////////////////////////////////////////////////////////////
//
// - Handle disconnection
