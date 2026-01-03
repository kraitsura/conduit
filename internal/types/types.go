package types

import (
	"time"
)

// Project represents a tracked project directory
type Project struct {
	// Path is the absolute path to the project
	Path string `json:"path"`

	// Name is the display name (usually directory basename)
	Name string `json:"name"`

	// GitRemote is the primary git remote URL (if available)
	GitRemote string `json:"git_remote,omitempty"`

	// LastActivity is when we last saw activity in this project
	LastActivity time.Time `json:"last_activity"`

	// ActiveAgents are currently running agents in this project
	ActiveAgents []Agent `json:"active_agents,omitempty"`

	// Stats contains aggregated statistics
	Stats *ProjectStats `json:"stats,omitempty"`
}

// Agent represents a detected AI coding agent process
type Agent struct {
	// PID is the process ID
	PID int `json:"pid"`

	// Type is the agent type (claude, cursor, aider, etc.)
	Type string `json:"type"`

	// ProcessName is the actual process name
	ProcessName string `json:"process_name"`

	// StartTime is when this agent was first detected
	StartTime time.Time `json:"start_time"`

	// ProjectPath is the working directory
	ProjectPath string `json:"project_path"`

	// Command is the full command line (if available)
	Command string `json:"command,omitempty"`
}

// Activity represents a tracked activity event
type Activity struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Project   string    `json:"project"`
	Type      string    `json:"type"` // agent_start, agent_stop, commit, file_change
	AgentType string    `json:"agent_type,omitempty"`
	Data      string    `json:"data,omitempty"` // JSON blob for extra data
}

// ActivityType constants
const (
	ActivityAgentStart = "agent_start"
	ActivityAgentStop  = "agent_stop"
	ActivityCommit     = "commit"
	ActivityBranch     = "branch"
	ActivityFileChange = "file_change"
)

// Insight represents a user-captured note, bug, or idea
type Insight struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // note, bug, idea
	Content     string    `json:"content"`
	ProjectPath string    `json:"project_path,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	FilePath    string    `json:"file_path,omitempty"`
	SessionID   string    `json:"session_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// InsightType constants
const (
	InsightNote = "note"
	InsightBug  = "bug"
	InsightIdea = "idea"
)

// InsightFilter is used to query insights
type InsightFilter struct {
	ProjectPath string
	Type        string // note, bug, idea, or empty for all
	Since       time.Time
	Limit       int
}

// ProjectStats contains aggregated project statistics
type ProjectStats struct {
	TotalAgentTime   time.Duration `json:"total_agent_time"`
	CommitsToday     int           `json:"commits_today"`
	CommitsThisWeek  int           `json:"commits_this_week"`
	LastCommitTime   time.Time     `json:"last_commit_time,omitempty"`
	LastCommitAuthor string        `json:"last_commit_author,omitempty"`
	FilesChanged     int           `json:"files_changed"`
}

// Commit represents a git commit
type Commit struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	IsAI      bool      `json:"is_ai"` // Detected as AI-authored
}

// DaemonStatus represents the daemon's current state
type DaemonStatus struct {
	Running      bool      `json:"running"`
	PID          int       `json:"pid,omitempty"`
	StartTime    time.Time `json:"start_time,omitempty"`
	RootPath     string    `json:"root_path"`
	ProjectCount int       `json:"project_count"`
	ActiveAgents int       `json:"active_agents"`
	LastPoll     time.Time `json:"last_poll,omitempty"`
}
