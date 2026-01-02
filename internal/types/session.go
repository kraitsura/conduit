package types

import (
	"time"
)

// SessionGap is the duration of inactivity that defines a session boundary
const SessionGap = 30 * time.Minute

// ConduitSession represents a work window, always scoped to ONE project.
// Sessions are bounded by 30-minute gaps in activity.
type ConduitSession struct {
	// ID is a unique identifier for this session
	ID string `json:"id"`

	// ProjectPath is the absolute path to the project
	ProjectPath string `json:"project_path"`

	// ProjectName is the display name (basename)
	ProjectName string `json:"project_name"`

	// StartTime is when this session began
	StartTime time.Time `json:"start_time"`

	// EndTime is when this session ended (nil if active)
	EndTime *time.Time `json:"end_time,omitempty"`

	// Chats are the agent conversations during this session
	Chats []AgentChat `json:"chats,omitempty"`

	// Commits made during this session
	Commits []CommitRef `json:"commits,omitempty"`

	// FilesTouched are files modified during this session
	FilesTouched []string `json:"files_touched,omitempty"`

	// Branch is the git branch during this session
	Branch string `json:"branch,omitempty"`

	// IsActive indicates the session is currently ongoing
	IsActive bool `json:"is_active"`
}

// Duration returns the session duration
func (s *ConduitSession) Duration() time.Duration {
	if s.EndTime != nil {
		return s.EndTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}

// TotalChatTime returns the sum of all chat durations in this session
func (s *ConduitSession) TotalChatTime() time.Duration {
	var total time.Duration
	for _, chat := range s.Chats {
		total += chat.Duration()
	}
	return total
}

// AgentChat represents an individual conversation with an AI agent.
type AgentChat struct {
	// ID is a unique identifier for this chat
	ID string `json:"id"`

	// AgentType is the type of agent (claude, cursor, aider, codex, etc.)
	AgentType string `json:"agent_type"`

	// Name is a display name (first message preview or workspace name)
	Name string `json:"name,omitempty"`

	// ProjectPath is the project this chat is associated with
	ProjectPath string `json:"project_path"`

	// StartTime is when this chat began
	StartTime time.Time `json:"start_time"`

	// EndTime is when this chat ended (nil if active)
	EndTime *time.Time `json:"end_time,omitempty"`

	// MessageCount is the number of messages in this chat
	MessageCount int `json:"message_count,omitempty"`

	// IsActive indicates the chat is currently ongoing
	IsActive bool `json:"is_active"`
}

// Duration returns the chat duration
func (c *AgentChat) Duration() time.Duration {
	if c.EndTime != nil {
		return c.EndTime.Sub(c.StartTime)
	}
	return time.Since(c.StartTime)
}

// CommitRef is a lightweight reference to a commit within a session
type CommitRef struct {
	// Hash is the git commit hash
	Hash string `json:"hash"`

	// Message is the commit message (first line)
	Message string `json:"message"`

	// Timestamp is when the commit was made
	Timestamp time.Time `json:"timestamp"`

	// IsAI indicates if this was an AI-authored commit
	IsAI bool `json:"is_ai"`
}

// ProjectSummary provides an aggregated view of a project's current state
type ProjectSummary struct {
	// Path is the absolute path to the project
	Path string `json:"path"`

	// Name is the display name
	Name string `json:"name"`

	// Branch is the current git branch
	Branch string `json:"branch,omitempty"`

	// ActiveAgents are currently running agents in this project
	ActiveAgents []AgentChat `json:"active_agents,omitempty"`

	// CurrentSession is the active session (if any)
	CurrentSession *ConduitSession `json:"current_session,omitempty"`

	// UncommittedFiles is the count of uncommitted changes
	UncommittedFiles int `json:"uncommitted_files"`

	// SessionsToday is the number of sessions today
	SessionsToday int `json:"sessions_today"`

	// AgentTimeToday is the total agent time today
	AgentTimeToday time.Duration `json:"agent_time_today"`

	// CommitsToday is the number of commits today
	CommitsToday int `json:"commits_today"`

	// AICommitsToday is the number of AI-authored commits today
	AICommitsToday int `json:"ai_commits_today"`

	// LastActivity is the most recent activity timestamp
	LastActivity time.Time `json:"last_activity"`
}

// SessionFilter provides options for querying sessions
type SessionFilter struct {
	// ProjectPath filters to a specific project (empty = all)
	ProjectPath string

	// Since filters to sessions starting after this time
	Since time.Time

	// Until filters to sessions starting before this time
	Until time.Time

	// ActiveOnly filters to only active sessions
	ActiveOnly bool

	// Limit caps the number of results (0 = no limit)
	Limit int
}

// DisplayMode determines how the CLI should render output
type DisplayMode int

const (
	// DisplayGlobal shows overview of all projects
	DisplayGlobal DisplayMode = iota

	// DisplayProject shows details for a single project
	DisplayProject
)
