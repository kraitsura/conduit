package agent

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// Detector finds running AI agent processes
type Detector struct {
	patterns []*regexp.Regexp
}

// NewDetector creates a detector with the given patterns
func NewDetector(patterns []string) *Detector {
	d := &Detector{
		patterns: make([]*regexp.Regexp, 0, len(patterns)),
	}

	for _, p := range patterns {
		// Case-insensitive matching
		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(p))
		if err == nil {
			d.patterns = append(d.patterns, re)
		}
	}

	return d
}

// ProcessInfo holds raw process information
type ProcessInfo struct {
	PID     int
	Command string
	CWD     string
}

// DetectAgents finds all running AI agents and their working directories
func (d *Detector) DetectAgents() ([]types.Agent, error) {
	// Get all processes with their commands
	processes, err := d.getProcesses()
	if err != nil {
		return nil, err
	}

	var agents []types.Agent

	for _, proc := range processes {
		agentType := d.matchAgentType(proc.Command)
		if agentType == "" {
			continue
		}

		// Get working directory for this process
		cwd, err := d.getProcessCWD(proc.PID)
		if err != nil {
			cwd = "" // May not have permission
		}

		agents = append(agents, types.Agent{
			PID:         proc.PID,
			Type:        agentType,
			ProcessName: extractProcessName(proc.Command),
			StartTime:   time.Now(), // We'd need to track this separately
			ProjectPath: cwd,
			Command:     proc.Command,
		})
	}

	return agents, nil
}

// DetectAgentsInPath finds agents running in a specific directory (or subdirs)
func (d *Detector) DetectAgentsInPath(rootPath string) ([]types.Agent, error) {
	allAgents, err := d.DetectAgents()
	if err != nil {
		return nil, err
	}

	var filtered []types.Agent
	for _, agent := range allAgents {
		if strings.HasPrefix(agent.ProjectPath, rootPath) {
			filtered = append(filtered, agent)
		}
	}

	return filtered, nil
}

// getProcesses returns all running processes using ps
func (d *Detector) getProcesses() ([]ProcessInfo, error) {
	// ps -eo pid,command - shows PID and full command
	cmd := exec.Command("ps", "-eo", "pid,command")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var processes []ProcessInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))

	// Skip header line
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split into PID and command (command may have spaces)
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		pid, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}

		processes = append(processes, ProcessInfo{
			PID:     pid,
			Command: strings.TrimSpace(parts[1]),
		})
	}

	return processes, nil
}

// getProcessCWD gets the current working directory of a process
func (d *Detector) getProcessCWD(pid int) (string, error) {
	// On macOS, use lsof to get cwd
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-Fn")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Look for 'cwd' type and extract path
	scanner := bufio.NewScanner(bytes.NewReader(output))
	foundCWD := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "fcwd" {
			foundCWD = true
			continue
		}
		if foundCWD && strings.HasPrefix(line, "n") {
			// This is the path
			return line[1:], nil
		}
	}

	return "", nil
}

// matchAgentType checks if a command matches any agent pattern
func (d *Detector) matchAgentType(command string) string {
	for _, re := range d.patterns {
		if re.MatchString(command) {
			// Extract the matched pattern as type
			return strings.ToLower(re.FindString(command))
		}
	}
	return ""
}

// extractProcessName gets the base process name from command
func extractProcessName(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	// Get the executable name (last component of path)
	exe := parts[0]
	if idx := strings.LastIndex(exe, "/"); idx >= 0 {
		exe = exe[idx+1:]
	}

	return exe
}

// GroupAgentsByProject groups agents by their project path
func GroupAgentsByProject(agents []types.Agent) map[string][]types.Agent {
	grouped := make(map[string][]types.Agent)
	for _, agent := range agents {
		if agent.ProjectPath != "" {
			grouped[agent.ProjectPath] = append(grouped[agent.ProjectPath], agent)
		}
	}
	return grouped
}
