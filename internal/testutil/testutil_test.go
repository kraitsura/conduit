package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// TestTempDir verifies temporary directory creation and cleanup.
func TestTempDir(t *testing.T) {
	t.Run("creates directory", func(t *testing.T) {
		dir := TempDir(t)

		if dir == "" {
			t.Fatal("TempDir returned empty string")
		}

		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}
	})

	// Note: Cleanup happens after the test, so we can't verify it here
	// In a real scenario, the t.Cleanup() call handles removal
}

// TestTempFile verifies temporary file creation.
func TestTempFile(t *testing.T) {
	dir := TempDir(t)

	path := TempFile(t, dir, "test.txt", "hello world")

	if path == "" {
		t.Fatal("TempFile returned empty path")
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("content: got %q, want %q", string(content), "hello world")
	}
}

// TestMkdirAll verifies directory tree creation.
func TestMkdirAll(t *testing.T) {
	dir := TempDir(t)
	nested := filepath.Join(dir, "a", "b", "c")

	MkdirAll(t, nested)

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}
}

// TestCreateGitRepo verifies git repository creation.
func TestCreateGitRepo(t *testing.T) {
	dir := TempDir(t)

	repoPath := CreateGitRepo(t, dir, "my-repo")

	// Verify repo directory exists
	info, err := os.Stat(repoPath)
	if err != nil {
		t.Fatalf("repo directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("repo path is not a directory")
	}

	// Verify .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	info, err = os.Stat(gitDir)
	if err != nil {
		t.Fatalf(".git directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error(".git is not a directory")
	}
}

// TestProjectBuilder verifies project builder pattern.
func TestProjectBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		p := NewProject().Build()

		if p.Path == "" {
			t.Error("default Path should not be empty")
		}
		if p.Name == "" {
			t.Error("default Name should not be empty")
		}
		if p.LastActivity.IsZero() {
			t.Error("default LastActivity should not be zero")
		}
	})

	t.Run("with custom values", func(t *testing.T) {
		now := time.Now()
		agents := []types.Agent{{PID: 123, Type: "claude"}}
		stats := &types.ProjectStats{CommitsToday: 5}

		p := NewProject().
			WithPath("/custom/path").
			WithName("custom-name").
			WithGitRemote("https://github.com/user/repo.git").
			WithLastActivity(now).
			WithAgents(agents...).
			WithStats(stats).
			Build()

		if p.Path != "/custom/path" {
			t.Errorf("Path: got %q", p.Path)
		}
		if p.Name != "custom-name" {
			t.Errorf("Name: got %q", p.Name)
		}
		if p.GitRemote != "https://github.com/user/repo.git" {
			t.Errorf("GitRemote: got %q", p.GitRemote)
		}
		if !p.LastActivity.Equal(now) {
			t.Errorf("LastActivity: got %v", p.LastActivity)
		}
		if len(p.ActiveAgents) != 1 {
			t.Errorf("ActiveAgents: got %d", len(p.ActiveAgents))
		}
		if p.Stats.CommitsToday != 5 {
			t.Errorf("Stats.CommitsToday: got %d", p.Stats.CommitsToday)
		}
	})

	t.Run("chaining returns new builder", func(t *testing.T) {
		b := NewProject()
		b2 := b.WithPath("/path")

		// Both should point to same builder (fluent API)
		if b != b2 {
			t.Error("WithPath should return same builder instance")
		}
	})
}

// TestAgentBuilder verifies agent builder pattern.
func TestAgentBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		a := NewAgent().Build()

		if a.PID == 0 {
			t.Error("default PID should not be 0")
		}
		if a.Type == "" {
			t.Error("default Type should not be empty")
		}
		if a.StartTime.IsZero() {
			t.Error("default StartTime should not be zero")
		}
	})

	t.Run("with custom values", func(t *testing.T) {
		now := time.Now()

		a := NewAgent().
			WithPID(5678).
			WithType("cursor").
			WithProcessName("Cursor").
			WithStartTime(now).
			WithProjectPath("/my/project").
			WithCommand("/usr/bin/cursor --flag").
			Build()

		if a.PID != 5678 {
			t.Errorf("PID: got %d", a.PID)
		}
		if a.Type != "cursor" {
			t.Errorf("Type: got %q", a.Type)
		}
		if a.ProcessName != "Cursor" {
			t.Errorf("ProcessName: got %q", a.ProcessName)
		}
		if !a.StartTime.Equal(now) {
			t.Errorf("StartTime: got %v", a.StartTime)
		}
		if a.ProjectPath != "/my/project" {
			t.Errorf("ProjectPath: got %q", a.ProjectPath)
		}
		if a.Command != "/usr/bin/cursor --flag" {
			t.Errorf("Command: got %q", a.Command)
		}
	})
}

// TestCommitBuilder verifies commit builder pattern.
func TestCommitBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		c := NewCommit().Build()

		if c.Hash == "" {
			t.Error("default Hash should not be empty")
		}
		if c.Author == "" {
			t.Error("default Author should not be empty")
		}
		if c.Email == "" {
			t.Error("default Email should not be empty")
		}
		if c.IsAI {
			t.Error("default IsAI should be false")
		}
	})

	t.Run("with AI commit", func(t *testing.T) {
		c := NewCommit().
			WithAuthor("Claude").
			WithEmail("noreply@anthropic.com").
			WithMessage("fix: bug").
			WithIsAI(true).
			Build()

		if c.Author != "Claude" {
			t.Errorf("Author: got %q", c.Author)
		}
		if !c.IsAI {
			t.Error("IsAI should be true")
		}
	})
}

// TestActivityBuilder verifies activity builder pattern.
func TestActivityBuilder(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		a := NewActivity().Build()

		if a.ID == 0 {
			t.Error("default ID should not be 0")
		}
		if a.Project == "" {
			t.Error("default Project should not be empty")
		}
		if a.Type == "" {
			t.Error("default Type should not be empty")
		}
	})

	t.Run("with custom values", func(t *testing.T) {
		now := time.Now()

		a := NewActivity().
			WithID(42).
			WithTimestamp(now).
			WithProject("/my/project").
			WithType(types.ActivityCommit).
			WithAgentType("claude").
			WithData(`{"hash": "abc123"}`).
			Build()

		if a.ID != 42 {
			t.Errorf("ID: got %d", a.ID)
		}
		if a.Type != types.ActivityCommit {
			t.Errorf("Type: got %q", a.Type)
		}
		if a.AgentType != "claude" {
			t.Errorf("AgentType: got %q", a.AgentType)
		}
		if a.Data != `{"hash": "abc123"}` {
			t.Errorf("Data: got %q", a.Data)
		}
	})
}

// TestAssertions verifies assertion helpers.
func TestAssertions(t *testing.T) {
	// These tests use a fake *testing.T to capture failures
	// In real usage, failures would be reported to the actual test

	t.Run("AssertEqual", func(t *testing.T) {
		AssertEqual(t, 1, 1, "integers should be equal")
		AssertEqual(t, "hello", "hello", "strings should be equal")
		AssertEqual(t, true, true, "bools should be equal")
	})

	t.Run("AssertTrue", func(t *testing.T) {
		AssertTrue(t, true, "true is true")
		AssertTrue(t, 1 == 1, "1 equals 1")
	})

	t.Run("AssertFalse", func(t *testing.T) {
		AssertFalse(t, false, "false is false")
		AssertFalse(t, 1 == 2, "1 does not equal 2")
	})

	t.Run("AssertNoError", func(t *testing.T) {
		AssertNoError(t, nil, "nil is not an error")
	})

	t.Run("AssertLen", func(t *testing.T) {
		AssertLen(t, []int{1, 2, 3}, 3, "slice has 3 elements")
		AssertLen(t, []string{}, 0, "empty slice has 0 elements")
	})
}

// TestTimeHelpers verifies time utility functions.
func TestTimeHelpers(t *testing.T) {
	now := time.Now()

	t.Run("DaysAgo", func(t *testing.T) {
		result := DaysAgo(7)
		expected := now.AddDate(0, 0, -7)

		// Allow 1 second tolerance for test execution time
		diff := result.Sub(expected)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("DaysAgo(7) off by %v", diff)
		}
	})

	t.Run("HoursAgo", func(t *testing.T) {
		result := HoursAgo(24)
		expected := now.Add(-24 * time.Hour)

		diff := result.Sub(expected)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("HoursAgo(24) off by %v", diff)
		}
	})

	t.Run("MinutesAgo", func(t *testing.T) {
		result := MinutesAgo(30)
		expected := now.Add(-30 * time.Minute)

		diff := result.Sub(expected)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("MinutesAgo(30) off by %v", diff)
		}
	})

	t.Run("zero values", func(t *testing.T) {
		d0 := DaysAgo(0)
		h0 := HoursAgo(0)
		m0 := MinutesAgo(0)

		// All should be approximately now
		if d0.Sub(now) > time.Second {
			t.Error("DaysAgo(0) should be ~now")
		}
		if h0.Sub(now) > time.Second {
			t.Error("HoursAgo(0) should be ~now")
		}
		if m0.Sub(now) > time.Second {
			t.Error("MinutesAgo(0) should be ~now")
		}
	})
}

// TestBuilderImmutability verifies builders don't share state incorrectly.
func TestBuilderImmutability(t *testing.T) {
	t.Run("project builder", func(t *testing.T) {
		b1 := NewProject().WithName("project1")
		p1 := b1.Build()

		b1.WithName("project2")
		p2 := b1.Build()

		// p1 should not be affected by later changes
		if p1.Name != "project1" {
			t.Errorf("p1.Name changed: got %q", p1.Name)
		}
		if p2.Name != "project2" {
			t.Errorf("p2.Name: got %q", p2.Name)
		}
	})

	t.Run("agent builder", func(t *testing.T) {
		b := NewAgent().WithPID(100)
		a1 := b.Build()

		b.WithPID(200)
		a2 := b.Build()

		if a1.PID != 100 {
			t.Errorf("a1.PID changed: got %d", a1.PID)
		}
		if a2.PID != 200 {
			t.Errorf("a2.PID: got %d", a2.PID)
		}
	})
}

// BenchmarkProjectBuilder measures builder performance.
func BenchmarkProjectBuilder(b *testing.B) {
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewProject().
			WithPath("/path").
			WithName("name").
			WithGitRemote("remote").
			WithLastActivity(now).
			Build()
	}
}

// BenchmarkTempDir measures temp directory creation overhead.
func BenchmarkTempDir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dir, _ := os.MkdirTemp("", "bench-*")
		os.RemoveAll(dir)
	}
}
