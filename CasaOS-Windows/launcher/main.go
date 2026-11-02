// Command CasaOS is the launcher of the Windows build of CasaOS: a window with a
// button that starts and stops the server. Nothing runs when the server is stopped:
// no Windows service is installed and nothing starts with Windows.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
	"golang.org/x/sys/windows"
)

const mutexName = "Local\\CasaOSWindowsLauncher"

func init() {
	// the window and its message loop must stay on the same thread
	runtime.LockOSThread()
}

func setupLog(dir string) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "launcher.log")
	if fi, err := os.Stat(path); err == nil && fi.Size() > 1<<20 {
		os.Remove(path) // nolint: errcheck
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		log.SetOutput(f)
	}
}

// signalRunningInstance sends a message to the window of the launcher that is already
// open, and tells if there is one.
func signalRunningInstance(message uintptr) bool {
	hwnd, _, _ := procFindWindowW.Call(ptr(windowClass), 0)
	if hwnd == 0 {
		return false
	}
	procPostMessageW.Call(hwnd, message, 0, 0) // nolint: errcheck
	return true
}

// waitForExit waits for the launcher that is open to end.
func waitForExit(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if hwnd, _, _ := procFindWindowW.Call(ptr(windowClass), 0); hwnd == 0 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// prepareServices is run at the end of the installation. A security program analyzes an
// unknown program the first time it runs, which can hold it for a long time: better to
// have that happen now than when the user presses the button for the first time.
func prepareServices(paths Paths) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			cmd := exec.CommandContext(ctx, filepath.Join(paths.Bin, name+".exe"), "-v")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
			_ = cmd.Run()
		}(svc.name)
	}
	wg.Wait()
}

func main() {
	quit := flag.Bool("quit", false, "stop the server and close the launcher that is open")
	prepare := flag.Bool("prepare", false, "run every service once, so that a security program can analyze them early")
	autostart := flag.Bool("autostart", false, "start the server as soon as the window opens")
	hidden := flag.Bool("hidden", false, "open in the notification area, without the window")
	screenshot := flag.String("screenshot", "", "save a picture of the window to this file, for testing")
	flag.Parse()

	if *quit {
		if signalRunningInstance(wmAppQuit) {
			waitForExit(2 * time.Minute)
		}
		return
	}

	if *prepare {
		prepareServices(newPaths())
		return
	}

	// DPI awareness must be set before any window is created
	procSetProcessDpiContext.Call(dpiAwarenessPerMonitorV2) // nolint: errcheck

	mutex, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(utf16(mutexName))))
	if mutex == 0 {
		log.Println("mutex:", err)
	} else if err == windows.ERROR_ALREADY_EXISTS {
		signalRunningInstance(wmAppShow)
		return
	}

	paths := newPaths()
	setupLog(vfs.Real("/var/log/casaos"))
	log.Println("launcher started")

	var server *Server
	server = NewServer(paths, func() {
		if ui != nil && ui.hwnd != 0 {
			procPostMessageW.Call(ui.hwnd, wmAppState, 0, 0) // nolint: errcheck
		}
	})

	runUI(server, paths, *autostart, *hidden, *screenshot)

	// the window is gone: nothing must stay behind
	server.StopNow()
	log.Println("launcher ended")
}
