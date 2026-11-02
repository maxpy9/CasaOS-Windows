//go:build !windows

package vfs

// Real returns the real filesystem path of a CasaOS path.
func Real(p string) string { return p }

// Virtual returns the CasaOS path of a real filesystem path.
func Virtual(p string) string { return p }

// SocketPath returns the real path of a Unix socket.
func SocketPath(p string) string { return p }

// Drives returns the letters of the drives exposed under /media.
func Drives() []string { return nil }
