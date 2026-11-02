package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	winio "github.com/Microsoft/go-winio"
)

// DockerState tells what can be said about Docker Desktop.
type DockerState int

const (
	DockerNotInstalled DockerState = iota
	DockerStopped
	DockerStarting
	DockerRunning
)

const (
	dockerStartTimeout = 3 * time.Minute
	dockerCmdTimeout   = 90 * time.Second
	dockerPsTimeout    = 10 * time.Second
)

// Docker makes the apps of CasaOS follow the server: they stop with it and they are
// back when it starts again. Docker Desktop itself is only stopped if it is the
// launcher that started it, since the user may need it for something else.
type Docker struct {
	paths     Paths
	startedBy bool

	mu       sync.Mutex
	starting bool
}

func NewDocker(paths Paths) *Docker {
	return &Docker{paths: paths}
}

func dockerDesktopExe() string {
	for _, env := range []string{"ProgramFiles", "ProgramW6432", "LOCALAPPDATA"} {
		base := os.Getenv(env)
		if base == "" {
			continue
		}
		for _, rel := range []string{`Docker\Docker\Docker Desktop.exe`, `Programs\Docker\Docker\Docker Desktop.exe`} {
			p := filepath.Join(base, rel)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

func dockerCLI() string {
	if p, err := exec.LookPath("docker"); err == nil {
		return p
	}
	if base := os.Getenv("ProgramFiles"); base != "" {
		p := filepath.Join(base, `Docker\Docker\resources\bin\docker.exe`)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// engineRunning tells if the engine really answers. The named pipe alone is not enough:
// while Docker Desktop is starting the pipe exists, but a request on it just waits.
func engineRunning() bool {
	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, `\\.\pipe\docker_engine`)
			},
		},
	}

	resp, err := client.Get("http://docker/_ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// State returns the state of Docker Desktop.
func (d *Docker) State() DockerState {
	if engineRunning() {
		return DockerRunning
	}

	d.mu.Lock()
	starting := d.starting
	d.mu.Unlock()
	if starting {
		return DockerStarting
	}

	if dockerDesktopExe() == "" && dockerCLI() == "" {
		return DockerNotInstalled
	}
	return DockerStopped
}

func (d *Docker) run(timeout time.Duration, args ...string) (string, error) {
	cli := dockerCLI()
	if cli == "" {
		return "", exec.ErrNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cli, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// StartDesktop opens Docker Desktop, if it is not running yet.
func (d *Docker) StartDesktop() {
	if engineRunning() {
		return
	}

	exe := dockerDesktopExe()
	if exe == "" {
		return
	}

	d.mu.Lock()
	if d.starting {
		d.mu.Unlock()
		return
	}
	d.starting = true
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.starting = false
		d.mu.Unlock()
	}()

	if err := exec.Command(exe).Start(); err != nil {
		log.Println("docker desktop:", err)
		return
	}
	d.startedBy = true

	deadline := time.Now().Add(dockerStartTimeout)
	for time.Now().Before(deadline) {
		if engineRunning() {
			return
		}
		time.Sleep(2 * time.Second)
	}
	log.Println("docker desktop: the engine did not start in time")
}

// OnServerStarted starts Docker Desktop if the user wants it, and the apps that were
// running the last time the server was stopped.
func (d *Docker) OnServerStarted() {
	if loadSettings(d.paths).AutoDocker {
		d.StartDesktop()
	}

	if !engineRunning() {
		return
	}

	ids := d.loadStopped()
	if len(ids) == 0 {
		return
	}

	log.Println("starting the apps stopped with the server:", ids)
	if out, err := d.run(dockerCmdTimeout, append([]string{"start"}, ids...)...); err != nil {
		log.Println("docker start:", err, out)
	}
	os.Remove(d.stoppedFile()) // nolint: errcheck
}

// OnServerStopping stops the apps of CasaOS, so that nothing is left running. It never
// touches the containers that do not belong to CasaOS.
func (d *Docker) OnServerStopping() {
	if !engineRunning() {
		return
	}

	ids := d.runningApps()
	if len(ids) > 0 {
		log.Println("stopping the apps:", ids)
		d.saveStopped(ids)
		if out, err := d.run(dockerCmdTimeout, append([]string{"stop", "--time", "10"}, ids...)...); err != nil {
			log.Println("docker stop:", err, out)
		}
	}

	if d.startedBy {
		d.startedBy = false
		if out, err := d.run(dockerCmdTimeout, "desktop", "stop"); err != nil {
			// an older Docker Desktop has no such command: it just stays open
			log.Println("docker desktop stop:", err, out)
		}
	}
}

// runningApps returns the running containers of the apps installed by CasaOS, which
// are the compose projects with their working folder among the apps of CasaOS.
func (d *Docker) runningApps() []string {
	output, err := d.run(dockerPsTimeout, "ps", "--format", `{{.ID}}|{{.Label "com.docker.compose.project.working_dir"}}`)
	if err != nil {
		log.Println("docker ps:", err, output)
		return nil
	}

	apps := strings.ToLower(filepath.Join(d.paths.Lib, "apps")) + string(filepath.Separator)

	ids := []string{}
	for _, line := range strings.Split(output, "\n") {
		id, dir, ok := strings.Cut(strings.TrimSpace(line), "|")
		if ok && strings.HasPrefix(strings.ToLower(dir), apps) {
			ids = append(ids, id)
		}
	}
	return ids
}

func (d *Docker) stoppedFile() string {
	return filepath.Join(d.paths.Lib, "apps-stopped-with-server.json")
}

func (d *Docker) saveStopped(ids []string) {
	data, err := json.Marshal(ids)
	if err == nil {
		err = os.WriteFile(d.stoppedFile(), data, 0o644)
	}
	if err != nil {
		log.Println("saving the stopped apps:", err)
	}
}

func (d *Docker) loadStopped() []string {
	data, err := os.ReadFile(d.stoppedFile())
	if err != nil {
		return nil
	}
	ids := []string{}
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil
	}
	return ids
}

// Settings are the choices of the user in the window.
type Settings struct {
	AutoDocker bool `json:"auto_docker"`
	TrayHinted bool `json:"tray_hinted"`
}

var settingsMu sync.Mutex

func settingsFile(paths Paths) string {
	return filepath.Join(paths.Etc, "launcher.json")
}

func loadSettings(paths Paths) Settings {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	s := Settings{AutoDocker: true}
	if data, err := os.ReadFile(settingsFile(paths)); err == nil {
		if err := json.Unmarshal(data, &s); err != nil {
			log.Println("settings:", err)
		}
	}
	return s
}

func saveSettings(paths Paths, s Settings) {
	settingsMu.Lock()
	defer settingsMu.Unlock()

	data, err := json.Marshal(s)
	if err == nil {
		if err = os.MkdirAll(paths.Etc, 0o755); err == nil {
			err = os.WriteFile(settingsFile(paths), data, 0o644)
		}
	}
	if err != nil {
		log.Println("settings:", err)
	}
}
