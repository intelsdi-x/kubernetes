package hugepages

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/golang/glog"
	"github.com/google/cadvisor/machine"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubelet/cadvisor"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/util/strings"
	"k8s.io/kubernetes/pkg/volume"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

// http://issue.k8s.io/2630
const perm os.FileMode = 0777

// ProbeVolumePlugins is the primary entrypoint for volume plugins.
func ProbeVolumePlugins(cadvisorInterface cadvisor.Interface) []volume.VolumePlugin {
	return []volume.VolumePlugin{
		&hugePagesPlugin{nil, cadvisorInterface},
	}
}

type hugePagesPlugin struct {
	host              volume.VolumeHost
	cadvisorInterface cadvisor.Interface
}

var _ volume.VolumePlugin = &hugePagesPlugin{}

const (
	hugePagesPluginName = "kubernetes.io/hugepages"
)

func getPath(uid types.UID, volName string, host volume.VolumeHost) string {
	return host.GetPodVolumeDir(uid, strings.EscapeQualifiedNameForDisk(hugePagesPluginName), volName)
}

func (plugin *hugePagesPlugin) Init(host volume.VolumeHost) error {
	plugin.host = host
	return nil
}

func (plugin *hugePagesPlugin) GetPluginName() string {
	return hugePagesPluginName
}

func (plugin *hugePagesPlugin) GetVolumeName(spec *volume.Spec) (string, error) {
	volumeSource, _ := getVolumeSource(spec)
	if volumeSource == nil {
		return "", fmt.Errorf("Spec does not reference a HugePages volume type")
	}

	// Return user defined volume name, since this is an ephemeral volume type
	return spec.Name(), nil
}

func toBytes(value string) (int, error) {
	size, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return -1, err
	}
	kind := value[len(value)-1:]
	switch kind {
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	case "G":
		return size * 1024 * 1024 * 1024, nil
	}
	return -1, fmt.Errorf("cannot decode kind")
}

func (plugin *hugePagesPlugin) CanSupport(spec *volume.Spec) bool {
	if plugin.cadvisorInterface == nil {
		glog.Warning("cadvisor hasn't been passed")
		return false
	}

	hostInfo, err := plugin.cadvisorInterface.MachineInfo()
	if err != nil {
		glog.Warning("cannot gather machine info")
		return false
	}

	if err = machine.UpdateHugePages(hostInfo); err != nil {
		glog.Warningf("Cannot update hugepages size")
		return false
	}

	if hostInfo.HugePageSize <= 0 {
		glog.Warning("Wrong hugepage size")
		return false
	}

	if hostInfo.HugePagesTotal <= 0 {
		glog.Warning("No hugepages available")
		return false
	}

	availableHugePages := (int)(hostInfo.HugePagesTotal - hostInfo.HugePagesRsvd)
	requestedTotalHugePagesSize, err := toBytes(spec.Volume.HugePages.MinSize)
	if err != nil {
		return false
	}

	requestedPageSize, err := toBytes(spec.Volume.HugePages.PageSize)
	if err != nil {
		return false
	}
	glog.Infof("############################################################# HPR: %v, HPT: %v", hostInfo.HugePagesRsvd, hostInfo.HugePagesTotal)
	glog.Infof("############################################################# AHP: %v, RPS: %v, RTHPS: %v", availableHugePages, requestedPageSize, requestedTotalHugePagesSize)

	if availableHugePages-(requestedTotalHugePagesSize/requestedPageSize) < 0 {
		glog.Warning("Cannot reserve this amount of hugepages on this host")
		return false
	}

	if spec.Volume != nil && spec.Volume.HugePages != nil {
		return true
	}
	return false
}

func (plugin *hugePagesPlugin) RequiresRemount() bool {
	return false
}

func (plugin *hugePagesPlugin) SupportsMountOption() bool {
	return false
}

func (plugin *hugePagesPlugin) SupportsBulkVolumeVerification() bool {
	return false
}

func (plugin *hugePagesPlugin) NewMounter(spec *volume.Spec, pod *v1.Pod, opts volume.VolumeOptions) (volume.Mounter, error) {
	return &hugePages{
		pod:      pod,
		volName:  spec.Name(),
		mounter:  plugin.host.GetMounter(),
		pageSize: spec.Volume.HugePages.PageSize,
		size:     spec.Volume.HugePages.MaxSize,
		minSize:  spec.Volume.HugePages.MinSize,
		plugin:   plugin,
	}, nil
}

func (plugin *hugePagesPlugin) NewUnmounter(volName string, podUID types.UID) (volume.Unmounter, error) {
	return &hugePages{
		pod:     &v1.Pod{ObjectMeta: metav1.ObjectMeta{UID: podUID}},
		volName: volName,
		mounter: plugin.host.GetMounter(),
		plugin:  plugin,
	}, nil
}

func (plugin *hugePagesPlugin) ConstructVolumeSpec(volName, mountPath string) (*volume.Spec, error) {
	hugePagesVolume := &v1.Volume{
		Name: volName,
		VolumeSource: v1.VolumeSource{
			HugePages: &v1.HugePagesVolumeSource{},
		},
	}

	return volume.NewSpecFromVolume(hugePagesVolume), nil
}

// HugePages are https://www.kernel.org/doc/Documentation/vm/hugetlbpage.txt
// This struct implements  "k8s.io/pkg/volume.Mounter" and "k8s.io/pkg/volme.UnMounter" interface.Mounter" and "k8s.io/pkg/volme.UnMounter" interfaces
type hugePages struct {
	pod      *v1.Pod
	volName  string
	mounter  mount.Interface
	plugin   *hugePagesPlugin
	pageSize string
	minSize  string
	size     string
	volume.MetricsNil
}

func (hp *hugePages) GetAttributes() volume.Attributes {
	return volume.Attributes{
		ReadOnly:        false,
		Managed:         false,
		SupportsSELinux: false,
	}
}

// Checks prior to mount operations to verify that the required components (binaries, etc.)
// to mount the volume are available on the underlying node.
// If not, it returns an error
func (hp *hugePages) CanMount() error {
	return nil
}

// SetUp creates new directory.
func (hp *hugePages) SetUp(fsGroup *int64) error {
	return hp.SetUpAt(hp.GetPath(), fsGroup)
}

// SetUpAt creates new directory.
func (hp *hugePages) SetUpAt(dir string, fsGroup *int64) error {
	notMnt, err := hp.mounter.IsLikelyNotMountPoint(dir)
	// Getting an os.IsNotExist err from is a contingency; the directory
	// may not exist yet, in which case, setup should run.
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if volumeutil.IsReady(hp.getMetaDir()) && !notMnt {
		return nil
	}

	// If the plugin readiness file is present for this volume, and the
	// storage medium is the default, then the volume is ready.  If the
	// medium is memory, and a mountpoint is present, then the volume is
	// ready.

	err = hp.setupHugePages(dir)

	volume.SetVolumeOwnership(hp, fsGroup)

	if err == nil {
		volumeutil.SetReady(hp.getMetaDir())
	}

	return err
}

// setupTmpfs creates a tmpfs mount at the specified directory with the
// specified SELinux context.
func (hp *hugePages) setupHugePages(dir string) error {
	if hp.mounter == nil {
		return fmt.Errorf("mounter is nil")
	}
	if err := hp.setupDir(dir); err != nil {
		return err
	}

	options := []string{
		fmt.Sprintf("size=%s", hp.size),
		fmt.Sprintf("pagesize=%s", hp.pageSize),
		fmt.Sprintf("min_size=%s", hp.minSize),
	}

	return hp.mounter.Mount("nodev", dir, "hugetlbfs", options)
}

// setupDir creates the directory with the specified SELinux context and
// the default permissions specified by the perm constant.
func (hp *hugePages) setupDir(dir string) error {
	// Create the directory if it doesn't already exist.
	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}

	// stat the directory to read permission bits
	fileinfo, err := os.Lstat(dir)
	if err != nil {
		return err
	}

	if fileinfo.Mode().Perm() != perm.Perm() {
		// If the permissions on the created directory are wrong, the
		// kubelet is probably running with a umask set.  In order to
		// avoid clearing the umask for the entire process or locking
		// the thread, clearing the umask, creating the dir, restoring
		// the umask, and unlocking the thread, we do a chmod to set
		// the specific bits we need.
		err := os.Chmod(dir, perm)
		if err != nil {
			return err
		}

		fileinfo, err = os.Lstat(dir)
		if err != nil {
			return err
		}

		if fileinfo.Mode().Perm() != perm.Perm() {
			glog.Errorf("Expected directory %q permissions to be: %s; got: %s", dir, perm.Perm(), fileinfo.Mode().Perm())
		}
	}

	return nil
}

func (hp *hugePages) GetPath() string {
	return getPath(hp.pod.UID, hp.volName, hp.plugin.host)
}

// TearDown simply discards everything in the directory.
func (hp *hugePages) TearDown() error {
	return hp.TearDownAt(hp.GetPath())
}

// TearDownAt simply discards everything in the directory.
func (hp *hugePages) TearDownAt(dir string) error {
	if pathExists, pathErr := volumeutil.PathExists(dir); pathErr != nil {
		return fmt.Errorf("Error checking if path exists: %v", pathErr)
	} else if !pathExists {
		glog.Warningf("Warning: Unmount skipped because path does not exist: %v", dir)
		return nil
	}

	return hp.teardownHugePages(dir)
}

func (hp *hugePages) teardownHugePages(dir string) error {
	if hp.mounter == nil {
		return fmt.Errorf("mounter is nil")
	}
	if err := hp.mounter.Unmount(dir); err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}

func (hp *hugePages) getMetaDir() string {
	return path.Join(hp.plugin.host.GetPodPluginDir(hp.pod.UID, strings.EscapeQualifiedNameForDisk(hugePagesPluginName)), hp.volName)
}

func getVolumeSource(spec *volume.Spec) (*v1.HugePagesVolumeSource, bool) {
	var readOnly bool
	var volumeSource *v1.HugePagesVolumeSource

	if spec.Volume != nil && spec.Volume.HugePages != nil {
		volumeSource = spec.Volume.HugePages
		readOnly = spec.ReadOnly
	}
	return volumeSource, readOnly
}
