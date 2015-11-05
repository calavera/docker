package bind

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/volume"
	"github.com/docker/docker/volume/local"
)

const (
	// VolumeDataPathName is the name of the directory where the volume data is stored.
	// It uses a very distintive name to avoid collisions migrating data between Docker versions.
	VolumeDataPathName = "_data"
	// volumesPathName is the name of the host directory where all the bind volumes are stored.
	volumesPathName = "volumes"
)

func New(scope string, rootUID, rootGID int) (*local.Root, error) {
	return local.NewVolumeRoot(scope, volume.DefaultDriverName, volumesPathName, rootUID, rootGID, newBindVolume)
}

func newBindVolume(name, rootPath, driverName string, _ map[string]string) local.Volume {
	return &bindVolume{
		driverName: driverName,
		name:       name,
		path:       dataPath(name, rootPath),
	}
}

// dataPath returns the constructed path of this volume.
func dataPath(volumeName, rootPath string) string {
	return filepath.Join(rootPath, volumeName, VolumeDataPathName)
}

// bindVolume implements the Volume interface from the volume package and
// represents the volumes created by Root.
type bindVolume struct {
	m         sync.Mutex
	usedCount int
	// unique name of the volume
	name string
	// path is the path on the host where the data lives
	path string
	// driverName is the name of the driver that created the volume.
	driverName string
}

// Name returns the name of the given Volume.
func (v *bindVolume) Name() string {
	return v.name
}

// DriverName returns the driver that created the given Volume.
func (v *bindVolume) DriverName() string {
	return v.driverName
}

// Path returns the data location.
func (v *bindVolume) Path() string {
	return v.path
}

// Options implements the bindVolume interface, returning the data location for bind mount.
func (v *bindVolume) Options() (volume.MountOpts, error) {
	return volume.MountOpts{
		Source: v.path,
		Device: "bind",
	}, nil
}

// Umount is for satisfying the bindVolume interface and does not do anything in this driver.
func (v *bindVolume) Unmount() error {
	return nil
}

func (v *bindVolume) Init(rootUID, rootGID int) error {
	if err := idtools.MkdirAllAs(v.path, 0755, rootUID, rootGID); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("volume already exists under %s", filepath.Dir(v.path))
		}
		return err
	}
	return nil
}

func (v *bindVolume) RealPath() (string, error) {
	realPath, err := filepath.EvalSymlinks(v.path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		realPath = filepath.Dir(v.path)
	}
	return realPath, nil
}
