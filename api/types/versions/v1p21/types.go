// Package v1p21 provides specific API types for the API version 1, patch 21.
package v1p21

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/runconfig"
)

// NetworkSettings is a backward compatible struct for APIs prior to 1.22
type NetworkSettings struct {
	*types.NetworkSettingsBase
	EndpointID          string
	Gateway             string
	GlobalIPv6Address   string
	GlobalIPv6PrefixLen int
	IPAddress           string
	IPPrefixLen         int
	IPv6Gateway         string
	MacAddress          string
}

// ContainerJSON is a backcompatibility struct for the API 1.21
type ContainerJSON struct {
	*types.ContainerJSONBase
	Mounts          []types.MountPoint
	Config          *runconfig.Config
	NetworkSettings *NetworkSettings
}
