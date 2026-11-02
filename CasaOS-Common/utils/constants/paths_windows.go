//go:build windows

package constants

import "github.com/IceWhaleTech/CasaOS-Common/utils/vfs"

var (
	DefaultConfigPath   = vfs.Real("/etc/casaos")
	DefaultConstantPath = vfs.Real("/usr/share/casaos")
	DefaultDataPath     = vfs.Real("/var/lib/casaos")
	DefaultFilePath     = vfs.Real("/var/lib/casaos/files")
	DefaultLogPath      = vfs.Real("/var/log/casaos")
	DefaultRuntimePath  = vfs.Real("/var/run/casaos")
)
