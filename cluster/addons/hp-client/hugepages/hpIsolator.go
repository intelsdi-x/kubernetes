package hugepages

import (
	"encoding/json"
	"fmt"
	"strings"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/kubernetes/cluster/addons/hp-client/hugepages/discovery"
	"k8s.io/kubernetes/cluster/addons/iso-client/opaque"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/api/v1/helper"
	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle"
)

type hugepageIsolator struct {
	name			string
	hpTotal			int
}

type isoSpec struct {
	Hugepage string `json:"hugepage"`
}

const (
	MegaBytesToBytes = 1048576
)

// implementation of Init() method in  "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (h *hugepageIsolator) Init() error {
	return opaque.AdvertiseOpaqueResource(h.name, h.hpTotal)
}

// Constructor for Isolator
func New(name string) (*hugepageIsolator, error) {
	totalHPAvail, err := discovery.DiscoverHugepageNumber()
	return &hugepageIsolator{
		name:		name,
		hpTotal:	totalHPAvail,
	}, err
}

func (h *hugepageIsolator) gatherContainerRequest(container v1.Container) int64 {
	resource, ok := container.Resources.Requests[helper.OpaqueIntResourceName(h.name)]
	if !ok {
		return 0
	}
	return resource.Value()
}

func (h *hugepageIsolator) countHugepagesFromOIR(pod *v1.Pod) int64 {
        var hpAccu int64
        for _, container := range pod.Spec.Containers {
                hpAccu = hpAccu + h.gatherContainerRequest(container)
        }
        return hpAccu
}

func getAnnotationFromPod(pod *v1.Pod) (*isoSpec, error){
	glog.Infof("Inside getAnnotationFromPod function Value of pod.Annotations is %v", pod.Annotations)
	spec := &isoSpec{}
	if pod.Annotations["pod.alpha.kubernetes.io/isolation-api"] != "" {
		err := json.Unmarshal([]byte(pod.Annotations["pod.alpha.kubernetes.io/isolation-api"]), spec)
		if err != nil {
			glog.Fatalf("Cannot unmarshal isoSpec: %v", err)
			return spec, err
		}
	}
	return spec, nil
}

func getHugepageSize(spec *isoSpec) (string) {
	glog.Infof("Inside getHugepageSizefunction")
	if spec != nil {
		if spec.Hugepage != "" {
			return spec.Hugepage
		}
	}
	return "2MB"	
}

func convertHugepageSizetoInt(size string) (int64, error) {
	//TODO: make generic for MB & GB
	if strings.HasSuffix(size, "MB") {
		pagesizeString := strings.TrimSuffix(size, "MB")
		pagesizeInt, err := strconv.Atoi(pagesizeString)
		if err != nil {
			glog.Fatalf("Error converting page size string to integer.Error obtained: %v", err)
			return -1, err
		}
		return int64(pagesizeInt), nil
	}
	return 2, nil
}

func getHugepageTotalRequested(hpNum int64, hpSize int64) (hpTotalRequest int64) {
	//MegaBytestoBytes is a const
	return hpNum * hpSize * MegaBytesToBytes

}

func getHugepageMap(hpsize string, hpRequest string) (hugePageMapValue map[string]string) {
	hugePageMapValue = make(map[string]string)
	hugePageMapValue[hpsize] = hpRequest
	
	return 
}

func getCgroupResourceValue (hpsize string, hpRequest string) (string) {
	cgroupResourceValue := fmt.Sprintf("%s,%s", hpsize, hpRequest)
	glog.Infof("cgroupResourceValue returned to Event: %s", cgroupResourceValue)
	
	return cgroupResourceValue
}

func (h *hugepageIsolator) PreStartPod(podName string, containerName string, pod *v1.Pod, resource *lifecycle.CgroupInfo) ([]*lifecycle.IsolationControl, error) {
	oirHugepages := h.countHugepagesFromOIR(pod)
	glog.Infof("Pod %s requested %d hugepages", pod.Name, oirHugepages)
	 
	if oirHugepages == 0 {
		glog.Infof("Pod %q isn't managed by this isolator", pod.Name)
		return []*lifecycle.IsolationControl{}, nil
	}
	
	recievedSpec, err := getAnnotationFromPod(pod)
	if err != nil{
		glog.Infof("Error: %v could not retrieve annotation from pod spec", err)
		return []*lifecycle.IsolationControl{}, err	
	}
	hugepageSize := getHugepageSize(recievedSpec)
	glog.Infof("Pod %s requested %s size hugepages", pod.Name, hugepageSize)
	
	hugepagesizeInt, err := convertHugepageSizetoInt(hugepageSize)
	if err != nil {
		glog.Infof("Error: %v could not convert hugepagesize to int", err)
		return []*lifecycle.IsolationControl{}, err	
	}
	hpRequestBytes :=  getHugepageTotalRequested(oirHugepages, hugepagesizeInt)
	hpRequestBytesString := strconv.FormatInt(hpRequestBytes, 10)
	cgroupResourceValue := getCgroupResourceValue(hugepageSize, hpRequestBytesString)
	hugePageMapValue := getHugepageMap(hugepageSize, hpRequestBytesString)
	cgroupResource := []*lifecycle.IsolationControl{
		{
			MapValue: hugePageMapValue,
			Value: cgroupResourceValue,
			Kind:  lifecycle.IsolationControl_CGROUP_HUGETLB_LIMIT,
		},
	}
	
	return cgroupResource, nil
}

func (h *hugepageIsolator) PostStopPod(podName string, containerName string, cgroupInfo *lifecycle.CgroupInfo) error {
	//TODO: Add clean up of hugepages on PostStopPod
	return nil
}

func (h *hugepageIsolator) PreStartContainer(podName, containerName string) ([]*lifecycle.IsolationControl, error) {
	glog.Infof("hugepageIsolator[PreStartContainer]:\npodName: %s\ncontainerName: %v", podName, containerName)
	return []*lifecycle.IsolationControl{}, nil
}

func (h *hugepageIsolator) PostStopContainer(podName, containerName string) error {
	glog.Infof("hugepageIsolator[PostStopContainer]:\npodName: %s\ncontainerName: %v", podName, containerName)
	return nil
}

func (h *hugepageIsolator) ShutDown() {
	opaque.RemoveOpaqueResource(h.name)
}

// implementation of Name method in "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/lifecycle".Isolator
func (h *hugepageIsolator) Name() string {
	return h.name
}
