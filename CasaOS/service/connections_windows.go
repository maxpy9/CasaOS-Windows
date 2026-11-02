//go:build windows

package service

import "errors"

// ErrNotSupportedOnWindows is returned by the features that rely on the Linux kernel or on Samba.
var ErrNotSupportedOnWindows = errors.New("this feature is not supported on Windows")

func (s *connectionsStruct) MountSmaba(username, host, directory, port, mountPoint, password string) error {
	return ErrNotSupportedOnWindows
}

func (s *connectionsStruct) UnmountSmaba(mountPoint string) error {
	return ErrNotSupportedOnWindows
}
