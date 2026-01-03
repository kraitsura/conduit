package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

func TestSaveAndGetConduitSession(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Create a test session
	session := &types.ConduitSession{
		ProjectPath: "/test/project",
		ProjectName: "project",
		StartTime:   time.Now().Add(-1 * time.Hour),
		Branch:      "main",
		IsActive:    true,
	}

	// Save session
	if err := store.SaveConduitSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Verify ID was generated
	if session.ID == "" {
		t.Error("Session ID was not generated")
	}

	// Retrieve session
	retrieved, err := store.GetConduitSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved session is nil")
	}

	if retrieved.ProjectPath != session.ProjectPath {
		t.Errorf("ProjectPath = %q, want %q", retrieved.ProjectPath, session.ProjectPath)
	}

	if retrieved.Branch != session.Branch {
		t.Errorf("Branch = %q, want %q", retrieved.Branch, session.Branch)
	}

	if !retrieved.IsActive {
		t.Error("Session should be active")
	}
}

func TestGetConduitSessions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Create test sessions
	now := time.Now()
	sessions := []*types.ConduitSession{
		{
			ProjectPath: "/project/a",
			StartTime:   now.Add(-3 * time.Hour),
		},
		{
			ProjectPath: "/project/a",
			StartTime:   now.Add(-2 * time.Hour),
		},
		{
			ProjectPath: "/project/b",
			StartTime:   now.Add(-1 * time.Hour),
		},
	}

	for _, s := range sessions {
		if err := store.SaveConduitSession(s); err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}
	}

	// Test filter by project
	projectASessions, err := store.GetConduitSessions(types.SessionFilter{
		ProjectPath: "/project/a",
	})
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	if len(projectASessions) != 2 {
		t.Errorf("Expected 2 sessions for project A, got %d", len(projectASessions))
	}

	// Test filter by time
	recentSessions, err := store.GetConduitSessions(types.SessionFilter{
		Since: now.Add(-90 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	if len(recentSessions) != 1 {
		t.Errorf("Expected 1 recent session, got %d", len(recentSessions))
	}

	// Test limit
	limitedSessions, err := store.GetConduitSessions(types.SessionFilter{
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("Failed to get sessions: %v", err)
	}

	if len(limitedSessions) != 2 {
		t.Errorf("Expected 2 sessions with limit, got %d", len(limitedSessions))
	}
}

func TestAgentChatOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Create a test chat
	chat := &types.AgentChat{
		AgentType:    "claude",
		Name:         "Test chat",
		ProjectPath:  "/test/project",
		StartTime:    time.Now().Add(-30 * time.Minute),
		MessageCount: 5,
	}

	// Save chat
	if err := store.SaveAgentChat(chat); err != nil {
		t.Fatalf("Failed to save chat: %v", err)
	}

	if chat.ID == "" {
		t.Error("Chat ID was not generated")
	}

	// Get active chats
	activeChats, err := store.GetActiveChats()
	if err != nil {
		t.Fatalf("Failed to get active chats: %v", err)
	}

	if len(activeChats) != 1 {
		t.Errorf("Expected 1 active chat, got %d", len(activeChats))
	}

	// End chat
	if err := store.EndAgentChat(chat.ID); err != nil {
		t.Fatalf("Failed to end chat: %v", err)
	}

	// Verify no active chats
	activeChats, err = store.GetActiveChats()
	if err != nil {
		t.Fatalf("Failed to get active chats: %v", err)
	}

	if len(activeChats) != 0 {
		t.Errorf("Expected 0 active chats after end, got %d", len(activeChats))
	}
}

func TestSessionCommitLinking(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	// Create a session
	session := &types.ConduitSession{
		ProjectPath: "/test/project",
		StartTime:   time.Now(),
	}
	if err := store.SaveConduitSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Link commits
	commits := []string{"abc123", "def456", "ghi789"}
	for _, hash := range commits {
		if err := store.LinkCommitToSession(session.ID, hash); err != nil {
			t.Fatalf("Failed to link commit: %v", err)
		}
	}

	// Retrieve commits
	retrieved, err := store.GetSessionCommits(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session commits: %v", err)
	}

	if len(retrieved) != len(commits) {
		t.Errorf("Expected %d commits, got %d", len(commits), len(retrieved))
	}

	// Test duplicate handling
	if err := store.LinkCommitToSession(session.ID, "abc123"); err != nil {
		t.Fatalf("Failed to link duplicate commit: %v", err)
	}

	retrieved, _ = store.GetSessionCommits(session.ID)
	if len(retrieved) != len(commits) {
		t.Errorf("Duplicate should be ignored, expected %d commits, got %d", len(commits), len(retrieved))
	}
}

func TestGetCurrentSession(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	defer store.Close()

	projectPath := "/test/project"

	// No active session initially
	current, err := store.GetCurrentSession(projectPath)
	if err != nil {
		t.Fatalf("Failed to get current session: %v", err)
	}
	if current != nil {
		t.Error("Expected no current session initially")
	}

	// Create an active session
	session := &types.ConduitSession{
		ProjectPath: projectPath,
		StartTime:   time.Now(),
	}
	if err := store.SaveConduitSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Should find current session
	current, err = store.GetCurrentSession(projectPath)
	if err != nil {
		t.Fatalf("Failed to get current session: %v", err)
	}
	if current == nil {
		t.Error("Expected to find current session")
	}
	if current.ID != session.ID {
		t.Errorf("Current session ID = %q, want %q", current.ID, session.ID)
	}

	// End the session
	if err := store.EndConduitSession(session.ID); err != nil {
		t.Fatalf("Failed to end session: %v", err)
	}

	// No active session after ending
	current, err = store.GetCurrentSession(projectPath)
	if err != nil {
		t.Fatalf("Failed to get current session: %v", err)
	}
	if current != nil {
		t.Error("Expected no current session after ending")
	}
}

func TestGenerateID(t *testing.T) {
	ids := make(map[string]bool)

	// Generate multiple IDs and verify uniqueness
	for i := 0; i < 100; i++ {
		id := generateID()
		if len(id) != 12 {
			t.Errorf("ID length = %d, want 12", len(id))
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestDatabaseIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open and create schema
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open store: %v", err)
	}
	store.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Reopen and verify schema persists
	store, err = Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen store: %v", err)
	}
	defer store.Close()

	// Should be able to query new tables
	_, err = store.GetConduitSessions(types.SessionFilter{Limit: 1})
	if err != nil {
		t.Errorf("Failed to query conduit_sessions: %v", err)
	}

	_, err = store.GetActiveChats()
	if err != nil {
		t.Errorf("Failed to query agent_chats: %v", err)
	}
}
