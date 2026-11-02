package service

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
)

// isDockerDesktopStopped tells if an error means that the named pipe of Docker Desktop is not there.
func isDockerDesktopStopped(err error) bool {
	return runtime.GOOS == "windows" && err != nil && (strings.Contains(err.Error(), "/pipe/") || errors.Is(err, context.DeadlineExceeded))
}

// hostVolumePath returns the real path of the source of a volume, and if that
// source is on the filesystem of this computer.
//
// On Windows the containers run in the Linux virtual machine of Docker Desktop.
// Only the folders of CasaOS and the drives of the computer are translated: any
// other path, such as /var/run/docker.sock, belongs to that virtual machine.
func hostVolumePath(source string) (string, bool) {
	if runtime.GOOS != "windows" {
		return source, true
	}

	if filepath.IsAbs(source) {
		return source, true
	}

	s := filepath.ToSlash(source)
	if s == "/DATA" || strings.HasPrefix(s, "/DATA/") || strings.HasPrefix(s, "/media/") || strings.HasPrefix(s, "/var/lib/casaos/") {
		return vfs.Real(s), true
	}

	return source, false
}

// useHostVolumePaths makes the volumes of the app point at the real folders. Docker
// Desktop translates a Windows path to the matching path of its virtual machine.
func (a *ComposeApp) useHostVolumePaths() {
	if runtime.GOOS != "windows" {
		return
	}

	for _, app := range a.Services {
		for j, volume := range app.Volumes {
			if _, ok := a.Volumes[volume.Source]; ok {
				// this is a internal volume, so skip.
				continue
			}

			if path, onHost := hostVolumePath(volume.Source); onHost {
				app.Volumes[j].Source = path
			}
		}
	}
}
