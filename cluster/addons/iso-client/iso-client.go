package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	// For newer version of k8s use this package
	//	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
	"k8s.io/kubernetes/pkg/util/uuid"
)

const (
	// kubelet eventDispatcher address
	eventDispatcherAddress = "localhost:5433"
	// iso-client own address
	eventHandlerLocalAddress = "localhost:5444"
	// name of isolator
	name = "iso"
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
	glog.Info("Starting serving")
	if err := e.GrpcServer.Serve(e.Socket); err != nil {
		glog.Fatalf("EventHandler server stopped serving : %v", err)
		wg.Done()
	}
	wg.Done()
	glog.Info("Stopping eventHandlerServer")
}

// TODO: implement PostStop
func (e *eventHandler) Notify(context context.Context, event *lifecycle.Event) (reply *lifecycle.EventReply, err error) {
	switch event.Kind {
	case lifecycle.Event_POD_PRE_START:
		glog.Infof("Received PreStart event with such payload: %v\n", event.CgroupInfo)
		var pod api.Pod
		err := json.Unmarshal(event.Pod, &pod)
		if err != nil {
			glog.Fatalf("Something went wrong: %v", err)
			return nil, err
		}
		//TODO: Clean this mess up
		glog.Infof("Pod annotations is: %v", pod.Annotations)
		isoApiAnnoJson := pod.Annotations["pod.alpha.kubernetes.io/isolation-api"]
		var isoApiAnno map[string]string
		json.Unmarshal([]byte(isoApiAnnoJson), &isoApiAnno)
		glog.Infof("Pod isolation-api is: %v", isoApiAnno)
		if isoApiAnno["cpu-pinning"] != "" {
			glog.Infof("Pod %s is managed by this isolator. Value of cpu-pinning: %s", pod.Name, isoApiAnno["cpu-pinning"])
			// Pinning created pod to static 0-1 core
			path := fmt.Sprintf("%s%s%s", "/sys/fs/cgroup/cpuset", event.CgroupInfo.Path, "/cpuset.cpus")
			glog.Infof("Our path: %v", path)
			err = ioutil.WriteFile(path, []byte(isoApiAnno["cpu-pinning"]), 0644)
			if err != nil {
				return nil, fmt.Errorf("Ooops: %v", err)
			}

			return &lifecycle.EventReply{
				Error:      "",
				CgroupInfo: event.CgroupInfo,
			}, nil
		} else {
			glog.Infof("Pod %s is not managed by this isolator", pod.Name)
			return &lifecycle.EventReply{
				Error:      "",
				CgroupInfo: event.CgroupInfo,
			}, nil

		}
	default:
		return nil, fmt.Errorf("Wrong event type")
	}

}

// TODO: handle more than just unregistering  evenDispatcherClient
func handleSIGTERM(c chan os.Signal, client *eventDispatcherClient) {
	<-c
	unregisterRequest := &lifecycle.UnregisterRequest{
		Name:  client.Name,
		Token: client.Token,
	}
	rep, err := client.Unregister(client.Ctx, unregisterRequest)
	if err != nil {
		glog.Fatalf("Failed to unregister handler: %v")
		os.Exit(1)
	}
	glog.Infof("Unregistering iso: %v\n", rep)

	os.Exit(0)

}

type eventDispatcherClient struct {
	Token   string
	Name    string
	Address string
	Ctx     context.Context
	// EventDispatcherClient specified in protobuf
	lifecycle.EventDispatcherClient
}

// Constructor for eventDispatcherClient
func NewEventDispatcherClient(name string, address string) (*eventDispatcherClient, error) {
	clientConn, err := grpc.Dial(eventDispatcherAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &eventDispatcherClient{
			EventDispatcherClient: lifecycle.NewEventDispatcherClient(clientConn),
			Ctx:     context.Background(),
			Name:    name,
			Address: address,
		},
		nil

}

// Create RegisterRequest and register eventDispatcherClient
func (edc *eventDispatcherClient) Register() (reply *lifecycle.RegisterReply, err error) {
	registerToken := string(uuid.NewUUID())
	registerRequest := &lifecycle.RegisterRequest{
		SocketAddress: edc.Address,
		Name:          edc.Name,
		Token:         registerToken,
	}

	glog.Infof("Attempting to register evenDispatcherClient. Request: %v", registerRequest)
	reply, err = edc.EventDispatcherClient.Register(edc.Ctx, registerRequest)
	if err != nil {
		return reply, err
	}
	edc.Token = reply.Token
	return reply, nil
}

// TODO: split it to smaller functions
func main() {
	flag.Parse()
	glog.Info("Starting ...")
	server := NewEventHandler(name, eventHandlerLocalAddress)
	err := server.RegisterEventHandler()
	if err != nil {
		glog.Fatalf("Cannot register EventHandler: %v", err)
		os.Exit(1)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go server.Serve(wg)

	// Sening address of local eventHandlerServer
	client, err := NewEventDispatcherClient(name, eventHandlerLocalAddress)
	if err != nil {
		glog.Fatalf("Cannot create eventDispatcherClient: %v", err)
		os.Exit(1)
	}

	reply, err := client.Register()
	if err != nil {
		glog.Fatalf("Failed to register handler: %v . Reply: %v", err, reply)
		os.Exit(1)
	}
	glog.Infof("Registering eventDispatcherClient. Reply: %v", reply)

	// Handling SigTerm
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleSIGTERM(c, client)

	wg.Wait()
}
