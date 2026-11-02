//go:build windows

package wrapper

import (
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
	"github.com/moby/sys/mountinfo"
	"golang.org/x/sys/windows"
)

// This interface is used to mock the mountinfo package in unit tests.
type MountInfoWrapper interface {
	GetMounts(f mountinfo.FilterFunc) ([]*mountinfo.Info, error)
}

type MountInfo struct{}

func NewMountInfo() MountInfoWrapper {
	return &MountInfo{}
}

func fsType(letter string) string {
	root, err := windows.UTF16PtrFromString(letter + `:\`)
	if err != nil {
		return ""
	}

	buffer := make([]uint16, windows.MAX_PATH+1)
	if err := windows.GetVolumeInformation(root, nil, 0, nil, nil, nil, &buffer[0], uint32(len(buffer))); err != nil {
		return ""
	}

	return strings.ToLower(windows.UTF16ToString(buffer))
}

// GetMounts returns the drives of the computer, that are shown as /media/<letter>.
func (m *MountInfo) GetMounts(f mountinfo.FilterFunc) ([]*mountinfo.Info, error) {
	mounts := []*mountinfo.Info{}

	for i, letter := range vfs.Drives() {
		info := &mountinfo.Info{
			ID:         i + 1,
			Mountpoint: "/media/" + letter,
			Source:     letter + ":",
			FSType:     fsType(letter),
			Options:    "rw",
		}

		skip, stop := false, false
		if f != nil {
			skip, stop = f(info)
		}
		if !skip {
			mounts = append(mounts, info)
		}
		if stop {
			break
		}
	}

	return mounts, nil
}
