// +build linux freebsd

// Package local provides the default implementation for volumes. It
// is used to mount data volume containers and directories local to
// the host server.
package local

import (
	"path/filepath"
	"strings"
)

var oldVfsDir = filepath.Join("vfs", "dir")

// scopedPath verifies that the path where the volume is located
// is under Docker's root and the valid local paths.
func (r *Root) scopedPath(realPath string) bool {
	// Volumes path for Docker version >= 1.7
	if strings.HasPrefix(realPath, r.path) && realPath != r.path {
		return true
	}

	// Volumes path for Docker version < 1.7
	rootBase := filepath.Dir(r.path)
	if strings.HasPrefix(realPath, filepath.Join(rootBase, oldVfsDir)) {
		return true
	}

	return false
}
