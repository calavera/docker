// Package local provides the default implementation for volumes. It
// is used to mount data volume containers and directories local to
// the host server.
package local

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	derr "github.com/docker/docker/errors"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/utils"
	"github.com/docker/docker/volume"
)

var (
	// ErrNotFound is the typed error returned when the requested volume name can't be found
	ErrNotFound = errors.New("volume not found")
	// volumeNameRegex ensures the name asigned for the volume is valid.
	// This name is used to create the bind directory, so we need to avoid characters that
	// would make the path to escape the root directory.
	volumeNameRegex = utils.RestrictedNamePattern
)

type volumeFactory func(name, volumePath, driverName string, options map[string]string) Volume

type Volume interface {
	// Init initializes the volume for the container root user.
	Init(rootUID, rootGID int) error
	// RealPath is the absolute path or the volume in the host.
	RealPath() (string, error)
}

// NewVolumeRoot instantiates a new Root instance with the provided scope. Scope
// is the base path that the Root instance uses to store its
// volumes. The base path is created here if it does not exist.
func NewVolumeRoot(scope, name, base string, rootUID, rootGID int, factory volumeFactory) (*Root, error) {
	rootDirectory := filepath.Join(scope, base)

	if err := idtools.MkdirAllAs(rootDirectory, 0700, rootUID, rootGID); err != nil {
		return nil, err
	}

	r := &Root{
		scope:   scope,
		name:    name,
		path:    rootDirectory,
		volumes: make(map[string]Volume),
		rootUID: rootUID,
		rootGID: rootGID,
		factory: factory,
	}

	dirs, err := ioutil.ReadDir(rootDirectory)
	if err != nil {
		return nil, err
	}

	opts := map[string]string{}
	for _, d := range dirs {
		name := filepath.Base(d.Name())
		r.volumes[name] = factory(name, r.path, r.Name(), opts)
	}

	return r, nil
}

// Root implements the Driver interface for the volume package and
// manages the creation/removal of volumes. It uses only standard vfs
// commands to create/remove dirs within its provided scope.
type Root struct {
	m       sync.Mutex
	name    string
	scope   string
	path    string
	volumes map[string]Volume
	rootUID int
	rootGID int
	factory volumeFactory
}

// List lists all the volumes
func (r *Root) List() []volume.Volume {
	var ls []volume.Volume
	for _, v := range r.volumes {
		lv, _ := validVolume(v)
		ls = append(ls, lv)
	}
	return ls
}

// Name returns the name of Root, defined in the volume package in the DefaultDriverName constant.
func (r *Root) Name() string {
	return r.name
}

// Create creates a new volume.Volume with the provided name, creating
// the underlying directory tree required for this volume in the
// process.
func (r *Root) Create(name string, options map[string]string) (volume.Volume, error) {
	if err := r.validateName(name); err != nil {
		return nil, err
	}

	r.m.Lock()
	defer r.m.Unlock()

	v, exists := r.volumes[name]
	if exists {
		return validVolume(v)
	}

	v = r.factory(name, r.path, r.Name(), options)
	if err := v.Init(r.rootUID, r.rootGID); err != nil {
		return nil, err
	}

	r.volumes[name] = v
	return validVolume(v)
}

// Remove removes the specified volume and all underlying data. If the
// given volume does not belong to this driver and an error is
// returned. The volume is reference counted, if all references are
// not released then the volume is not removed.
func (r *Root) Remove(v volume.Volume) error {
	r.m.Lock()
	defer r.m.Unlock()

	lv, ok := v.(Volume)
	if !ok {
		return errors.New("unknown volume type")
	}

	realPath, err := lv.RealPath()
	if err != nil {
		return err
	}

	if !r.scopedPath(realPath) {
		return fmt.Errorf("Unable to remove a directory of out the Docker root %s: %s", r.scope, realPath)
	}

	if err := removePath(realPath); err != nil {
		return err
	}

	delete(r.volumes, v.Name())
	return removePath(filepath.Dir(v.Path()))
}

func removePath(path string) error {
	if err := os.RemoveAll(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

// Get looks up the volume for the given name and returns it if found
func (r *Root) Get(name string) (volume.Volume, error) {
	r.m.Lock()
	v, exists := r.volumes[name]
	r.m.Unlock()
	if !exists {
		return nil, ErrNotFound
	}
	return validVolume(v)
}

func (r *Root) validateName(name string) error {
	if !volumeNameRegex.MatchString(name) {
		return derr.ErrorCodeVolumeName.WithArgs(name, utils.RestrictedNameChars)
	}
	return nil
}

func validVolume(v Volume) (volume.Volume, error) {
	vv, ok := v.(volume.Volume)
	if !ok {
		return nil, fmt.Errorf("%v does not implement the Volume interface", v)
	}
	return vv, nil
}
