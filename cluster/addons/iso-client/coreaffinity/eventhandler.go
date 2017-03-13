package coreaffinity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"sync"

	"golang.org/x/net/context"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
)

const (
	cgroupPrefix = "/sys/fs/cgroup/cpuset"
	cgroupSufix  = "/cpuset.cpus"
)

type EventHandler interface {
	RegisterEventHandler() error
	Serve(sync.WaitGroup)
}

type eventHandler struct {
	Name       string
	Address    string
	GrpcServer *grpc.Server
	Socket     net.Listener
}

// Constructor for EventHandler
func NewEventHandler(eventHandlerName string, eventHandlerAddress string) (e *eventHandler) {
	return &eventHandler{
		Name:    eventHandlerName,
		Address: eventHandlerAddress,
	}
}

// Registering eventHandler server
func (e *eventHandler) RegisterEventHandler() (err error) {
	e.Socket, err = net.Listen("tcp", e.Address)
	if err != nil {
		return fmt.Errorf("Failed to bind to socket address: %v", err)
	}
	e.GrpcServer = grpc.NewServer()

	lifecycle.RegisterEventHandlerServer(e.GrpcServer, e)
	glog.Info("EvenHandler Server has been registered")

	return nil

}

// Start serving Grpc
func (e *eventHandler) Serve(wg sync.WaitGroup) {
	defer wg.Done()
	glog.Info("Starting serving")
	if err := e.GrpcServer.Serve(e.Socket); err != nil {
		glog.Fatalf("EventHandler server stopped serving : %v", err)
	}
	glog.Info("Stopping eventHandlerServer")
}

// isolation api
type isoSpec struct {
	CoreAffinity string `json:"core-affinity"`
}

// set cpuset
func setCpuSet(cgroupPath string, value string) error {
	path := fmt.Sprintf("%s%s%s", cgroupPrefix, cgroupPath, cgroupSufix)
	err := ioutil.WriteFile(path, []byte(value), 0644)
	if err != nil {
		return fmt.Errorf("Cannot set %s : %v", path, err)
	}
	return nil
}

// extract Pod object from Event
func getPod(bytePod []byte) (pod *api.Pod, err error) {
	pod = &api.Pod{}
	err = json.Unmarshal(bytePod, pod)
	if err != nil {
		glog.Fatalf("Cannot Unmarshal pod: %v", err)
		return
	}
	return
}

// check wheter pod should be isolated
func isIsoPod(pod *api.Pod) bool {
	if pod.Annotations["pod.alpha.kubernetes.io/isolation-api"] != "" {
		return true
	}
	return false
}

// extract isoSpec from annotations
func getIsoSpec(annotations map[string]string) (spec *isoSpec, err error) {
	spec = &isoSpec{}
	err = json.Unmarshal([]byte(annotations["pod.alpha.kubernetes.io/isolation-api"]), spec)
	if err != nil {
		glog.Fatalf("Cannot unmarshal isoSpec: %v", err)
		return nil, err
	}
	return
}

// validate isoSpec
func validateIsoSpec(spec *isoSpec) (err error) {
	if spec.CoreAffinity == "" {
		return fmt.Errorf("Required field core-affinity is missing.Given json: %v", spec)
	}
	return nil
}

// TODO: implement PostStop
func (e *eventHandler) Notify(context context.Context, event *lifecycle.Event) (reply *lifecycle.EventReply, err error) {
	switch event.Kind {
	case lifecycle.Event_POD_PRE_START:
		glog.Infof("Received PreStart event: %v\n", event.CgroupInfo)
		pod, err := getPod(event.Pod)
		if err != nil {
			return &lifecycle.EventReply{
				Error:      err.Error(),
				CgroupInfo: event.CgroupInfo,
			}, err
		}

		//TODO: Clean this mess up
		if !isIsoPod(pod) {
			glog.Infof("Pod %s is not managed by this isolator", pod.Name)
			return nil, nil
		}

		glog.Infof("Pod %s is managed by this isolator", pod.Name)
		spec, err := getIsoSpec(pod.Annotations)
		if err != nil {
			return &lifecycle.EventReply{
				Error:      err.Error(),
				CgroupInfo: event.CgroupInfo,
			}, err
		}
		// TODO: Decide whether typo should error POD or not
		err = validateIsoSpec(spec)
		if err != nil {
			return &lifecycle.EventReply{
				Error:      err.Error(),
				CgroupInfo: event.CgroupInfo,
			}, err
		}
		glog.Infof("Pod %s is valid. Value of core-affinity: %s", pod.Name, spec.CoreAffinity)
		err = setCpuSet(event.CgroupInfo.Path, spec.CoreAffinity)
		if err != nil {
			return &lifecycle.EventReply{
				Error:      err.Error(),
				CgroupInfo: event.CgroupInfo,
			}, err
		}

		return &lifecycle.EventReply{
			Error:      "",
			CgroupInfo: event.CgroupInfo,
		}, nil
	default:
		return nil, fmt.Errorf("Wrong event type")
	}

}
