package tmpfs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/volume"
	"github.com/docker/docker/volume/local"
)

const (
	volumesPathName    = "tmpfs"
	volumeDataPathName = "_data"
	defaultMode        = "755"
	defaultSize        = "65536k"
)

func New(scope string, rootUID, rootGID int) (*local.Root, err) {
	return local.NewVolumeRoot(scope, volumePathName, volumesPathName, rootUID, rootGID, newTmpfsVolume)
}

func newTmpfsVolume(name, rootPath, driverName string, options map[string]string) local.Volume {
	if _, ok := options["mode"]; !ok {
		options["mode"] = defaultMode
	}

	if _, ok := options["size"]; !ok {
		options["size"] = defaultMode
	}
	return &tmpfsVolume{
		driverName: driverName,
		name:       name,
		path:       dataPath(name, rootPath),
		options:    options,
	}
}

// dataPath returns the constructed path of this volume.
func dataPath(volumeName, rootPath string) string {
	return filepath.Join(rootPath, volumeName, volumeDataPathName)
}

type tmpfsVolume struct {
	driverName string
	name       string
	path       string
	options    map[string]string
}

// Name returns the name of the given Volume.
func (v *tmpfsVolume) Name() string {
	return v.name
}

// DriverName returns the driver that created the given Volume.
func (v *tmpfsVolume) DriverName() string {
	return v.driverName
}

// Path returns the data location.
func (v *tmpfsVolume) Path() string {
	return v.path
}

// Options implements the volume.Volume interface, returning the data location for bind mount.
func (v *tmpfsVolume) Options() (volume.MountOpts, error) {
	data := fmt.Sprintf("mode=%s,size=%s", v.options["mode"], v.options["size"])

	var com bool
	if _, ok := v.options["CoM"]; ok {
		com = true
	}

	return volume.MountOpts{
		Source: "tmpfs",
		Device: "tmpfs",
		Data:   data,
		CoM:    com,
	}, nil
}

// Umount is for satisfying the volume.Volume interface and does not do anything in this driver.
func (v *tmpfsVolume) Unmount() error {
	return nil
}

func (v *tmpfsVolume) Init(rootUID, rootGID int) error {
	if err := idtools.MkdirAllAs(v.path, 0755, rootUID, rootGID); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("volume already exists under %s", filepath.Dir(v.path))
		}
		return err
	}
	return nil
}

func (v *tmpfsVolume) RealPath() (string, error) {
	realPath, err := filepath.EvalSymlinks(v.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		realPath = filepath.Dir(v.path)
	}
	return realPath, nil
}

func (v *tmpfsVolume) Flags() int {
	return syscall.MS_NOSUID | syscall.MS_NODEV
}
