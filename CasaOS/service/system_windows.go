//go:build windows

package service

import (
	"strings"

	exec2 "github.com/IceWhaleTech/CasaOS-Common/utils/exec"
	timeutils "github.com/IceWhaleTech/CasaOS-Common/utils/time"
	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
	"github.com/shirou/gopsutil/v3/net"
)

// virtualAdapters are the fragments of the names of the network adapters that do
// not stand for a real network card.
var virtualAdapters = []string{
	"loopback", "vethernet", "hyper-v", "virtualbox", "vmware", "vmnet", "wsl",
	"bluetooth", "tailscale", "zerotier", "wireguard", "openvpn", "tap-", "isatap", "teredo",
	"docker", "local area connection*", "connessione alla rete locale* ",
}

func isVirtualAdapter(name string) bool {
	name = strings.ToLower(name)
	for _, v := range virtualAdapters {
		if strings.Contains(name, v) {
			return true
		}
	}
	return false
}

func hasFlag(i net.InterfaceStat, flag string) bool {
	for _, f := range i.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// platformGetNet returns the names of the network adapters that are up.
func platformGetNet(physics bool) []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return []string{}
	}

	names := []string{}
	for _, i := range interfaces {
		if !hasFlag(i, "up") || hasFlag(i, "loopback") {
			continue
		}
		if physics && (i.HardwareAddr == "" || isVirtualAdapter(i.Name)) {
			continue
		}
		names = append(names, i.Name)
	}
	return names
}

func platformGetNetState(name string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, i := range interfaces {
		if i.Name == name {
			if hasFlag(i, "up") {
				return "up"
			}
			return "down"
		}
	}
	return ""
}

func platformGetTimeZone() string {
	return timeutils.GetSystemTimeZoneName()
}

func platformReboot() error {
	return exec2.Command("shutdown.exe", "/r", "/t", "5", "/c", "CasaOS").Run()
}

func platformShutdown() error {
	return exec2.Command("shutdown.exe", "/s", "/t", "5", "/c", "CasaOS").Run()
}

// platformVirtualDir returns the content of the folders that only exist in the
// paths of CasaOS: the root and the folder that holds the drives of the computer.
func platformVirtualDir(path string) ([]string, bool) {
	switch strings.TrimRight(strings.ReplaceAll(path, `\`, "/"), "/") {
	case "":
		return []string{"DATA", "media"}, true
	case "/media":
		return vfs.Drives(), true
	}
	return nil, false
}

// platformIsProtected tells if a path must not be deleted, moved or renamed.
func platformIsProtected(path string) bool {
	clean := strings.TrimRight(strings.ReplaceAll(vfs.Virtual(path), `\`, "/"), "/")
	if clean == "" || clean == "/DATA" || clean == "/media" {
		return true
	}
	if rest, ok := strings.CutPrefix(clean, "/media/"); ok {
		return !strings.Contains(rest, "/")
	}
	return false
}
