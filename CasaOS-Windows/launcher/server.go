package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/IceWhaleTech/CasaOS-Common/utils/vfs"
	"golang.org/x/sys/windows"
)

// ServerState is the state of the CasaOS server, as shown by the window.
type ServerState int

const (
	StateStopped ServerState = iota
	StateStarting
	StateRunning
	StateStopping
	StateError
)

// service is one of the programs of CasaOS. ready is the file in the runtime folder
// that the program writes when it can be reached by the programs that follow.
type service struct {
	name  string
	ready string
}

// The order is the one of the dependencies between the systemd units on Linux.
var services = []service{
	{"casaos-gateway", "management.url"},
	{"casaos-message-bus", "message-bus.url"},
	{"casaos-user-service", "user-service.url"},
	{"casaos", "casaos.url"},
	{"casaos-local-storage", ""},
	{"casaos-app-management", "app-management.url"},
}

const (
	startTimeout   = 60 * time.Second
	holdTimeout    = 30 * time.Minute
	stopAppsWait   = 15 * time.Second
	maxRestarts    = 5
	restartWindow  = 2 * time.Minute
	createNoWindow = 0x08000000
)

// Paths are the folders used by the server.
type Paths struct {
	Bin, WWW, UIEvents string
	Etc, Lib, Log, Run string
	Share, Data        string
	DataEnv            string
}

func newPaths() Paths {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	dir := filepath.Dir(exe)

	p := Paths{
		Bin:      filepath.Join(dir, "bin"),
		WWW:      filepath.Join(dir, "www"),
		UIEvents: filepath.Join(dir, "ui-message-bus.json"),
		Etc:      vfs.Real("/etc/casaos"),
		Lib:      vfs.Real("/var/lib/casaos"),
		Log:      vfs.Real("/var/log/casaos"),
		Run:      vfs.Real("/var/run/casaos"),
		Share:    vfs.Real("/usr/share/casaos"),
		Data:     vfs.Real("/DATA"),
	}

	// the folder chosen in the installer
	if b, err := os.ReadFile(filepath.Join(dir, "data-path.txt")); err == nil {
		if v := strings.TrimSpace(strings.TrimPrefix(string(b), "\xef\xbb\xbf")); v != "" && os.Getenv("CASAOS_DATA") == "" {
			p.Data = v
			p.DataEnv = v
		}
	}

	return p
}

func (p Paths) prepare() error {
	dirs := []string{p.Etc, p.Lib, filepath.Join(p.Lib, "db"), filepath.Join(p.Lib, "www", "modules"), p.Log, p.Run, p.Share, p.Data}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Server starts and stops the programs of CasaOS.
type Server struct {
	paths Paths
	job   windows.Handle

	mu       sync.Mutex
	state    ServerState
	detail   string
	urlLocal string
	urlLAN   string
	procs    map[string]*exec.Cmd
	restarts map[string][]time.Time
	stopping bool

	onChange func()

	docker *Docker
}

func NewServer(paths Paths, onChange func()) *Server {
	s := &Server{
		paths:    paths,
		procs:    map[string]*exec.Cmd{},
		restarts: map[string][]time.Time{},
		onChange: onChange,
		docker:   NewDocker(paths),
	}
	s.job = newKillOnCloseJob()
	return s
}

// newKillOnCloseJob returns a job that ends its processes when the launcher ends,
// even if it crashes: the server never stays behind without its window.
func newKillOnCloseJob() windows.Handle {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		log.Println("job object:", err)
		return 0
	}

	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		log.Println("job object:", err)
	}
	return job
}

// Snapshot returns the state to show.
func (s *Server) Snapshot() (ServerState, string, string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state, s.detail, s.urlLocal, s.urlLAN
}

func (s *Server) set(state ServerState, detail string) {
	s.mu.Lock()
	s.state, s.detail = state, detail
	s.mu.Unlock()
	if s.onChange != nil {
		s.onChange()
	}
}

// Start starts the server in the background.
func (s *Server) Start() {
	s.mu.Lock()
	if s.state == StateStarting || s.state == StateRunning || s.state == StateStopping {
		s.mu.Unlock()
		return
	}
	s.state, s.detail = StateStarting, tr("starting")
	s.stopping = false
	s.urlLocal, s.urlLAN = "", ""
	s.restarts = map[string][]time.Time{}
	s.mu.Unlock()
	s.onChange()

	go func() {
		if err := s.start(); err != nil {
			log.Println("start failed:", err)
			s.killAll()
			s.set(StateError, err.Error())
		}
	}()
}

func (s *Server) start() error {
	if err := s.paths.prepare(); err != nil {
		return fmt.Errorf("%s: %w", tr("errFolders"), err)
	}

	// what is left of a previous run points at programs that do not exist anymore
	if entries, err := os.ReadDir(s.paths.Run); err == nil {
		for _, e := range entries {
			if err := os.RemoveAll(filepath.Join(s.paths.Run, e.Name())); err != nil {
				log.Println("cleanup:", err)
			}
		}
	}
	freeBusyGatewayPort(filepath.Join(s.paths.Etc, "gateway.ini"))

	for i, svc := range services {
		s.set(StateStarting, fmt.Sprintf(tr("startingService"), i+1, len(services)))
		if err := s.spawn(svc); err != nil {
			return fmt.Errorf("%s %s: %w", tr("errService"), svc.name, err)
		}
		if err := s.waitReady(svc); err != nil {
			return fmt.Errorf("%s %s: %w", tr("errService"), svc.name, err)
		}
	}

	s.set(StateStarting, tr("startingWeb"))

	port, err := s.gatewayPort()
	if err != nil {
		return fmt.Errorf("%s: %w", tr("errPort"), err)
	}

	local := "http://localhost" + portSuffix(port)
	if err := waitHTTP(local+"/v1/users/status", 20*time.Second); err != nil {
		return fmt.Errorf("%s: %w", tr("errWeb"), err)
	}

	s.registerUIEvents()

	lan := ""
	if ip := lanIP(); ip != "" {
		lan = "http://" + ip + portSuffix(port)
	}

	s.mu.Lock()
	s.urlLocal, s.urlLAN = local, lan
	s.state, s.detail = StateRunning, ""
	s.mu.Unlock()
	s.onChange()

	log.Println("server running at", local, lan)

	go s.docker.OnServerStarted()
	go s.watchHolds()

	return nil
}

func (s *Server) spawn(svc service) error {
	exe := filepath.Join(s.paths.Bin, svc.name+".exe")
	if _, err := os.Stat(exe); err != nil {
		return err
	}

	args := []string{}
	if svc.name == "casaos-gateway" {
		args = append(args, "-w", s.paths.WWW)
	}

	logFile, err := os.OpenFile(filepath.Join(s.paths.Log, svc.name+".console.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, args...)
	cmd.Dir = s.paths.Lib
	cmd.Env = os.Environ()
	if s.paths.DataEnv != "" {
		cmd.Env = append(cmd.Env, "CASAOS_DATA="+s.paths.DataEnv)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// the hidden console is inherited by the programs the services run in turn, which
	// so do not open a window of their own
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return err
	}

	if s.job != 0 {
		h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
		if err == nil {
			if err := windows.AssignProcessToJobObject(s.job, h); err != nil {
				log.Println("job object:", err)
			}
			windows.CloseHandle(h) // nolint: errcheck
		}
	}

	s.mu.Lock()
	s.procs[svc.name] = cmd
	s.mu.Unlock()

	log.Println("started", svc.name, "pid", cmd.Process.Pid)

	go func() {
		err := cmd.Wait()
		logFile.Close()
		s.onExit(svc, cmd, err)
	}()

	return nil
}

// onExit restarts a service that ended by itself, as systemd does on Linux. A service
// that keeps on ending stops the whole server, since it would not work anyway.
func (s *Server) onExit(svc service, cmd *exec.Cmd, err error) {
	s.mu.Lock()
	current := s.procs[svc.name] == cmd
	if current {
		delete(s.procs, svc.name)
	}
	stopping := s.stopping
	running := s.state == StateRunning
	s.mu.Unlock()

	if !current || stopping {
		return
	}

	log.Println("service ended:", svc.name, err)

	if !running {
		return // the start sequence reports the failure
	}

	// it may have given up on a service that is being held, which is nobody's fault
	for deadline := time.Now().Add(holdTimeout); s.anyHeld() && time.Now().Before(deadline); {
		time.Sleep(2 * time.Second)
	}

	s.mu.Lock()
	now := time.Now()
	recent := []time.Time{now}
	for _, t := range s.restarts[svc.name] {
		if now.Sub(t) < restartWindow {
			recent = append(recent, t)
		}
	}
	s.restarts[svc.name] = recent
	s.mu.Unlock()

	if len(recent) > maxRestarts {
		s.killAll()
		s.set(StateError, fmt.Sprintf(tr("errCrash"), svc.name))
		return
	}

	time.Sleep(time.Second)

	s.mu.Lock()
	stopping = s.stopping
	s.mu.Unlock()
	if stopping {
		return
	}

	if err := s.spawn(svc); err != nil {
		log.Println("restart failed:", svc.name, err)
	}
}

func (s *Server) waitReady(svc service) error {
	deadline := time.Now().Add(startTimeout)
	if svc.ready == "" {
		deadline = time.Now().Add(2 * time.Second)
	}
	holdLimit := time.Now().Add(holdTimeout)
	nextCheck := time.Now().Add(2 * time.Second)
	lastHeld := time.Time{}

	for time.Now().Before(deadline) {
		s.mu.Lock()
		_, alive := s.procs[svc.name]
		stopping := s.stopping
		s.mu.Unlock()

		if stopping {
			return errors.New(tr("stopping"))
		}

		if !alive {
			// it may have given up on a service that is being held: it gets another try
			if time.Since(lastHeld) > 30*time.Second {
				return errors.New(tr("errEnded"))
			}
			if err := s.spawn(svc); err != nil {
				return err
			}
		}

		if svc.ready != "" {
			if fi, err := os.Stat(filepath.Join(s.paths.Run, svc.ready)); err == nil && fi.Size() > 0 {
				return nil
			}
		}

		if time.Now().After(nextCheck) {
			nextCheck = time.Now().Add(2 * time.Second)
			if s.anyHeld() {
				lastHeld = time.Now()
				deadline = holdLimit
				s.set(StateStarting, tr("heldByAntivirus"))
			} else if !lastHeld.IsZero() && time.Since(lastHeld) > 3*time.Second {
				// released: the usual time is enough from here on
				lastHeld = time.Time{}
				deadline = time.Now().Add(startTimeout)
				s.set(StateStarting, tr("starting"))
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	if svc.ready == "" {
		return nil
	}
	return errors.New(tr("errTimeout"))
}

func (s *Server) gatewayPort() (string, error) {
	address, err := os.ReadFile(filepath.Join(s.paths.Run, "management.url"))
	if err != nil {
		return "", err
	}

	client := http.Client{Timeout: 5 * time.Second}

	var lastErr error
	for i := 0; i < 30; i++ {
		resp, err := client.Get(strings.TrimSpace(string(address)) + "/v1/gateway/port")
		if err == nil {
			var body struct {
				Data string `json:"data"`
			}
			err = json.NewDecoder(resp.Body).Decode(&body)
			resp.Body.Close()
			if err == nil && body.Data != "" {
				return body.Data, nil
			}
		}
		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}
	return "", fmt.Errorf("no port: %v", lastErr)
}

// heldProcess tells if every thread of a process is suspended. CasaOS never does that to
// itself: it is what a security program does while it analyzes an unknown program.
func (s *Server) anyHeld() bool {
	s.mu.Lock()
	pids := make([]uint32, 0, len(s.procs))
	for _, cmd := range s.procs {
		if cmd.Process != nil {
			pids = append(pids, uint32(cmd.Process.Pid))
		}
	}
	s.mu.Unlock()

	return anyProcessSuspended(pids)
}

// watchHolds tells the user when a running service is being held, since the dashboard
// stops answering without any other explanation.
func (s *Server) watchHolds() {
	for {
		time.Sleep(5 * time.Second)

		s.mu.Lock()
		running := s.state == StateRunning
		current := s.detail
		s.mu.Unlock()

		if !running {
			return
		}

		detail := ""
		if s.anyHeld() {
			detail = tr("heldByAntivirus")
		}
		if detail != current {
			s.set(StateRunning, detail)
		}
	}
}

// registerUIEvents does what start.d/register-ui-events.sh does on Linux.
func (s *Server) registerUIEvents() {
	events, err := os.ReadFile(s.paths.UIEvents)
	if err != nil {
		log.Println("ui events:", err)
		return
	}
	address, err := os.ReadFile(filepath.Join(s.paths.Run, "message-bus.url"))
	if err != nil {
		log.Println("ui events:", err)
		return
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(strings.TrimSpace(string(address))+"/v2/message_bus/event_type", "application/json", bytes.NewReader(events))
	if err != nil {
		log.Println("ui events:", err)
		return
	}
	io.Copy(io.Discard, resp.Body) // nolint: errcheck
	resp.Body.Close()
}

// Stop stops the server in the background.
func (s *Server) Stop(done func()) {
	s.mu.Lock()
	if s.state == StateStopped || s.state == StateError || s.state == StateStopping {
		s.mu.Unlock()
		if done != nil {
			done()
		}
		return
	}
	s.state, s.detail = StateStopping, tr("stopping")
	s.stopping = true
	s.mu.Unlock()
	s.onChange()

	go func() {
		// the apps stop on their own time: a slow Docker must not hold up the server
		appsStopped := make(chan struct{})
		go func() {
			s.docker.OnServerStopping()
			close(appsStopped)
		}()

		s.killAll()
		s.cleanRuntime()

		s.mu.Lock()
		s.state, s.detail = StateStopped, ""
		s.urlLocal, s.urlLAN = "", ""
		s.mu.Unlock()
		s.onChange()

		log.Println("server stopped")

		if done != nil {
			// the launcher is closing: give the apps a moment before it goes
			select {
			case <-appsStopped:
			case <-time.After(stopAppsWait):
			}
			done()
		}
	}()
}

func (s *Server) killAll() {
	s.mu.Lock()
	s.stopping = true
	procs := make([]*exec.Cmd, 0, len(s.procs))
	for i := len(services) - 1; i >= 0; i-- {
		if cmd, ok := s.procs[services[i].name]; ok {
			procs = append(procs, cmd)
		}
	}
	s.procs = map[string]*exec.Cmd{}
	s.mu.Unlock()

	for _, cmd := range procs {
		if err := cmd.Process.Kill(); err != nil {
			log.Println("kill:", err)
		}
	}
}

func (s *Server) cleanRuntime() {
	// the processes need a moment to release their files
	time.Sleep(300 * time.Millisecond)
	if entries, err := os.ReadDir(s.paths.Run); err == nil {
		for _, e := range entries {
			os.RemoveAll(filepath.Join(s.paths.Run, e.Name())) // nolint: errcheck
		}
	}
}

// StopNow ends the processes at once: Windows is shutting down.
func (s *Server) StopNow() {
	s.killAll()
}

func portSuffix(port string) string {
	if port == "" || port == "80" {
		return ""
	}
	return ":" + port
}

func waitHTTP(url string, timeout time.Duration) error {
	client := http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			io.Copy(io.Discard, resp.Body) // nolint: errcheck
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			err = fmt.Errorf("status %d", resp.StatusCode)
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	return lastErr
}

// lanIP returns the address of this computer on the local network. No packet is sent:
// connecting a UDP socket only makes Windows choose the interface of the default route.
func lanIP() string {
	conn, err := net.Dial("udp4", "8.8.8.8:53")
	if err != nil {
		return ""
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP.IsLoopback() {
		return ""
	}
	return addr.IP.String()
}

var iniPort = regexp.MustCompile(`(?im)^(\s*port\s*=\s*)(\d+)\s*$`)

// freeBusyGatewayPort empties the port in the config of the gateway when another
// program has taken it since the last run: the gateway then looks for a free one.
func freeBusyGatewayPort(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	match := iniPort.FindSubmatch(content)
	if match == nil {
		return
	}

	port, err := strconv.Atoi(string(match[2]))
	if err != nil {
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		listener.Close()
		return
	}

	log.Println("port", port, "is busy, the gateway will choose another one")
	if err := os.WriteFile(path, iniPort.ReplaceAll(content, []byte("${1}")), 0o644); err != nil {
		log.Println("gateway.ini:", err)
	}
}
