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
	"sort"
	"sync"

	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/kubelet/util/format"
)

type detected struct {
	thresholds    []Threshold
	firstObserved time.Time
	lastObserved  time.Time
}

// DiskAndMemController implements qosmanager.Controller
type DiskAndMemController struct {
	sync.RWMutex
	qosContext                 Context
	nodeConditions             map[api.NodeConditionType]detected
	thresholdsMet              []Threshold
	thresholdsFirstObservedAt  thresholdsObservedAt
	resourceToRankFunc         map[api.ResourceName]rankFunc
	resourceToNodeReclaimFuncs map[api.ResourceName]nodeReclaimFuncs
}

// NewDiskAndMemController returns a new QOS controller for memory and disk.
func NewDiskAndMemController() *DiskAndMemController {
	return &DiskAndMemController{
		nodeConditions:            map[api.NodeConditionType]detected{},
		thresholdsMet:             []Threshold{},
		thresholdsFirstObservedAt: thresholdsObservedAt{},
	}
}

// Initialize sets up local context.
func (c *DiskAndMemController) Initialize(qosContext Context) {
	c.qosContext = qosContext
}

// GetNodeConditions returns currently known node conditions.
func (c *DiskAndMemController) GetNodeConditions() []*api.NodeCondition {
	memCond := &api.NodeCondition{Type: api.NodeMemoryPressure}
	if c.hasNodeCondition(api.NodeMemoryPressure) {
		memCond.Status = api.ConditionTrue
		memCond.Reason = "NodeHasInsufficientMemory"
		memCond.Message = "Node has insufficient memory available"
	} else {
		memCond.Status = api.ConditionFalse
		memCond.Reason = "NodeHasSufficientMemory"
		memCond.Message = "Node has sufficient memory available"
	}

	diskCond := &api.NodeCondition{Type: api.NodeDiskPressure}
	if c.hasNodeCondition(api.NodeDiskPressure) {
		diskCond.Status = api.ConditionTrue
		diskCond.Reason = "NodeHasDiskPressure"
		diskCond.Message = "Node has disk pressure"
	} else {
		diskCond.Status = api.ConditionFalse
		diskCond.Reason = "NodeHasNoDiskPressure"
		diskCond.Message = "Node has no disk pressure"
	}

	return []*api.NodeCondition{memCond, diskCond}
}

// Sync is the main callback that detects qos threshold violations.
func (c *DiskAndMemController) Sync() {
	// if we have nothing to do, just return
	thresholds := c.qosContext.config.Thresholds
	if len(thresholds) == 0 {
		return
	}

	// build the ranking functions (if not yet known)
	// TODO: have a function in cadvisor that lets us know if global housekeeping has completed
	if len(c.resourceToRankFunc) == 0 || len(c.resourceToNodeReclaimFuncs) == 0 {
		// this may error if cadvisor has yet to complete housekeeping, so we will just try again in next pass.
		hasDedicatedImageFs, err := c.qosContext.diskInfoProvider.HasDedicatedImageFs()
		if err != nil {
			return
		}
		c.resourceToRankFunc = buildResourceToRankFunc(hasDedicatedImageFs)
		c.resourceToNodeReclaimFuncs = buildResourceToNodeReclaimFuncs(c.qosContext.imageGC, hasDedicatedImageFs)
	}

	// make observations and get a function to derive pod usage stats relative to those observations.
	observations, statsFunc, err := makeSignalObservations(c.qosContext.summaryProvider)
	if err != nil {
		glog.Errorf("eviction manager: unexpected err: %v", err)
		return
	}

	// find the list of thresholds that are met independent of grace period
	now := m.qosContext.clock.Now()

	// determine the set of thresholds met independent of grace period
	thresholds = thresholdsMet(thresholds, observations, false)

	// determine the set of thresholds previously met that have not yet satisfied the associated min-reclaim
	if len(c.thresholdsMet) > 0 {
		thresholdsNotYetResolved := thresholdsMet(c.thresholdsMet, observations, true)
		thresholds = mergeThresholds(thresholds, thresholdsNotYetResolved)
	}

	// update internal state
	c.Lock()

	// the set of node conditions that are triggered by currently observed thresholds
	c.nodeConditions = nodeConditions(thresholds, now)

	// node conditions report true if it has been observed within the transition period window
	c.nodeConditions = pruneOldConditions(c.nodeConditions, now, c.qosContext.config.PressureTransitionPeriod)

	c.Unlock()
}

func pruneOldDetectedConditions(conds map[api.NodeConditionType]detected, now time.Time, ttl time.Duration) map[api.NodeConditionType]detected {
	pruned := map[api.NodeConditionType]detected{}
	for cType, detected := range conds {
		if now.Sub(detected.lastObserved) < ttl {
			pruned[cType] = detected
		}
	}
	return pruned
}

// FixNodeCondition attempts to fix the supplied condition.
func (c *DiskAndMemController) FixNodeCondition(cond *api.NodeCondition) {
	var thresholds []Threshold
	if thresholds, found := c.nodeConditions[cond.Type]; !found {
		glog.V(3).Infof("disk and memory QOS controller: attempting to fix unknown condition %v", cond.Type)
		return
	}

	// determine the set of thresholds we need to drive eviction behavior (i.e. all grace periods are met)
	c.thresholdsMet = thresholdsMetGracePeriod(thresholdsFirstObservedAt, now)

	// determine the set of resources under starvation for the supplied condition.
	starvedResources := getStarvedResources(thresholds)
	if len(starvedResources) == 0 {
		glog.V(3).Infof("disk and memory QOS controller: no resources are starved")
		return
	}

	// rank the resources to reclaim by eviction priority
	sort.Sort(byEvictionPriority(starvedResources))
	resourceToReclaim := starvedResources[0]
	glog.Warningf("disk and memory QOS controller: attempting to reclaim %v", resourceToReclaim)

	// determine if this is a soft or hard eviction associated with the resource
	softEviction := isSoftEviction(thresholds, resourceToReclaim)

	// record an event about the resources we are now attempting to reclaim via eviction
	c.qosContext.recorder.Eventf(c.qosContext.nodeRef, api.EventTypeWarning, "EvictionThresholdMet", "Attempting to reclaim %s", resourceToReclaim)

	// check if there are node-level resources we can reclaim to reduce pressure before evicting end-user pods.
	if c.reclaimNodeLevelResources(resourceToReclaim, observations) {
		glog.Infof("disk and memory qos controller: able to reduce %v pressure without evicting pods.", resourceToReclaim)
		return
	}

	glog.Infof("disk and memory qos controller: must evict pod(s) to reclaim %v", resourceToReclaim)

	// rank the pods for eviction
	rank, ok := m.resourceToRankFunc[resourceToReclaim]
	if !ok {
		glog.Errorf("disk and memory qos controller: no ranking function for resource %s", resourceToReclaim)
		return
	}

	// the only candidates viable for eviction are those pods that had anything running.
	activePods := podFunc()
	if len(activePods) == 0 {
		glog.Errorf("disk and memory qos controller: eviction thresholds have been met, but no pods are active to evict")
		return
	}

	// rank the running pods for eviction for the specified resource
	rank(activePods, statsFunc)

	glog.Infof("disk and memory qos controller: pods ranked for eviction: %s", format.Pods(activePods))

	// we kill at most a single pod during each eviction interval
	for i := range activePods {
		pod := activePods[i]
		status := api.PodStatus{
			Phase:   api.PodFailed,
			Message: message,
			Reason:  reason,
		}
		// record that we are evicting the pod
		m.recorder.Eventf(pod, api.EventTypeWarning, reason, message)
		gracePeriodOverride := int64(0)
		if softEviction {
			gracePeriodOverride = m.qosContext.config.MaxPodGracePeriodSeconds
		}
		// this is a blocking call and should only return when the pod and its containers are killed.
		err := m.killPodFunc(pod, status, &gracePeriodOverride)
		if err != nil {
			glog.Infof("disk and memory qos controller: pod %s failed to evict %v", format.Pod(pod), err)
			continue
		}
		// success, so we return until the next housekeeping interval
		glog.Infof("disk and memory qos controller: pod %s evicted successfully", format.Pod(pod))
		return
	}
	glog.Infof("disk and memory qos controller: unable to evict any pods from the node")
}

// reclaimNodeLevelResources attempts to reclaim node level resources.  returns true if thresholds were satisfied and no pod eviction is required.
func (c *DiskAndMemController) reclaimNodeLevelResources(resourceToReclaim api.ResourceName, observations signalObservations) bool {
	nodeReclaimFuncs := m.resourceToNodeReclaimFuncs[resourceToReclaim]
	for _, nodeReclaimFunc := range nodeReclaimFuncs {
		// attempt to reclaim the pressured resource.
		reclaimed, err := nodeReclaimFunc()
		if err == nil {
			// update our local observations based on the amount reported to have been reclaimed.
			// note: this is optimistic, other things could have been still consuming the pressured resource in the interim.
			signal := resourceToSignal[resourceToReclaim]
			value, ok := observations[signal]
			if !ok {
				glog.Errorf("eviction manager: unable to find value associated with signal %v", signal)
				continue
			}
			value.available.Add(*reclaimed)

			// evaluate all current thresholds to see if with adjusted observations, we think we have met min reclaim goals
			if len(thresholdsMet(m.thresholdsMet, observations, true)) == 0 {
				return true
			}
		} else {
			glog.Errorf("eviction manager: unexpected error when attempting to reduce %v pressure: %v", resourceToReclaim, err)
		}
	}
	return false
}

func (c *DiskAndMemController) nodeConditionTypes() []api.NodeConditionType {
	cTypes := []api.NodeConditionType{}
	for c := range c.nodeConditions {
		cTypes = append(cTypes, c)
	}
	return cTypes
}

func (c *DiskAndMemController) hasNodeCondition(cType api.NodeConditionType) bool {
	for c := range c.nodeConditions {
		if c == cType {
			return true
		}
	}
	return false
}
