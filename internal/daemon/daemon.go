package daemon

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/kraitsura/conduit/internal/agent"
	"github.com/kraitsura/conduit/internal/config"
	"github.com/kraitsura/conduit/internal/project"
	"github.com/kraitsura/conduit/internal/store"
	"github.com/kraitsura/conduit/internal/types"
)

const (
	// IdleTimeout - shutdown after this duration with no agents
	IdleTimeout = 30 * time.Minute
)

// Daemon is the main daemon process
type Daemon struct {
	cfg          *config.Config
	store        *store.Store
	detector     *agent.Detector
	projects     []types.Project
	agents       map[int]types.Agent
	listener     net.Listener
	startTime    time.Time
	lastActivity time.Time // Last time we saw any agent
}

// Request from CLI
type Request struct {
	Command string `json:"command"`
	Project string `json:"project,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// Response to CLI
type Response struct {
	Status     string              `json:"status"`
	Error      string              `json:"error,omitempty"`
	Daemon     *types.DaemonStatus `json:"daemon,omitempty"`
	Projects   []types.Project     `json:"projects,omitempty"`
	Agents     []types.Agent       `json:"agents,omitempty"`
	Activities []types.Activity    `json:"activities,omitempty"`
}

// Run starts the daemon (blocking) - used internally
func Run() error {
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("Conduit daemon starting...")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := config.EnsureDirectories(); err != nil {
		return err
	}

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now()
	d := &Daemon{
		cfg:          cfg,
		store:        db,
		detector:     agent.NewDetector(cfg.AgentPatterns),
		agents:       make(map[int]types.Agent),
		startTime:    now,
		lastActivity: now,
	}

	log.Printf("Watching: %s", cfg.RootPath)
	d.discoverProjects()

	if err := d.startSocketServer(); err != nil {
		return err
	}
	defer d.listener.Close()
	defer os.Remove(cfg.SocketPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down (signal)...")
		cancel()
	}()

	d.run(ctx, cancel)
	return nil
}

// EnsureRunning starts daemon if not already running, returns when ready
func EnsureRunning() error {
	if IsRunning() {
		return nil
	}

	// Start in background
	if err := startBackground(); err != nil {
		return err
	}

	// Wait for it to be ready (max 3 seconds)
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if IsRunning() {
			return nil
		}
	}

	return nil // Best effort - might still be starting
}

// startBackground spawns daemon as detached process
func startBackground() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := exec.Command(exe, "daemon", "run")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	return cmd.Start()
}

// IsRunning checks if daemon is running
func IsRunning() bool {
	cfg, err := config.Load()
	if err != nil {
		return false
	}

	conn, err := net.Dial("unix", cfg.SocketPath)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (d *Daemon) discoverProjects() {
	projects, err := project.DiscoverProjects(d.cfg.RootPath)
	if err != nil {
		log.Printf("Error discovering projects: %v", err)
		return
	}

	d.projects = projects
	log.Printf("Found %d projects", len(projects))

	for _, p := range projects {
		d.store.SaveProject(&p)
	}
}

func (d *Daemon) run(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(time.Duration(d.cfg.PollInterval) * time.Second)
	defer ticker.Stop()

	idleCheck := time.NewTicker(1 * time.Minute)
	defer idleCheck.Stop()

	// Sync Claude chats less frequently
	chatSync := time.NewTicker(30 * time.Second)
	defer chatSync.Stop()

	d.poll()
	d.syncClaudeChats()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			d.poll()

		case <-chatSync.C:
			d.syncClaudeChats()
			d.checkSessionTimeouts()

		case <-idleCheck.C:
			// Auto-shutdown if idle too long
			if len(d.agents) == 0 && time.Since(d.lastActivity) > IdleTimeout {
				log.Printf("No activity for %v, shutting down...", IdleTimeout)
				cancel()
				return
			}
		}
	}
}

func (d *Daemon) poll() {
	agents, err := d.detector.DetectAgentsInPath(d.cfg.RootPath)
	if err != nil {
		log.Printf("Error detecting agents: %v", err)
		return
	}

	// Track new agents
	currentPIDs := make(map[int]bool)
	for _, a := range agents {
		currentPIDs[a.PID] = true

		if _, exists := d.agents[a.PID]; !exists {
			log.Printf("+ %s (PID %d) in %s", a.Type, a.PID, a.ProjectPath)
			d.agents[a.PID] = a
			d.lastActivity = time.Now()

			// Ensure there's an active conduit session for this project
			d.ensureActiveSession(a.ProjectPath)

			d.store.StartAgentSession(&a)
			d.store.LogActivity(&types.Activity{
				Timestamp: time.Now(),
				Project:   a.ProjectPath,
				Type:      types.ActivityAgentStart,
				AgentType: a.Type,
			})
		}
	}

	// Track stopped agents
	for pid, a := range d.agents {
		if !currentPIDs[pid] {
			log.Printf("- %s (PID %d)", a.Type, pid)
			delete(d.agents, pid)
			d.lastActivity = time.Now()

			d.store.EndAgentSession(pid)
			d.store.LogActivity(&types.Activity{
				Timestamp: time.Now(),
				Project:   a.ProjectPath,
				Type:      types.ActivityAgentStop,
				AgentType: a.Type,
			})
		}
	}

	// Keep lastActivity fresh if agents are running
	if len(d.agents) > 0 {
		d.lastActivity = time.Now()
	}

	// Update project timestamps
	for _, a := range d.agents {
		for i := range d.projects {
			if d.projects[i].Path == a.ProjectPath {
				d.projects[i].LastActivity = time.Now()
				d.store.SaveProject(&d.projects[i])
				break
			}
		}
	}
}

func (d *Daemon) startSocketServer() error {
	os.Remove(d.cfg.SocketPath)

	listener, err := net.Listen("unix", d.cfg.SocketPath)
	if err != nil {
		return err
	}

	d.listener = listener
	go d.acceptConnections()

	log.Printf("Ready")
	return nil
}

func (d *Daemon) acceptConnections() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			return
		}
		go d.handleConnection(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Any CLI connection counts as activity
	d.lastActivity = time.Now()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var req Request
	if err := decoder.Decode(&req); err != nil {
		encoder.Encode(Response{Status: "error", Error: err.Error()})
		return
	}

	var resp Response

	switch req.Command {
	case "status":
		resp = d.handleStatus()
	case "projects":
		resp = d.handleProjects()
	case "agents":
		resp = d.handleAgents(req.Project)
	case "activities":
		resp = d.handleActivities(req.Project, req.Limit)
	case "refresh":
		d.discoverProjects()
		resp = Response{Status: "ok"}
	case "current_session":
		sessResp := d.handleCurrentSession(req.Project)
		resp = sessResp.Response
		encoder.Encode(sessResp)
		return
	case "sessions":
		sessResp := d.handleSessions(req.Project, 0, 0, false, req.Limit)
		resp = sessResp.Response
		encoder.Encode(sessResp)
		return
	case "project_summary":
		sessResp := d.handleProjectSummary(req.Project)
		resp = sessResp.Response
		encoder.Encode(sessResp)
		return
	default:
		resp = Response{Status: "error", Error: "unknown command"}
	}

	encoder.Encode(resp)
}

func (d *Daemon) handleStatus() Response {
	var agents []types.Agent
	for _, a := range d.agents {
		agents = append(agents, a)
	}

	return Response{
		Status: "ok",
		Daemon: &types.DaemonStatus{
			Running:      true,
			PID:          os.Getpid(),
			StartTime:    d.startTime,
			RootPath:     d.cfg.RootPath,
			ProjectCount: len(d.projects),
			ActiveAgents: len(d.agents),
			LastPoll:     time.Now(),
		},
		Projects: d.projects,
		Agents:   agents,
	}
}

func (d *Daemon) handleProjects() Response {
	return Response{
		Status:   "ok",
		Projects: d.projects,
	}
}

func (d *Daemon) handleAgents(projectPath string) Response {
	var agents []types.Agent
	for _, a := range d.agents {
		if projectPath == "" || a.ProjectPath == projectPath {
			agents = append(agents, a)
		}
	}
	return Response{
		Status: "ok",
		Agents: agents,
	}
}

func (d *Daemon) handleActivities(projectPath string, limit int) Response {
	if limit == 0 {
		limit = 50
	}

	activities, err := d.store.GetActivities(projectPath, limit)
	if err != nil {
		return Response{Status: "error", Error: err.Error()}
	}

	return Response{
		Status:     "ok",
		Activities: activities,
	}
}
