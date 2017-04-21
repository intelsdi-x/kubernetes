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
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"strconv"
	"sync"

	"github.com/golang/glog"
	proto "github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	hugepage "github.com/opencontainers/runc/libcontainer/configs"
)

type EventDispatcherEventType int

const (
	ISOLATOR_LIST_CHANGED EventDispatcherEventType = iota
)

type EventDispatcherEvent struct {
	Type EventDispatcherEventType
	Body string
}

// EventDispatcher manages a set of registered isolators and dispatches
// lifecycle events to them.
type EventDispatcher interface {
	// PreStartPod is invoked after the pod sandbox is created but before any
	// of a pod's containers are started.
	PreStartPod(pod *v1.Pod, cgroupPath string) (*lifecycle.EventReply, error)

	// PostStopPod is invoked after all of a pod's containers have permanently
	// stopped running, but before the pod sandbox is destroyed.
	PostStopPod(cgroupPath string) error

	// PreStartContainer is invoked before the container is created.
	PreStartContainer(podName, containerName string) (*lifecycle.EventReply, error)

	// PostStopContainer is invoked after a container has terminated. The
	// container sandbox may or may not exist.
	PostStopContainer(podName, containerName string) error

	// Start starts the dispatcher. After the dispatcher is started, isolators
	// can register themselves to receive lifecycle events.
	Start(socketAddress string)

	// Get communication channel.
	GetEventChannel() chan EventDispatcherEvent
}

// Represents a registered isolator
type registeredIsolator struct {
	// name by which the isolator registered itself, unique
	name string
	// location of the isolator service.
	socketAddress string
}

type eventDispatcher struct {
	sync.Mutex
	started      bool
	isolators    map[string]*registeredIsolator
	eventChannel chan EventDispatcherEvent
}

var eventDispatcherEnabled bool
var dispatcherSingleton EventDispatcher
var once sync.Once

// Enables the event dispatcher subsystem for this Kubelet instance.
//
// This function must be called before the first call to
// GetEventDispatcherSingleton(), otherwise that and all subsequent calls
// will return a noop event dispatcher.
func EnableEventDispatcher() {
	eventDispatcherEnabled = true
}

func GetEventDispatcherSingleton() EventDispatcher {
	once.Do(func() {
		if eventDispatcherEnabled {
			dispatcherSingleton = &eventDispatcher{
				isolators:    map[string]*registeredIsolator{},
				eventChannel: make(chan EventDispatcherEvent),
			}
			dispatcherSingleton.Start(":5433") // "life" on a North American keypad
		} else {
			dispatcherSingleton = &eventDispatcherNoop{}
		}
	})
	return dispatcherSingleton
}

func (ed *eventDispatcher) GetEventChannel() chan EventDispatcherEvent {
	return ed.eventChannel
}

func (ed *eventDispatcher) dispatchEvent(ev *lifecycle.Event) (*lifecycle.EventReply, error) {
	mergedReplies := &lifecycle.EventReply{}
	var errlist []error
	// TODO(CD): Re-evaluate nondeterministic delegation order arising
	//           from Go map iteration.
	for name, isolator := range ed.isolators {
		// TODO(CD): Improve this by building a cancelable context
		ctx := context.Background()

		// Create a gRPC client connection
		// TODO(CD): Use SSL to connect to isolators
		cxn, err := grpc.Dial(isolator.socketAddress, grpc.WithInsecure())
		if err != nil {
			glog.Fatalf("failed to connect to isolator [%s] at [%s]: %v", isolator.name, isolator.socketAddress, err)
			errlist = append(errlist, err)
		}
		defer cxn.Close()
		client := lifecycle.NewIsolatorClient(cxn)

		glog.Infof("Dispatching to isolator: %s", name)
		reply, err := client.Notify(ctx, ev)
		if err != nil {
			errlist = append(errlist, err)
		}
		proto.Merge(mergedReplies, reply)

	}
	return mergedReplies, utilerrors.NewAggregate(errlist)
}

func (ed *eventDispatcher) updateIsolators() {
	isolators := make([]string, len(ed.isolators))
	idx := 0
	for isolator := range ed.isolators {
		isolators[idx] = isolator
		idx += 1
	}
	ed.eventChannel <- EventDispatcherEvent{
		Type: ISOLATOR_LIST_CHANGED,
		Body: strings.Join(isolators, "."),
	}
}

func (ed *eventDispatcher) PreStartPod(pod *v1.Pod, cgroupPath string) (*lifecycle.EventReply, error) {
	jsonPod, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	// construct and send an event
	return ed.dispatchEvent(
		&lifecycle.Event{
			Kind:    lifecycle.Event_POD_PRE_START,
			PodName: pod.Name,
			Pod:     jsonPod,
			CgroupInfo: &lifecycle.CgroupInfo{
				Kind: lifecycle.CgroupInfo_POD,
				Path: cgroupPath,
			},
		})
}

func (ed *eventDispatcher) PostStopPod(cgroupPath string) error {
	// construct and send an event
	_, err := ed.dispatchEvent(
		&lifecycle.Event{
			Kind: lifecycle.Event_POD_POST_STOP,
			CgroupInfo: &lifecycle.CgroupInfo{
				Kind: lifecycle.CgroupInfo_POD,
				Path: cgroupPath,
			},
		})
	return err
}

func (ed *eventDispatcher) PreStartContainer(podName, containerName string) (*lifecycle.EventReply, error) {
	// construct and send an event
	return ed.dispatchEvent(
		&lifecycle.Event{
			Kind:          lifecycle.Event_CONTAINER_PRE_START,
			PodName:       podName,
			ContainerName: containerName,
		})
}

func (ed *eventDispatcher) PostStopContainer(podName, containerName string) error {
	// construct and send an event
	_, err := ed.dispatchEvent(
		&lifecycle.Event{
			Kind:          lifecycle.Event_CONTAINER_POST_STOP,
			PodName:       podName,
			ContainerName: containerName,
		})
	return err
}

func (ed *eventDispatcher) Start(socketAddress string) {
	ed.Lock()
	defer ed.Unlock()

	if ed.started {
		glog.Info("event dispatcher is already running")
		return
	}

	glog.Infof("starting event dispatcher server at [%s]", socketAddress)

	// Set up server bind address.
	lis, err := net.Listen("tcp", socketAddress)
	if err != nil {
		glog.Fatalf("failed to bind to socket address: %v", err)
	}

	// Create a grpc.Server.
	s := grpc.NewServer()

	// Register self as KubeletEventDispatcherServer.
	lifecycle.RegisterEventDispatcherServer(s, ed)

	// Start listen in a separate goroutine.
	go func() {
		if err := s.Serve(lis); err != nil {
			glog.Fatalf("failed to start event dispatcher server: %v", err)
		}
	}()

	ed.started = true
}

func (ed *eventDispatcher) Register(ctx context.Context, request *lifecycle.RegisterRequest) (*lifecycle.RegisterReply, error) {
	// Create a registeredIsolator instance
	isolator := &registeredIsolator{
		name:          request.Name,
		socketAddress: request.SocketAddress,
	}

	glog.Infof("attempting to register isolator [%s]", isolator.name)

	// Check registered name for uniqueness
	reg := ed.isolator(isolator.name)
	if reg != nil {
		glog.Infof("re-registering isolator [%s]", isolator.name)
	}
	// Save registeredIsolator
	ed.isolators[isolator.name] = isolator
	ed.updateIsolators()

	return &lifecycle.RegisterReply{}, nil
}

func (ed *eventDispatcher) Unregister(ctx context.Context, request *lifecycle.UnregisterRequest) (*lifecycle.UnregisterReply, error) {
	reg := ed.isolator(request.Name)
	if reg == nil {
		msg := fmt.Sprintf("unregistration failed: no isolator named [%s] is currently registered.", request.Name)
		glog.Warning(msg)
		return &lifecycle.UnregisterReply{Error: msg}, nil
	}
	delete(ed.isolators, request.Name)
	ed.updateIsolators()
	return &lifecycle.UnregisterReply{}, nil
}

// Returns a pointer to an updated copy of the supplied resource config,
// based on the isolation controls in the event reply. The original resource
// config is not updated in-place.
func ResourceConfigFromReply(reply *lifecycle.EventReply, resources *ResourceConfig) *ResourceConfig {
	// This is a safe copy; ResourceConfig contains only pointers to primitives.
	updatedResources := &ResourceConfig{}
	*updatedResources = *resources

	for _, control := range reply.IsolationControls {
		switch control.Kind {
		case lifecycle.IsolationControl_CGROUP_CPUSET_CPUS:
			updatedResources.CpusetCpus = &control.Value
		case lifecycle.IsolationControl_CGROUP_CPUSET_MEMS:
			updatedResources.CpusetMems = &control.Value
		case lifecycle.IsolationControl_CGROUP_HUGETLB_2MB_LIMIT, lifecycle.IsolationControl_CGROUP_HUGETLB_1GB_LIMIT:
			limit, err:= strconv.ParseUint(control.Value,10,64)
			if err != nil {
				glog.Warningf("Invalid value in isolation control [%s][%s], skipping", control.Kind, control.Value)
				continue
			}

			hugePageLimit := &hugepage.HugepageLimit{
				Pagesize: hugePageSizeStringFromKind(control.Kind),
				Limit:    limit}
			updatedResources.HugetlbLimit = append(updatedResources.HugetlbLimit, hugePageLimit)
		default:
			glog.Warningf("ignoring unknown isolation control kind [%s]", control.Kind)
		}
	}
	return updatedResources
}

// Updates the supplied container config in-place based on the isolation
// controls in the event reply.
func UpdateContainerConfigWithReply(reply *lifecycle.EventReply, config *runtime.ContainerConfig) {
	// Append environment variables to container config.
	for _, control := range reply.IsolationControls {
		switch control.Kind {
		case lifecycle.IsolationControl_CONTAINER_ENV_VAR:
			for k, v := range control.MapValue {
				entry := &runtime.KeyValue{Key: k, Value: v}
				config.Envs = append(config.Envs, entry)
			}
		default:
			continue
		}
	}

	if config.Linux == nil {
		glog.Infof("skipping Linux-only isolation settings")
		return
	}

	// Apply linux-specific settings to container config.
	for _, control := range reply.IsolationControls {
		switch control.Kind {
		case lifecycle.IsolationControl_CGROUP_CPUSET_CPUS:
			config.Linux.Resources.CpusetCpus = control.Value
		case lifecycle.IsolationControl_CGROUP_CPUSET_MEMS:
			config.Linux.Resources.CpusetMems = control.Value
		default:
			glog.Infof("encountered unknown isolation control kind: [%d]", control.Kind)
			continue
		}
	}
}

func (ed *eventDispatcher) isolator(name string) *registeredIsolator {
	for _, isolator := range ed.isolators {
		if isolator.name == name {
			return isolator
		}
	}
	return nil
}

<<<<<<< a27e8e6ee32f1f5050908fce766e359db8a564f1
// eventDispatcherNoop implements EventDispatcher interface.
// It is a no-op implementation and basically does nothing
// eventDispatcherNoop is used in case the QoS cgroup Hierarchy is
// enabled but enable-extended-isolation is not
type eventDispatcherNoop struct {
}

// Make sure that eventDispatcherNoop implements the EventDispatcher interface
var _ EventDispatcher = &eventDispatcherNoop{}

func (ed *eventDispatcherNoop) PreStartPod(pod *v1.Pod, cgroupPath string) (*lifecycle.EventReply, error) {
	return nil, nil
}

func (ed *eventDispatcherNoop) PostStopPod(cgroupPath string) error {
	return nil
}

func (ed *eventDispatcherNoop) PreStartContainer(podName, containerName string) (*lifecycle.EventReply, error) {
	return nil, nil
}

func (ed *eventDispatcherNoop) PostStopContainer(podName, containerName string) error {
	return nil
}

func (ed *eventDispatcherNoop) Start(socketAddress string) {

}

func (ed *eventDispatcherNoop) GetEventChannel() chan EventDispatcherEvent {
	return nil
}

func hugePageSizeStringFromKind(kind lifecycle.IsolationControl_Kind) string {
	if kind == lifecycle.IsolationControl_CGROUP_HUGETLB_2MB_LIMIT {
		return "2MB"
	} else if kind == lifecycle.IsolationControl_CGROUP_HUGETLB_1GB_LIMIT {
		return "1GB"
	} else {
		glog.Errorf("Invalid isolation control kind for HugePages [%s]",kind)
		return "Invalid"
	}
}
