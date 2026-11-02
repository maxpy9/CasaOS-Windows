// Package vfs maps the POSIX-style paths that CasaOS uses in its API, its config
// files and its web UI (/DATA/Documents, /var/lib/casaos/db, ...) onto the real
// filesystem.
//
// On Linux the mapping is the identity. On Windows the paths are translated, see
// vfs_windows.go for the layout.
package vfs

import "path"

// Join joins path elements of a virtual path, always using forward slashes.
func Join(elem ...string) string {
	return path.Join(elem...)
}
