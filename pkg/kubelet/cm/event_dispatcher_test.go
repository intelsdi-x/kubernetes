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
	"sort"
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
	ed := newEventDispatcher()
	testCases := []struct {
		registrationEvents       []*lifecycle.RegisterRequest
		requestedIsolators       []string
		requestedSockets         []string
		requestedIsolatorsNumber int
	}{
		{
			registrationEvents:       []*lifecycle.RegisterRequest{},
			requestedIsolators:       []string{},
			requestedSockets:         []string{},
			requestedIsolatorsNumber: 0,
		},
		{
			registrationEvents: []*lifecycle.RegisterRequest{
				{
					Name:          "isolator1",
					SocketAddress: "socket1",
				},
			},
			requestedIsolators:       []string{"isolator1"},
			requestedSockets:         []string{"socket1"},
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
			requestedIsolators:       []string{"isolator1", "isolator2"},
			requestedSockets:         []string{"socket1", "socket2"},
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
			requestedIsolators:       []string{"isolator1", "isolator2"},
			requestedSockets:         []string{"socket1", "socket2"},
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

		for registeredIsolator := range ed.isolators {
			if sort.SearchStrings(testCase.requestedIsolators, ed.isolators[registeredIsolator].name) < 0 {
				t.Errorf("Isolator: %q has not been found in expected isolators %q",
					registeredIsolator,
					testCase.requestedIsolators)
			}
			if sort.SearchStrings(testCase.requestedSockets, ed.isolators[registeredIsolator].socketAddress) < 0 {
				t.Errorf("Isolator's socket: %q has not been found in expected sockets %q",
					ed.isolators[registeredIsolator].socketAddress,
					testCase.requestedSockets)
			}
		}
	}
}

func TestEventDispatcher_Unregister(t *testing.T) {
	testCases := []struct {
		isolators    []*lifecycle.RegisterRequest
		isolatorName string
	}{
		{
			isolators: []*lifecycle.RegisterRequest{
				{
					Name:          "test1",
					SocketAddress: "socket1",
				},
				{
					Name:          "test2",
					SocketAddress: "socket2",
				},
			},
			isolatorName: "test1",
		},
		{
			isolators: []*lifecycle.RegisterRequest{
				{
					Name:          "test1",
					SocketAddress: "socket1",
				},
				{
					Name:          "test2",
					SocketAddress: "socket2",
				},
			},
			isolatorName: "test3",
		},
		{
			isolators: []*lifecycle.RegisterRequest{
				{
					Name:          "test1",
					SocketAddress: "socket1",
				},
				{
					Name:          "test2",
					SocketAddress: "socket2",
				},
				{
					Name:          "test2",
					SocketAddress: "socket2",
				},
				{
					Name:          "test3",
					SocketAddress: "socket3",
				},
			},
			isolatorName: "test2",
		},
		{
			isolators:    []*lifecycle.RegisterRequest{},
			isolatorName: "test1",
		},
	}

	for _, testCase := range testCases {
		ed := newEventDispatcher()
		ed.isolators = make(map[string]*registeredIsolator)
		for _, isolator := range testCase.isolators {
			go getIsolators(ed.GetEventChannel(), &EventDispatcherEvent{})
			ed.Register(ctx.Background(), isolator)
		}
		var event EventDispatcherEvent
		go getIsolators(ed.GetEventChannel(), &event)
		ed.Unregister(context.Background(), &lifecycle.UnregisterRequest{Name: testCase.isolatorName})
		if ed.isolators[testCase.isolatorName] != nil {
			t.Error("Unregistration failed: expected item to remove is still available")
		}
	}
}

type dummy struct {
	replyErr  string
	serverErr error
	lastEvent *lifecycle.Event
}

func (d *dummy) Notify(context ctx.Context, event *lifecycle.Event) (*lifecycle.EventReply, error) {
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

func (d *dummy) setReplyError(err string) {
	d.replyErr = err
}

func (d *dummy) setServerError(err error) {
	d.serverErr = err
}

func (d *dummy) getEvent() *lifecycle.Event {
	return d.lastEvent
}

type grpcDummyServer struct {
	server   *grpc.Server
	listener net.Listener
	object   *dummy
}

func (g grpcDummyServer) SetReplyError(err string) {
	g.object.replyErr = err
}

func (g grpcDummyServer) SetServerError(err error) {
	g.object.serverErr = err
}

func (g grpcDummyServer) GetAddress() string {
	return g.listener.Addr().String()
}

func (g grpcDummyServer) GetLastEventFromServer() *lifecycle.Event {
	return g.object.lastEvent
}

func (g grpcDummyServer) ResetDummyObject() {
	g.object.lastEvent = nil
	g.object.serverErr = nil
	g.object.replyErr = ""
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
		server:   grpc.NewServer(),
		listener: lis,
		object:   &dummy{},
	}

	lifecycle.RegisterIsolatorServer(grpcDummy.server, grpcDummy.object)
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

	ed := newEventDispatcher()
	ed.isolators = make(map[string]*registeredIsolator)
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
		server.ResetDummyObject()
		server.SetReplyError(testCase.replyError)
		server.SetServerError(testCase.serverError)

		eventReply, err := ed.PreStartPod(pod, "/")
		if err != nil {
			if !strings.Contains(err.Error(), testCase.serverError.Error()) {
				t.Errorf("Expected error(%q) isn't message of grpc error(%q)",
					testCase.serverError,
					err)
			}
			if eventReply.Error != "" {
				t.Errorf("EventReply Error should be \"\" but it wasn't: %q",
					eventReply.Error)
			}
		} else {
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

	ed := newEventDispatcher()
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
		server.ResetDummyObject()
		server.SetServerError(testCase.serverError)

		err := ed.PostStopPod("/")
		if err != nil {
			if !strings.Contains(err.Error(), testCase.serverError.Error()) {
				t.Errorf("Expected error(%q) isn't message of grpc error(%q)",
					testCase.serverError,
					err)
			}
		}
		if server.GetLastEventFromServer() == nil {
			t.Error("Lifecycle object should be accessable in lifecycle handler, but it wasn't")
		}

	}
}

func TestEventDispatcher_ResourceConfigFromReplies(t *testing.T) {
	testCases := []struct {
		isolators     []*lifecycle.IsolationControl
		resources     *ResourceConfig
		shouldBe      string
		requestedKind lifecycle.IsolationControl_Kind
	}{
		{
			isolators: []*lifecycle.IsolationControl{
				{
					Kind:  lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
					Value: "5",
				},
			},
			resources:     &ResourceConfig{},
			shouldBe:      "5",
			requestedKind: lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
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
			resources:     &ResourceConfig{},
			shouldBe:      "2",
			requestedKind: lifecycle.IsolationControl_CGROUP_CPUSET_CPUS,
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
			resources:     &ResourceConfig{},
			shouldBe:      "3",
			requestedKind: lifecycle.IsolationControl_CGROUP_CPUSET_MEMS,
		},
	}

	for _, testCase := range testCases {
		er := &lifecycle.EventReply{
			IsolationControls: testCase.isolators,
		}
		ed := newEventDispatcher()
		output := ed.ResourceConfigFromReplies(er, testCase.resources)

		switch testCase.requestedKind {
		case lifecycle.IsolationControl_CGROUP_CPUSET_CPUS:
			if *output.CpusetCpus != testCase.shouldBe {
				t.Errorf("Requested value of CPUSET cpu %q doesn't match recieved one %q",
					*output.CpusetCpus,
					testCase.shouldBe)
			}
			break
		case lifecycle.IsolationControl_CGROUP_CPUSET_MEMS:
			if *output.CpusetMems != testCase.shouldBe {
				t.Errorf("Requested value of CPUSET mem %q doesn't match recieved one %q",
					*output.CpusetCpus,
					testCase.shouldBe)
			}
			break
		}

	}
}
