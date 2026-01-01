// Package testutil provides shared testing utilities for the conduit project.
//
// This package establishes testing patterns used across all packages:
//   - Table-driven tests with descriptive test case names
//   - Builders for creating test fixtures
//   - Helpers for temporary directories and files
//   - In-memory database setup for store tests
package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// TempDir creates a temporary directory and returns a cleanup function.
// The directory is automatically removed when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "conduit-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// TempFile creates a temporary file with the given content.
func TempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

// MkdirAll creates a directory tree within a temp directory.
func MkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

// CreateGitRepo creates a fake git repository for testing.
// This creates a .git directory without actual git initialization.
func CreateGitRepo(t *testing.T, dir, name string) string {
	t.Helper()
	repoPath := filepath.Join(dir, name)
	gitPath := filepath.Join(repoPath, ".git")
	MkdirAll(t, gitPath)
	return repoPath
}

// ProjectBuilder provides a fluent interface for creating test projects.
type ProjectBuilder struct {
	project types.Project
}

// NewProject creates a new ProjectBuilder with defaults.
func NewProject() *ProjectBuilder {
	return &ProjectBuilder{
		project: types.Project{
			Path:         "/test/project",
			Name:         "test-project",
			LastActivity: time.Now(),
		},
	}
}

// WithPath sets the project path.
func (b *ProjectBuilder) WithPath(path string) *ProjectBuilder {
	b.project.Path = path
	return b
}

// WithName sets the project name.
func (b *ProjectBuilder) WithName(name string) *ProjectBuilder {
	b.project.Name = name
	return b
}

// WithGitRemote sets the git remote URL.
func (b *ProjectBuilder) WithGitRemote(remote string) *ProjectBuilder {
	b.project.GitRemote = remote
	return b
}

// WithLastActivity sets the last activity timestamp.
func (b *ProjectBuilder) WithLastActivity(t time.Time) *ProjectBuilder {
	b.project.LastActivity = t
	return b
}

// WithAgents adds active agents to the project.
func (b *ProjectBuilder) WithAgents(agents ...types.Agent) *ProjectBuilder {
	b.project.ActiveAgents = agents
	return b
}

// WithStats sets project statistics.
func (b *ProjectBuilder) WithStats(stats *types.ProjectStats) *ProjectBuilder {
	b.project.Stats = stats
	return b
}

// Build returns the constructed project.
func (b *ProjectBuilder) Build() types.Project {
	return b.project
}

// AgentBuilder provides a fluent interface for creating test agents.
type AgentBuilder struct {
	agent types.Agent
}

// NewAgent creates a new AgentBuilder with defaults.
func NewAgent() *AgentBuilder {
	return &AgentBuilder{
		agent: types.Agent{
			PID:         1234,
			Type:        "claude",
			ProcessName: "claude",
			StartTime:   time.Now(),
			ProjectPath: "/test/project",
		},
	}
}

// WithPID sets the process ID.
func (b *AgentBuilder) WithPID(pid int) *AgentBuilder {
	b.agent.PID = pid
	return b
}

// WithType sets the agent type.
func (b *AgentBuilder) WithType(t string) *AgentBuilder {
	b.agent.Type = t
	return b
}

// WithProcessName sets the process name.
func (b *AgentBuilder) WithProcessName(name string) *AgentBuilder {
	b.agent.ProcessName = name
	return b
}

// WithStartTime sets the start time.
func (b *AgentBuilder) WithStartTime(t time.Time) *AgentBuilder {
	b.agent.StartTime = t
	return b
}

// WithProjectPath sets the project path.
func (b *AgentBuilder) WithProjectPath(path string) *AgentBuilder {
	b.agent.ProjectPath = path
	return b
}

// WithCommand sets the full command line.
func (b *AgentBuilder) WithCommand(cmd string) *AgentBuilder {
	b.agent.Command = cmd
	return b
}

// Build returns the constructed agent.
func (b *AgentBuilder) Build() types.Agent {
	return b.agent
}

// CommitBuilder provides a fluent interface for creating test commits.
type CommitBuilder struct {
	commit types.Commit
}

// NewCommit creates a new CommitBuilder with defaults.
func NewCommit() *CommitBuilder {
	return &CommitBuilder{
		commit: types.Commit{
			Hash:      "abc123def456",
			Author:    "Test User",
			Email:     "test@example.com",
			Message:   "Test commit message",
			Timestamp: time.Now(),
			IsAI:      false,
		},
	}
}

// WithHash sets the commit hash.
func (b *CommitBuilder) WithHash(hash string) *CommitBuilder {
	b.commit.Hash = hash
	return b
}

// WithAuthor sets the commit author.
func (b *CommitBuilder) WithAuthor(author string) *CommitBuilder {
	b.commit.Author = author
	return b
}

// WithEmail sets the author email.
func (b *CommitBuilder) WithEmail(email string) *CommitBuilder {
	b.commit.Email = email
	return b
}

// WithMessage sets the commit message.
func (b *CommitBuilder) WithMessage(msg string) *CommitBuilder {
	b.commit.Message = msg
	return b
}

// WithTimestamp sets the commit timestamp.
func (b *CommitBuilder) WithTimestamp(t time.Time) *CommitBuilder {
	b.commit.Timestamp = t
	return b
}

// WithIsAI sets whether this is an AI commit.
func (b *CommitBuilder) WithIsAI(isAI bool) *CommitBuilder {
	b.commit.IsAI = isAI
	return b
}

// Build returns the constructed commit.
func (b *CommitBuilder) Build() types.Commit {
	return b.commit
}

// ActivityBuilder provides a fluent interface for creating test activities.
type ActivityBuilder struct {
	activity types.Activity
}

// NewActivity creates a new ActivityBuilder with defaults.
func NewActivity() *ActivityBuilder {
	return &ActivityBuilder{
		activity: types.Activity{
			ID:        1,
			Timestamp: time.Now(),
			Project:   "/test/project",
			Type:      types.ActivityAgentStart,
		},
	}
}

// WithID sets the activity ID.
func (b *ActivityBuilder) WithID(id int64) *ActivityBuilder {
	b.activity.ID = id
	return b
}

// WithTimestamp sets the activity timestamp.
func (b *ActivityBuilder) WithTimestamp(t time.Time) *ActivityBuilder {
	b.activity.Timestamp = t
	return b
}

// WithProject sets the project path.
func (b *ActivityBuilder) WithProject(project string) *ActivityBuilder {
	b.activity.Project = project
	return b
}

// WithType sets the activity type.
func (b *ActivityBuilder) WithType(actType string) *ActivityBuilder {
	b.activity.Type = actType
	return b
}

// WithAgentType sets the agent type.
func (b *ActivityBuilder) WithAgentType(agentType string) *ActivityBuilder {
	b.activity.AgentType = agentType
	return b
}

// WithData sets the activity data.
func (b *ActivityBuilder) WithData(data string) *ActivityBuilder {
	b.activity.Data = data
	return b
}

// Build returns the constructed activity.
func (b *ActivityBuilder) Build() types.Activity {
	return b.activity
}

// AssertEqual is a simple equality assertion helper.
func AssertEqual[T comparable](t *testing.T, got, want T, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertTrue asserts that a condition is true.
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true", msg)
	}
}

// AssertFalse asserts that a condition is false.
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false", msg)
	}
}

// AssertNoError asserts that an error is nil.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertError asserts that an error is not nil.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got nil", msg)
	}
}

// AssertLen asserts the length of a slice.
func AssertLen[T any](t *testing.T, slice []T, expected int, msg string) {
	t.Helper()
	if len(slice) != expected {
		t.Errorf("%s: got length %d, want %d", msg, len(slice), expected)
	}
}

// DaysAgo returns a time.Time that is n days ago.
func DaysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n)
}

// HoursAgo returns a time.Time that is n hours ago.
func HoursAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * time.Hour)
}

// MinutesAgo returns a time.Time that is n minutes ago.
func MinutesAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * time.Minute)
}
