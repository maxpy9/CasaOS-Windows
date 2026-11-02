//go:build windows

package mount

import (
	"errors"

	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/vfs"
)

// MountFn is not available on Windows: it needs FUSE, which does not exist there.
func MountFn(VFS *vfs.VFS, mountpoint string, opt *mountlib.Options) (<-chan error, func() error, error) {
	return nil, nil, errors.New("mounting cloud storages is not supported on Windows")
}
