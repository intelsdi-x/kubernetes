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

package qosmanager

import (
	"sync"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/record"
	"k8s.io/kubernetes/pkg/kubelet/lifecycle"
	"k8s.io/kubernetes/pkg/kubelet/qos"
	"k8s.io/kubernetes/pkg/kubelet/server/stats"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
	"k8s.io/kubernetes/pkg/util/clock"
	"k8s.io/kubernetes/pkg/util/wait"
)

// Controller is a sub-controller for monitoring and rectifying various resource
// exhaustion conditions.
type Controller interface {
	Initialize(cxt Context)
	Sync()
	GetNodeConditions() []*api.NodeCondition
	FixNodeCondition(*api.NodeCondition)
}

// Context knows how to perform some kinds of corrections (eviction for now)
// and also get stats for running containers.
type Context struct {
	config           Config
	clock            clock.Clock
	summaryProvider  stats.SummaryProvider
	diskInfoProvider DiskInfoProvider
	imageGC          ImageGC
	activePods       ActivePodsFunc
	evictPod         KillPodFunc
	nodeRef          *api.ObjectReference
	recorder         record.EventRecorder
}

// reportedNodeCondition pairs a condition with the controller it came from.
type reportedNodeCondition struct {
	condition  *api.NodeCondition
	controller Controller
}

// Manager evaluates when an eviction threshold for node stability has been met on the node.
type Manager interface {
	// Start starts the control loop to monitor eviction thresholds at specified interval.
	Start(diskInfoProvider DiskInfoProvider, podFunc ActivePodsFunc, monitoringInterval time.Duration) error

	// GetNodeConditions returns a list of current resource exhaustion conditions.
	GetNodeConditions() []api.NodeConditionType
}

// implements qosmanager.Manager, lifecycle.PodAdmitHandler
type manager struct {
	sync.RWMutex

	// sub-controllers responsible for tracking and fixing resource exhaustion
	controllers []Controller
	//  used to track time
	clock clock.Clock
	// config is how the manager is configured
	config Config
	// the function to invoke to kill a pod
	killPodFunc KillPodFunc
	// the interface that knows how to do image gc
	imageGC ImageGC
	// nodeRef is a reference to the node
	nodeRef *api.ObjectReference
	// used to record events about the node
	recorder record.EventRecorder
	// used to measure usage stats on system
	summaryProvider stats.SummaryProvider
	// most recent node conditions, as reported by the controllers
	nodeConditions map[api.NodeConditionType][]reportedNodeCondition
}

// NewManager returns a QOS Manager and a pod admission handler.
func NewManager(
	summaryProvider stats.SummaryProvider,
	config Config,
	killPodFunc KillPodFunc,
	imageGC ImageGC,
	recorder record.EventRecorder,
	nodeRef *api.ObjectReference,
	clock clock.Clock) (Manager, lifecycle.PodAdmitHandler, error) {
	dmCtrl := diskmem.NewDiskAndMemController()
	m := &manager{
		clock:              clock,
		controllers:        []Controller{dmCtl}, // TODO(CD): Make configurable.
		killPodFunc:        killPodFunc,
		imageGC:            imageGC,
		config:             config,
		recorder:           recorder,
		summaryProvider:    summaryProvider,
		nodeRef:            nodeRef,
		nodeConditions:     map[QosController][]*api.NodeCondition{},
		nodeConditionTypes: map[api.NodeConditionType]struct{}{},
	}
	return m, m, nil
}

// Start starts the control loop to observe and respond to low compute resources.
func (m *manager) Start(diskInfoProvider DiskInfoProvider, podFunc ActivePodsFunc, monitoringInterval time.Duration) error {
	// Initialize all sub-controllers.
	for _, ctrl := range m.controllers {
		ctrl.Initialize(qoscontroller.Context{
			config:           m.config,
			clock:            m.clock,
			summaryProvider:  m.summaryProvider,
			diskInfoProvider: diskInfoProvider,
			imageGC:          m.imageGC,
			activePods:       podFunc,
			evictPod:         m.evictPod,
			nodeRef:          m.nodeRef,
			recorder:         m.recorder,
		})
	}

	// Start the QOS manager monitoring.
	go wait.Until(m.sync, monitoringInterval, wait.NeverStop)
	return nil
}

func (m *manager) Admit(attrs *lifecycle.PodAdmitAttributes) lifecycle.PodAdmitResult {
	m.RLock()
	defer m.RUnlock()
	if len(m.nodeConditions) == 0 {
		return lifecycle.PodAdmitResult{Admit: true}
	}

	// the node has memory pressure, admit if not best-effort
	if hasNodeCondition(m.nodeConditions, api.NodeMemoryPressure) {
		notBestEffort := qos.BestEffort != qos.GetPodQOS(attrs.Pod)
		if notBestEffort {
			return lifecycle.PodAdmitResult{Admit: true}
		}
	}

	// reject pods when under memory pressure (if pod is best effort), or if under disk pressure.
	glog.Warningf("Failed to admit pod %v - %s", format.Pod(attrs.Pod), "node has conditions: %v", m.nodeConditions)
	return lifecycle.PodAdmitResult{
		Admit:   false,
		Reason:  reason,
		Message: message,
	}
}

func (m *manager) GetNodeConditions() []api.NodeCondition {
	// Return a copy of the combined condition list from each sub-controller.
	m.RLock()
	defer m.RUnlock()
	result := []*api.NodeCondition{}
	for _, c := range m.nodeConditions {
		cp = *api.NodeCondition{}
		*cp = *c
		result = append(result, cp)
	}
	return result
}

// getNodeConditions caches current node conditions as reported by
// the sub-controllers.
func (m *manager) getNodeConditions() []*api.NodeCondition {
	m.Lock()
	defer m.Unlock()
	for _, ctrl := range m.controllers {
		for _, cond := range ctrl.GetNodeConditions() {
			m.nodeConditions[cond.Type] = &reportedNodeCondition{cond, ctrl}
		}
	}
}

func (m *manager) sync() {
	for _, ctrl := range m.controllers {
		m.Sync()
	}
	getNodeConditions()

	// TODO(CD): Sort node conditions based on configurable priority.
	//
	// For now, just attempt to fix the "first" condition if we have > 0
	// Because of how go's maps work this is a random selection.
	if len(m.nodeConditions) > 0 {
		for _, reported := range m.nodeConditions {
			reported.controller.FixNodeCondition(reported.condition)
			break // Only try to fix one condition per sync iteration for now.
		}
	}
}

// hasNodeCondition returns true if sub-controllers have reported a true
// condition of the supplied type.
func (m *manager) hasNodeCondition(cType api.NodeConditionType) bool {
	if cond, found := m.nodeConditions[cType]; found {
		return cond.Status == api.ConditionStatusTrue
	}
	return false
}
