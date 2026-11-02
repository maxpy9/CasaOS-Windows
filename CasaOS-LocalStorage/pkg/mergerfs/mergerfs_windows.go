//go:build windows

package mergerfs

import (
	"errors"
	"path/filepath"
)

// ErrNotSupported is returned by every operation, since mergerfs does not exist on Windows.
var ErrNotSupported = errors.New("mergerfs is not supported on Windows")

func ControlFile(fspath string) string { return filepath.Join(fspath, ".mergerfs") }

func ListValues(fspath string) (map[string]string, error) { return nil, ErrNotSupported }

func SetSource(fspath string, sources []string) error { return ErrNotSupported }

func GetSource(fspath string) ([]string, error) { return nil, ErrNotSupported }

func AddSource(fspath string, source string) error { return ErrNotSupported }

func RemoveSource(fspath string, source string) error { return ErrNotSupported }

func AddPath(fspath string, path string) error { return ErrNotSupported }

func RemovePath(fspath string, path string) error { return ErrNotSupported }
