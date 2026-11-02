//go:build windows

package systemctl

import (
	"errors"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ErrorNotSupported is returned by the operations that have no meaning on Windows,
// where the CasaOS services are started and stopped by the CasaOS launcher.
var ErrorNotSupported = errors.New("not supported on Windows")

type Service struct {
	Name    string
	Running bool
}

// casaosServices are the services of the Windows build, which run as plain processes.
var casaosServices = []string{
	"casaos-gateway",
	"casaos-message-bus",
	"casaos-user-service",
	"casaos",
	"casaos-local-storage",
	"casaos-app-management",
}

// runningProcesses returns the lower-case names, without extension, of the running processes.
func runningProcesses() (map[string]struct{}, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot) // nolint: errcheck

	names := map[string]struct{}{}

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		name := strings.ToLower(windows.UTF16ToString(entry.ExeFile[:]))
		names[strings.TrimSuffix(name, ".exe")] = struct{}{}
	}

	return names, nil
}

func ListServices(pattern string) ([]Service, error) {
	running, err := runningProcesses()
	if err != nil {
		return nil, err
	}

	prefix := strings.TrimSuffix(strings.ToLower(pattern), "*")

	services := make([]Service, 0, len(casaosServices))
	for _, name := range casaosServices {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		_, ok := running[name]
		services = append(services, Service{Name: name + ".service", Running: ok})
	}

	return services, nil
}

func IsServiceEnabled(name string) (bool, error) {
	return IsServiceRunning(name)
}

func IsServiceRunning(name string) (bool, error) {
	running, err := runningProcesses()
	if err != nil {
		return false, err
	}

	_, ok := running[strings.TrimSuffix(strings.ToLower(name), ".service")]

	return ok, nil
}

func EnableService(nameOrPath string) error { return ErrorNotSupported }

func DisableService(name string) error { return ErrorNotSupported }

func StartService(name string) error { return ErrorNotSupported }

func StopService(name string) error { return ErrorNotSupported }

func ReloadDaemon() error { return ErrorNotSupported }
