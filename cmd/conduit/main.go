package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kraitsura/conduit/internal/config"
	"github.com/kraitsura/conduit/internal/daemon"
	"github.com/kraitsura/conduit/internal/git"
	"github.com/kraitsura/conduit/internal/project"
	"github.com/kraitsura/conduit/internal/types"
)

func main() {
	if len(os.Args) < 2 {
		// Default: show context-aware status
		showContextAwareStatus(nil)
		return
	}

	// Handle global flag at top level
	if os.Args[1] == "--global" || os.Args[1] == "-g" {
		showContextAwareStatus(os.Args[1:])
		return
	}

	switch os.Args[1] {
	case "status", "s":
		showContextAwareStatus(os.Args[2:])

	case "projects", "p", "ls":
		listProjects()

	case "log", "l":
		projectPath := ""
		if len(os.Args) > 2 {
			projectPath = os.Args[2]
		}
		showLog(projectPath)

	case "agents", "a":
		showAgents()

	case "sessions":
		showSessions(os.Args[2:])

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: conduit search <query>")
			return
		}
		searchSessions(os.Args[2], os.Args[3:])

	case "stats":
		showStats(os.Args[2:])

	case "cd":
		if len(os.Args) < 3 {
			fmt.Println("Usage: conduit cd <project>")
			return
		}
		printProjectPath(os.Args[2])

	case "init":
		initConduit()

	case "daemon":
		daemonCmd()

	case "help", "h", "-h", "--help":
		showHelp()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		showHelp()
		os.Exit(1)
	}
}

// Request/Response types (must match daemon)
type Request struct {
	Command string `json:"command"`
	Project string `json:"project,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

type Response struct {
	Status     string              `json:"status"`
	Error      string              `json:"error,omitempty"`
	Daemon     *types.DaemonStatus `json:"daemon,omitempty"`
	Projects   []types.Project     `json:"projects,omitempty"`
	Agents     []types.Agent       `json:"agents,omitempty"`
	Activities []types.Activity    `json:"activities,omitempty"`
}

func sendRequest(req Request) (*Response, error) {
	// Auto-start daemon if not running
	daemon.EnsureRunning()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", cfg.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to daemon")
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(req); err != nil {
		return nil, err
	}

	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Status == "error" {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return &resp, nil
}

func showStatus(projectName string) {
	resp, err := sendRequest(Request{Command: "status"})
	if err != nil {
		// Fallback: show local status without daemon
		showLocalStatus()
		return
	}

	// Daemon status header
	fmt.Println("\033[1m CONDUIT \033[0m")
	fmt.Printf("Watching: %s\n", resp.Daemon.RootPath)
	fmt.Printf("Projects: %d | Active Agents: %d\n\n", resp.Daemon.ProjectCount, resp.Daemon.ActiveAgents)

	// Current project context
	cwd, _ := os.Getwd()
	currentProject := findProjectForPath(resp.Projects, cwd)

	if currentProject != nil {
		showProjectDetail(currentProject, resp.Agents)
	} else if len(resp.Agents) > 0 {
		fmt.Println("\033[1mActive Agents:\033[0m")
		for _, a := range resp.Agents {
			projectName := filepath.Base(a.ProjectPath)
			duration := time.Since(a.StartTime).Round(time.Minute)
			fmt.Printf("  %s %s in \033[36m%s\033[0m (%s)\n",
				agentIcon(a.Type), a.Type, projectName, formatDuration(duration))
		}
		fmt.Println()
	}

	// Recent projects
	fmt.Println("\033[1mRecent Projects:\033[0m")
	project.SortByActivity(resp.Projects)
	shown := 0
	for _, p := range resp.Projects {
		if shown >= 5 {
			break
		}

		status := "\033[90m-\033[0m"
		agentCount := countAgentsInProject(resp.Agents, p.Path)
		if agentCount > 0 {
			status = fmt.Sprintf("\033[32m%d agents\033[0m", agentCount)
		}

		ago := time.Since(p.LastActivity)
		fmt.Printf("  %-20s %s  %s ago\n", p.Name, status, formatDuration(ago))
		shown++
	}
}

func showLocalStatus() {
	// Show status without daemon (direct git/filesystem access)
	cfg, _ := config.Load()

	fmt.Println("\033[1m CONDUIT \033[0m (daemon not running)")
	fmt.Printf("Configured root: %s\n\n", cfg.RootPath)

	// Check if we're in a project
	cwd, _ := os.Getwd()

	if git.IsGitRepo(cwd) {
		fmt.Println("\033[1mCurrent Project:\033[0m", filepath.Base(cwd))

		branch := git.GetCurrentBranch(cwd)
		if branch != "" {
			fmt.Printf("  Branch: %s\n", branch)
		}

		uncommitted := git.GetUncommittedChanges(cwd)
		if uncommitted > 0 {
			fmt.Printf("  Uncommitted: %d files\n", uncommitted)
		}

		if commits, _ := git.GetCommitsToday(cwd); len(commits) > 0 {
			fmt.Printf("  Today: %d commits\n", len(commits))
		}

		if lastCommit, _ := git.GetLastCommit(cwd); lastCommit != nil {
			ago := time.Since(lastCommit.Timestamp)
			aiTag := ""
			if lastCommit.IsAI {
				aiTag = " \033[33m[AI]\033[0m"
			}
			fmt.Printf("  Last commit: %s ago%s\n", formatDuration(ago), aiTag)
			fmt.Printf("    %s\n", truncate(lastCommit.Message, 50))
		}
	}

	fmt.Println("\nStart daemon: \033[36mconduit daemon start\033[0m")
}

func showProjectDetail(p *types.Project, agents []types.Agent) {
	fmt.Printf("\033[1mCurrent: %s\033[0m\n", p.Name)

	// Active agents in this project
	projectAgents := filterAgentsForProject(agents, p.Path)
	if len(projectAgents) > 0 {
		for _, a := range projectAgents {
			duration := time.Since(a.StartTime).Round(time.Minute)
			fmt.Printf("  %s %s running (%s)\n", agentIcon(a.Type), a.Type, formatDuration(duration))
		}
	}

	// Git status
	branch := git.GetCurrentBranch(p.Path)
	if branch != "" {
		fmt.Printf("  Branch: %s\n", branch)
	}

	uncommitted := git.GetUncommittedChanges(p.Path)
	if uncommitted > 0 {
		fmt.Printf("  Uncommitted: %d files\n", uncommitted)
	}

	if commits, _ := git.GetCommitsToday(p.Path); len(commits) > 0 {
		aiCount := 0
		for _, c := range commits {
			if c.IsAI {
				aiCount++
			}
		}
		if aiCount > 0 {
			fmt.Printf("  Today: %d commits (%d AI-authored)\n", len(commits), aiCount)
		} else {
			fmt.Printf("  Today: %d commits\n", len(commits))
		}
	}

	fmt.Println()
}

func listProjects() {
	resp, err := sendRequest(Request{Command: "projects"})
	if err != nil {
		// Fallback: discover locally
		cfg, _ := config.Load()
		projects, err := project.DiscoverProjects(cfg.RootPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		resp = &Response{Projects: projects}
	}

	fmt.Println("\033[1mProjects:\033[0m")
	project.SortByActivity(resp.Projects)

	for _, p := range resp.Projects {
		ago := time.Since(p.LastActivity)
		fmt.Printf("  %-25s %s ago\n", p.Name, formatDuration(ago))
	}
}

func showLog(projectPath string) {
	// If no project specified, use current directory
	if projectPath == "" {
		cwd, _ := os.Getwd()
		projectPath = cwd
	}

	resp, err := sendRequest(Request{
		Command: "activities",
		Project: projectPath,
		Limit:   20,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\033[1mActivity Log:\033[0m")
	for _, a := range resp.Activities {
		ago := time.Since(a.Timestamp)
		icon := activityIcon(a.Type)
		projectName := filepath.Base(a.Project)

		desc := a.Type
		if a.AgentType != "" {
			desc = fmt.Sprintf("%s (%s)", a.Type, a.AgentType)
		}

		fmt.Printf("  %s %s  \033[36m%s\033[0m  %s ago\n",
			icon, desc, projectName, formatDuration(ago))
	}
}

func showAgents() {
	resp, err := sendRequest(Request{Command: "agents"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(resp.Agents) == 0 {
		fmt.Println("No active agents")
		return
	}

	fmt.Println("\033[1mActive Agents:\033[0m")
	for _, a := range resp.Agents {
		projectName := filepath.Base(a.ProjectPath)
		duration := time.Since(a.StartTime).Round(time.Minute)
		fmt.Printf("  %s %-10s PID %-6d \033[36m%s\033[0m  %s\n",
			agentIcon(a.Type), a.Type, a.PID, projectName, formatDuration(duration))
	}
}

func initConduit() {
	cfg := config.DefaultConfig()

	// Check if root path should be current directory
	if len(os.Args) > 2 && os.Args[2] == "." {
		cwd, _ := os.Getwd()
		cfg.RootPath = cwd
	}

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Println("Conduit initialized!")
	fmt.Printf("Config: %s\n", config.ConfigPath())
	fmt.Printf("Watching: %s\n", cfg.RootPath)
	fmt.Println("\nRun `conduit` to start tracking.")
}

func daemonCmd() {
	// `conduit daemon` with no args = run in foreground (for debugging)
	// `conduit daemon run` = same (used by auto-start)
	if len(os.Args) < 3 || os.Args[2] == "run" {
		if err := daemon.Run(); err != nil {
			fmt.Printf("Daemon error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Other subcommands (mostly for debugging)
	switch os.Args[2] {
	case "status":
		if daemon.IsRunning() {
			resp, _ := sendRequest(Request{Command: "status"})
			if resp != nil && resp.Daemon != nil {
				fmt.Printf("Daemon: running (PID %d, %d agents)\n", resp.Daemon.PID, resp.Daemon.ActiveAgents)
			} else {
				fmt.Println("Daemon: running")
			}
		} else {
			fmt.Println("Daemon: not running")
		}
	default:
		fmt.Println("Usage: conduit daemon [run|status]")
	}
}

func showHelp() {
	help := `
Conduit - Agentic Developer Observability

Usage: conduit [command]

Commands:
  (default)     Show status (adapts to context)
  status, s     Show status for a project
  projects, p   List all tracked projects
  agents, a     Show active AI agents
  sessions      Show session history
  search        Search sessions by name
  stats         Show statistics
  log, l        Show activity log
  cd            Print project path (for shell)
  init          Initialize config
  help, h       Show this help

Flags:
  --global, -g  Show global view (all projects)
  --today, -t   Filter to today
  --week, -w    Filter to this week

The daemon auto-starts when needed and auto-stops after 30 min idle.

Examples:
  conduit                    # Context-aware status
  conduit --global           # Global overview
  conduit sessions --today   # Today's sessions
  conduit search "auth"      # Search sessions
  conduit stats --week       # Weekly statistics
  conduit cd myproject       # Print project path
`
	fmt.Println(help)
}

// Helpers

func findProjectForPath(projects []types.Project, path string) *types.Project {
	for i := range projects {
		if path == projects[i].Path || strings.HasPrefix(path, projects[i].Path+"/") {
			return &projects[i]
		}
	}
	return nil
}

func filterAgentsForProject(agents []types.Agent, projectPath string) []types.Agent {
	var filtered []types.Agent
	for _, a := range agents {
		if a.ProjectPath == projectPath {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func countAgentsInProject(agents []types.Agent, projectPath string) int {
	count := 0
	for _, a := range agents {
		if a.ProjectPath == projectPath {
			count++
		}
	}
	return count
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func agentIcon(agentType string) string {
	switch strings.ToLower(agentType) {
	case "claude":
		return ""
	case "cursor":
		return ""
	case "aider":
		return ""
	default:
		return ""
	}
}

func activityIcon(activityType string) string {
	switch activityType {
	case types.ActivityAgentStart:
		return ""
	case types.ActivityAgentStop:
		return ""
	case types.ActivityCommit:
		return ""
	default:
		return ""
	}
}
