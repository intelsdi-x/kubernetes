package hook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	dockertypes "github.com/docker/engine-api/types"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api"
)

// ContainerEventHooks implements container lifecycle callbacks to be invoked by
// the runtime.
type ContainerEventHooks interface {
	Name() string
	PreCreate(pod *api.Pod, container *api.Container, config *dockertypes.ContainerCreateConfig)
	PreStart(pod *api.Pod, container *api.Container, config *dockertypes.ContainerCreateConfig)
	PostStart(pod *api.Pod, container *api.Container, config *dockertypes.ContainerCreateConfig)
	// PostStop(pod *api.Pod, container *api.Container)
}

type eventType string

const (
	preCreate eventType = "PRE_CREATE"
	preStart  eventType = "PRE_START"
	postStart eventType = "POST_START"
	// postStop  eventType = "POST_STOP"
)

type eventInfo struct {
	Name eventType `json:"name"`
	eventContext
}

type eventContext struct {
	Pod       *api.Pod                           `json:"pod"`
	Container *api.Container                     `json:"container"`
	Config    *dockertypes.ContainerCreateConfig `json:"config"`
}

// NewScriptContainerEventHooks returns a set of hooks that delegate to the
// supplied script.
func NewScriptContainerEventHooks(scriptPath string) ContainerEventHooks {
	return scriptContainerEventHook{scriptPath}
}

// implements hook.ContainerEventHooks by executing an external program.
// For each event callback, the program is invoked synchronously. The event
// details are passed through standard input. The called program is expected
// to make changes to the container configuration and emit a JSON-formatted
// response on standard output.
type scriptContainerEventHook struct{ scriptPath string }

// The Kubelet feeds the script program on standard input a JSON-encoded event
// that looks like:
//
// {
//   "name": "PRE_CREATE",
//   "eventContext": {
//     "pod": { ... },
//     "container": { ... },
//     "config": { ... }
//   }
// }
//
// The hook script must emit (only) the potentially modified eventContext part,
// also JSON-encoded, to standard output. Note that modified structures only
// have meaning in response to the PRE_CREATE event.
//
// {
//   "pod": { ... },
//   "container": { ... },
//   "config": { ... }
// }
//
// Finally, bytes the hook program outputs to standard error will be included
// in the Kubelet logs.
func (h scriptContainerEventHook) callWithEvent(
	eType eventType,
	pod *api.Pod,
	container *api.Container,
	config *dockertypes.ContainerCreateConfig) (*eventContext, error) {

	e := eventInfo{eType, eventContext{pod, container, config}}
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(h.scriptPath)
	cmd.Stdin = strings.NewReader(string(data))
	cmd.Env = []string{} // Clean subprocess environment.

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	cmd.Start()

	// Read standard error and log it, in a separate goroutine
	// to avoid blocking writes to the stdout pipe.
	go func() {
		eout, err := ioutil.ReadAll(stderr)
		if err != nil {
			glog.Errorf("Unable to read stderr from hook script [%s]: %s", h.scriptPath, err.Error())
		}
		glog.Infof("Container event hook script [stderr]:\n%s", string(eout))
	}()

	// Parse standard out as eventContext JSON.
	var resultContext *eventContext
	if err := json.NewDecoder(stdout).Decode(&resultContext); err != nil {
		return nil, err
	}

	// Log returned context.
	resultJSON, err := json.Marshal(resultContext)
	if err != nil {
		return nil, err
	}
	glog.Infof("Container event hook script (%s) returned pod, container and config:\n%s", eType, resultJSON)

	return resultContext, nil
}

func (h scriptContainerEventHook) Name() string {
	return fmt.Sprintf("Script container event hook [%s]", h.scriptPath)
}

func (h scriptContainerEventHook) PreCreate(pod *api.Pod, container *api.Container, config *dockertypes.ContainerCreateConfig) {
	ctx, err := h.callWithEvent(preCreate, pod, container, config)
	if err != nil {
		glog.Errorf("Failed to execute hook script [%s] for event %s: %s", h.scriptPath, preCreate, err.Error())
		return
	}

	// Modify referenced config.
	*config = *ctx.Config
}

func (h scriptContainerEventHook) PreStart(pod *api.Pod, container *api.Container, config *dockertypes.ContainerCreateConfig) {
	_, err := h.callWithEvent(preStart, pod, container, config)
	if err != nil {
		glog.Errorf("Failed to execute hook script [%s] for event %s: %s", h.scriptPath, preStart, err.Error())
		return
	}
}

func (h scriptContainerEventHook) PostStart(pod *api.Pod, container *api.Container, config *dockertypes.ContainerCreateConfig) {
	_, err := h.callWithEvent(postStart, pod, container, config)
	if err != nil {
		glog.Errorf("Failed to execute hook script [%s] for event %s: %s", h.scriptPath, postStart, err.Error())
		return
	}
}
