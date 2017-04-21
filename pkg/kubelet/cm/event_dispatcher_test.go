// +build linux

/*
Copyright 2015 The Kubernetes Authors.

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
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	ctx "golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
)

func getIsolators(eventCh chan EventDispatcherEvent, isolators *EventDispatcherEvent) {
	*isolators = <-eventCh
}

func TestEventDispatcher_Register(t *testing.T) {
	EnableEventDispatcher()
	ed := GetEventDispatcherSingleton().(*eventDispatcher)
	testCases := []struct {
		registrationEvents       []*lifecycle.RegisterRequest
		requestedIsolators       map[string]string
		requestedIsolatorsNumber int
	}{
		{
			registrationEvents:       []*lifecycle.RegisterRequest{},
			requestedIsolators:       map[string]string{},
			requestedIsolatorsNumber: 0,
		},
		{
			registrationEvents: []*lifecycle.RegisterRequest{
				{
					Name:          "isolator1",
					SocketAddress: "socket1",
				},
			},
			requestedIsolators: map[string]string{
				"isolator1": "socket1",
			},
			requestedIsolatorsNumber: 1,
		},
		{
			registrationEvents: []*lifecycle.RegisterRequest{
				{
					Name:          "isolator1",
					SocketAddress: "socket1",
				},
				{
					Name:          "isolator2",
					SocketAddress: "socket2",
				},
			},
			requestedIsolators: map[string]string{
				"isolator1": "socket1",
				"isolator2": "socket2",
			},
			requestedIsolatorsNumber: 2,
		},
		{
			registrationEvents: []*lifecycle.RegisterRequest{
				{
					Name:          "isolator1",
					SocketAddress: "socket1",
				},
				{
					Name:          "isolator2",
					SocketAddress: "socket2",
				},
				{
					Name:          "isolator2",
					SocketAddress: "socket2",
				},
			},
			requestedIsolators: map[string]string{
				"isolator1": "socket1",
				"isolator2": "socket2",
			},
			requestedIsolatorsNumber: 2,
		},
	}
	for _, testCase := range testCases {
		var events []EventDispatcherEvent
		ed.isolators = make(map[string]*registeredIsolator)

		for _, req := range testCase.registrationEvents {
			var registrationEvent EventDispatcherEvent
			go getIsolators(ed.GetEventChannel(), &registrationEvent)
			ed.Register(context.Background(), req)

			if registrationEvent.Type != ISOLATOR_LIST_CHANGED {
				t.Error("Event type is different than expected one")
			}

			events = append(events, registrationEvent)
		}

		if len(events) != len(testCase.registrationEvents) {
			t.Error("Number of events are different than number of registrations")
		}
		if len(ed.isolators) != testCase.requestedIsolatorsNumber {
			t.Errorf("Expected number of isolators (%d) is different than real one (%d)",
				len(ed.isolators),
				testCase.requestedIsolatorsNumber,
			)
		}

		for key, value := range testCase.requestedIsolators {
			if ed.isolators[key] == nil {
				t.Errorf("Expected isolator %q hasn't been registered", key)
			}
			if ed.isolators[key].socketAddress != value {
				t.Errorf("Expected socket address %q doesn't match to registered one %q",
					value,
					ed.isolators[key].socketAddress)
			}

		}
	}
}

func TestEventDispatcher_Unregister(t *testing.T) {
	EnableEventDispatcher()
	ed := GetEventDispatcherSingleton().(*eventDispatcher)

	testCases := []struct {
		isolators    []*lifecycle.UnregisterRequest
		isolatorName string
	}{
		{
			isolators: []*lifecycle.UnregisterRequest{
				{
					Name: "test1",
				},
				{
					Name: "test2",
				},
			},
			isolatorName: "test1",
		},
		{
			isolators: []*lifecycle.UnregisterRequest{
				{
					Name: "test1",
				},
				{
					Name: "test2",
				},
			},
			isolatorName: "test3",
		},
		{
			isolators: []*lifecycle.UnregisterRequest{
				{
					Name: "test1",
				},
				{
					Name: "test2",
				},
				{
					Name: "test2",
				},
				{
					Name: "test3",
				},
			},
			isolatorName: "test2",
		},
		{
			isolators:    []*lifecycle.UnregisterRequest{},
			isolatorName: "test1",
		},
	}

	for _, testCase := range testCases {
		ed.isolators = make(map[string]*registeredIsolator)
		for _, isolator := range testCase.isolators {
			go getIsolators(ed.GetEventChannel(), &EventDispatcherEvent{})
			ed.Register(ctx.Background(), &lifecycle.RegisterRequest{
				Name:          isolator.Name,
				SocketAddress: "dummyIsolator address",
			})
		}
		var event EventDispatcherEvent
		go getIsolators(ed.GetEventChannel(), &event)
		ed.Unregister(context.Background(), &lifecycle.UnregisterRequest{Name: testCase.isolatorName})
		if ed.isolators[testCase.isolatorName] != nil {
			t.Error("Unregistration failed: expected item to remove is still available")
		}
		if event.Type != ISOLATOR_LIST_CHANGED {
			t.Error("Recieved event Kind wasn't expected one")
		}
	}
}

type dummyIsolator struct {
	replyErr  string
	serverErr error
	lastEvent *lifecycle.Event
}

func (d *dummyIsolator) Notify(context ctx.Context, event *lifecycle.Event) (*lifecycle.EventReply, error) {
	d.lastEvent = event
	return &lifecycle.EventReply{
		IsolationControls: []*lifecycle.IsolationControl{
			{
				Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
				Value: "0",
			},
		},
		Error: d.replyErr,
	}, d.serverErr
}

func (d *dummyIsolator) setReplyError(err string) {
	d.replyErr = err
}

func (d *dummyIsolator) setServerError(err error) {
	d.serverErr = err
}

func (d *dummyIsolator) getEvent() *lifecycle.Event {
	return d.lastEvent
}

type grpcDummyServer struct {
	server        *grpc.Server
	listener      net.Listener
	dummyIsolator *dummyIsolator
}

func (g grpcDummyServer) GetAddress() string {
	return g.listener.Addr().String()
}

func (g grpcDummyServer) GetLastEventFromServer() *lifecycle.Event {
	return g.dummyIsolator.lastEvent
}

func (g grpcDummyServer) ResetDummyIsolator() {
	g.dummyIsolator.lastEvent = nil
	g.dummyIsolator.serverErr = nil
	g.dummyIsolator.replyErr = ""
}

func (g grpcDummyServer) Close() {
	g.server.Stop()
	g.listener.Close()
}

func runDummyServer() (*grpcDummyServer, error) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	grpcDummy := grpcDummyServer{
		server:        grpc.NewServer(),
		listener:      lis,
		dummyIsolator: &dummyIsolator{},
	}

	lifecycle.RegisterIsolatorServer(grpcDummy.server, grpcDummy.dummyIsolator)
	go grpcDummy.server.Serve(lis)
	return &grpcDummy, nil
}

func TestEventDispatcher_PreStartPod(t *testing.T) {
	server, err := runDummyServer()
	if err != nil {
		t.Skipf("Cannot run GRPC server: %q", err)
	}
	defer server.Close()

	isolator := &lifecycle.RegisterRequest{
		Name:          "isolator1",
		SocketAddress: server.GetAddress(),
	}

	pod := &v1.Pod{}
	EnableEventDispatcher()
	ed := GetEventDispatcherSingleton().(*eventDispatcher)
	ed.isolators = make(map[string]*registeredIsolator)
	go getIsolators(ed.GetEventChannel(), &EventDispatcherEvent{})
	ed.Register(ctx.Background(), isolator)

	testCases := []struct {
		replyError  string
		serverError error
	}{
		{
			replyError:  "Test Error",
			serverError: nil,
		},
		{
			replyError:  "Test Error",
			serverError: fmt.Errorf("Server Error"),
		},
		{
			replyError:  "",
			serverError: fmt.Errorf("Server Error"),
		},
		{
			replyError:  "",
			serverError: nil,
		},
	}

	for _, testCase := range testCases {
		server.ResetDummyIsolator()
		server.dummyIsolator.setReplyError(testCase.replyError)
		server.dummyIsolator.setServerError(testCase.serverError)

		eventReply, err := ed.PreStartPod(pod, "/")

		// Check if GRPC is working fine.
		if err != nil {
			// Check case if GRPC server error contains expected message.
			if !strings.Contains(err.Error(), testCase.serverError.Error()) {
				t.Errorf("Expected error(%q) isn't message of grpc error(%q)",
					testCase.serverError,
					err)
			}
			// GRPC server should return empty EventReply, so we expect that error message is ""
			if eventReply.Error != "" {
				t.Errorf("EventReply Error should be \"\" but it wasn't: %q",
					eventReply.Error)
			}
		} else {
			// In case of GRPC server doesn't return error code, check if EventReply contains expected
			// EvenReply error message.
			if eventReply.Error != testCase.replyError {
				t.Errorf("Expected eventReply error(%q) doesn't match to recived one(%q)",
					testCase.replyError, eventReply.Error)
			}
		}
	}
}

func TestEventDispatcher_PostStopPod(t *testing.T) {
	server, err := runDummyServer()
	if err != nil {
		t.Skipf("Cannot run GRPC server: %q", err)
	}
	defer server.Close()

	isolator := &lifecycle.RegisterRequest{
		Name:          "isolator1",
		SocketAddress: server.GetAddress(),
	}

	EnableEventDispatcher()
	ed := GetEventDispatcherSingleton().(*eventDispatcher)
	ed.isolators = make(map[string]*registeredIsolator)
	ed.Register(ctx.Background(), isolator)

	testCases := []struct {
		serverError error
	}{
		{
			serverError: nil,
		},
		{
			serverError: fmt.Errorf("Server Error"),
		},
	}

	for _, testCase := range testCases {
		server.ResetDummyIsolator()
		server.dummyIsolator.setServerError(testCase.serverError)

		err := ed.PostStopPod("/")

		// Check that GRPC server is returning expected error message.
		if err != nil {
			if !strings.Contains(err.Error(), testCase.serverError.Error()) {
				t.Errorf("Expected error(%q) isn't message of grpc error(%q)",
					testCase.serverError,
					err)
			}
		}

		// Verify that isolator receives not nil lifecycle.Event object
		if server.GetLastEventFromServer() == nil {
			t.Error("Lifecycle dummyIsolator should be accessable in lifecycle handler, but it wasn't")
		}

	}
}

func TestEventDispatcher_ResourceConfigFromReplies(t *testing.T) {
	testCases := []struct {
		isolators []*lifecycle.IsolationControl
		resources map[string]string
	}{
		{
			isolators: []*lifecycle.IsolationControl{
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
					Value: "5",
				},
			},
			resources: map[string]string{
				"CpusetCpus": "5",
			},
		},
		{
			isolators: []*lifecycle.IsolationControl{
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
					Value: "1",
				},
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
					Value: "2",
				},
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_MEMS,
					Value: "3",
				},
			},
			resources: map[string]string{
				"CpusetCpus": "2",
				"CpusetMems": "3",
			},
		},
		{
			isolators: []*lifecycle.IsolationControl{
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_MEMS,
					Value: "1",
				},
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
					Value: "2",
				},
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_MEMS,
					Value: "3",
				},
			},
			resources: map[string]string{
				"CpusetCpus": "2",
				"CpusetMems": "3",
			},
		},
	}

	for _, testCase := range testCases {
		eventReply := &lifecycle.EventReply{
			IsolationControls: testCase.isolators,
		}
		output := ResourceConfigFromReply(eventReply, &ResourceConfig{})

		reflectedStruct := reflect.ValueOf(output)

		for key, value := range testCase.resources {
			data := reflect.Indirect(reflectedStruct).FieldByName(key).Interface().(*string)
			if *data != value {
				t.Errorf("Invalid value of %q. Expected %s, has got %s",
					key,
					value,
					*data)
			}
		}
	}
}
