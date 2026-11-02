//go:build windows

package vfs

import (
	"fmt"
	"hash/crc32"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// Layout on Windows:
//
//	/DATA/...      -> %CASAOS_DATA%\...  (default %USERPROFILE%\CasaOS)
//	/media/C/...   -> C:\...             (every drive of the computer)
//	anything else  -> %CASAOS_ROOT%\...  (default %LOCALAPPDATA%\CasaOS\root)
var (
	root = rootDir()
	data = dataDir()
)

func rootDir() string {
	if v := os.Getenv("CASAOS_ROOT"); v != "" {
		return filepath.Clean(v)
	}
	if v := os.Getenv("LOCALAPPDATA"); v != "" {
		return filepath.Join(v, "CasaOS", "root")
	}
	return filepath.Join(os.TempDir(), "CasaOS", "root")
}

func dataDir() string {
	if v := os.Getenv("CASAOS_DATA"); v != "" {
		return filepath.Clean(v)
	}
	if v, err := os.UserHomeDir(); err == nil {
		return filepath.Join(v, "CasaOS")
	}
	return filepath.Join(root, "DATA")
}

func isDriveLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func hasDrive(p string) bool {
	return len(p) >= 2 && p[1] == ':' && isDriveLetter(p[0])
}

// Real returns the real filesystem path of a CasaOS path. A path that already is
// a real path and a relative path are returned as they are, so the function is
// safe to apply more than once.
func Real(p string) string {
	if p == "" {
		return p
	}
	if hasDrive(p) || strings.HasPrefix(p, `\\`) {
		return filepath.Clean(p)
	}
	if p[0] != '/' && p[0] != '\\' {
		return p
	}

	s := path.Clean(strings.ReplaceAll(p, `\`, "/"))

	if s == "/DATA" {
		return data
	}
	if rest, ok := strings.CutPrefix(s, "/DATA/"); ok {
		return filepath.Join(data, filepath.FromSlash(rest))
	}
	// besides /media/<letter>, these are the paths of the drives inside the virtual machine of Docker Desktop
	for _, prefix := range []string{"/media/", "/run/desktop/mnt/host/", "/host_mnt/"} {
		if rest, ok := strings.CutPrefix(s, prefix); ok {
			letter, tail, _ := strings.Cut(rest, "/")
			if len(letter) == 1 && isDriveLetter(letter[0]) {
				return strings.ToUpper(letter) + `:\` + filepath.FromSlash(tail)
			}
		}
	}
	return filepath.Join(root, filepath.FromSlash(s))
}

// SocketPath returns the real path of a Unix socket. The path of a socket cannot be
// longer than 108 bytes, which a long user name can exceed.
func SocketPath(p string) string {
	real := Real(p)
	if len(real) < 100 {
		return real
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("casaos-%08x.sock", crc32.ChecksumIEEE([]byte(real))))
}

// cutBase returns the part of p below base, compared case-insensitively as the
// filesystem does.
func cutBase(p, base string) (string, bool) {
	if !strings.EqualFold(p, base) && !(len(p) > len(base) && strings.EqualFold(p[:len(base)], base) && p[len(base)] == filepath.Separator) {
		return "", false
	}
	return filepath.ToSlash(p[len(base):]), true
}

// Virtual returns the CasaOS path of a real filesystem path. A path that already
// is a CasaOS path only gets its separators normalized.
func Virtual(p string) string {
	if p == "" {
		return p
	}
	if !hasDrive(p) {
		return strings.ReplaceAll(p, `\`, "/")
	}
	c := filepath.Clean(p)
	if rest, ok := cutBase(c, data); ok {
		return "/DATA" + rest
	}
	if rest, ok := cutBase(c, root); ok {
		if rest == "" {
			return "/"
		}
		return rest
	}
	return strings.TrimRight("/media/"+strings.ToUpper(c[:1])+filepath.ToSlash(c[2:]), "/")
}

// Drives returns the letters of the drives exposed under /media. Optical and
// unknown drives are left out, since they have nothing to browse or fail on access.
func Drives() []string {
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return nil
	}
	drives := []string{}
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		letter := string(rune('A' + i))
		name, err := windows.UTF16PtrFromString(letter + `:\`)
		if err != nil {
			continue
		}
		switch windows.GetDriveType(name) {
		case windows.DRIVE_FIXED, windows.DRIVE_REMOVABLE, windows.DRIVE_REMOTE, windows.DRIVE_RAMDISK:
			drives = append(drives, letter)
		}
	}
	return drives
}
