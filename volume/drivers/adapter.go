package volumedrivers

import "github.com/docker/docker/volume"

type volumeDriverAdapter struct {
	name  string
	proxy *volumeDriverProxy
}

func (a *volumeDriverAdapter) Name() string {
	return a.name
}

func (a *volumeDriverAdapter) Create(name string, opts map[string]string) (volume.Volume, error) {
	err := a.proxy.Create(name, opts)
	if err != nil {
		return nil, err
	}
	return &volumeAdapter{
		proxy:      a.proxy,
		name:       name,
		driverName: a.name}, nil
}

func (a *volumeDriverAdapter) Remove(v volume.Volume) error {
	return a.proxy.Remove(v.Name())
}

type volumeAdapter struct {
	proxy      *volumeDriverProxy
	name       string
	driverName string
	eMount     string // ephemeral host volume path
}

type proxyVolume struct {
	Name       string
	Mountpoint string
}

func (a *volumeAdapter) Name() string {
	return a.name
}

func (a *volumeAdapter) DriverName() string {
	return a.driverName
}

func (a *volumeAdapter) Path() string {
	if len(a.eMount) > 0 {
		return a.eMount
	}
	m, _ := a.proxy.Path(a.name)
	return m
}

func (a *volumeAdapter) Options() (volume.MountOpts, error) {
	var err error
	a.eMount, err = a.proxy.Mount(a.name)
	if err != nil {
		return volume.MountOpts{}, err
	}
	return volume.MountOpts{
		Source: a.eMount,
		Device: "bind",
	}, nil
}

func (a *volumeAdapter) Unmount() error {
	return a.proxy.Unmount(a.name)
}
