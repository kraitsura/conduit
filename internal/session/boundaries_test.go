package session

import (
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

func TestDetectSessionBoundaries(t *testing.T) {
	now := time.Now()
	projectPath := "/test/project"

	tests := []struct {
		name             string
		activities       []types.Activity
		wantSessionCount int
	}{
		{
			name:             "no activities",
			activities:       nil,
			wantSessionCount: 0,
		},
		{
			name: "single activity",
			activities: []types.Activity{
				{Timestamp: now, Project: projectPath, Type: types.ActivityAgentStart},
			},
			wantSessionCount: 1,
		},
		{
			name: "activities within gap",
			activities: []types.Activity{
				{Timestamp: now.Add(-20 * time.Minute), Project: projectPath, Type: types.ActivityAgentStart},
				{Timestamp: now.Add(-10 * time.Minute), Project: projectPath, Type: types.ActivityCommit},
				{Timestamp: now, Project: projectPath, Type: types.ActivityAgentStop},
			},
			wantSessionCount: 1,
		},
		{
			name: "activities with gap creates two sessions",
			activities: []types.Activity{
				{Timestamp: now.Add(-2 * time.Hour), Project: projectPath, Type: types.ActivityAgentStart},
				{Timestamp: now.Add(-90 * time.Minute), Project: projectPath, Type: types.ActivityAgentStop},
				// Gap of 60 minutes
				{Timestamp: now.Add(-30 * time.Minute), Project: projectPath, Type: types.ActivityAgentStart},
				{Timestamp: now, Project: projectPath, Type: types.ActivityAgentStop},
			},
			wantSessionCount: 2,
		},
		{
			name: "multiple gaps create multiple sessions",
			activities: []types.Activity{
				// Session 1: 10 min of activity
				{Timestamp: now.Add(-4 * time.Hour), Project: projectPath, Type: types.ActivityAgentStart},
				{Timestamp: now.Add(-4*time.Hour + 10*time.Minute), Project: projectPath, Type: types.ActivityAgentStop},
				// Gap of 50 min (> 30min)
				{Timestamp: now.Add(-3 * time.Hour), Project: projectPath, Type: types.ActivityAgentStart},
				{Timestamp: now.Add(-3*time.Hour + 15*time.Minute), Project: projectPath, Type: types.ActivityAgentStop},
				// Gap of 45 min (> 30min)
				{Timestamp: now.Add(-2 * time.Hour), Project: projectPath, Type: types.ActivityAgentStart},
			},
			wantSessionCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := DetectSessionBoundaries(tt.activities, projectPath)
			if len(sessions) != tt.wantSessionCount {
				t.Errorf("DetectSessionBoundaries() returned %d sessions, want %d",
					len(sessions), tt.wantSessionCount)
			}
		})
	}
}

func TestDetectSessionBoundariesFiltersProject(t *testing.T) {
	now := time.Now()

	activities := []types.Activity{
		{Timestamp: now.Add(-20 * time.Minute), Project: "/project/a", Type: types.ActivityAgentStart},
		{Timestamp: now.Add(-15 * time.Minute), Project: "/project/b", Type: types.ActivityAgentStart},
		{Timestamp: now.Add(-10 * time.Minute), Project: "/project/a", Type: types.ActivityCommit},
		{Timestamp: now, Project: "/project/b", Type: types.ActivityAgentStop},
	}

	// Filter to project A
	sessionsA := DetectSessionBoundaries(activities, "/project/a")
	if len(sessionsA) != 1 {
		t.Errorf("Expected 1 session for project A, got %d", len(sessionsA))
	}
	if sessionsA[0].ProjectPath != "/project/a" {
		t.Errorf("Session project = %q, want /project/a", sessionsA[0].ProjectPath)
	}

	// Filter to project B
	sessionsB := DetectSessionBoundaries(activities, "/project/b")
	if len(sessionsB) != 1 {
		t.Errorf("Expected 1 session for project B, got %d", len(sessionsB))
	}

	// No filter
	allSessions := DetectSessionBoundaries(activities, "")
	// Activities are interleaved, so they form one session
	if len(allSessions) != 1 {
		t.Errorf("Expected 1 session with no filter, got %d", len(allSessions))
	}
}

func TestSessionActivityTracking(t *testing.T) {
	now := time.Now()
	projectPath := "/test/project"

	activities := []types.Activity{
		{Timestamp: now.Add(-30 * time.Minute), Project: projectPath, Type: types.ActivityAgentStart, AgentType: "claude"},
		{Timestamp: now.Add(-20 * time.Minute), Project: projectPath, Type: types.ActivityCommit},
		{Timestamp: now.Add(-10 * time.Minute), Project: projectPath, Type: types.ActivityAgentStart, AgentType: "cursor"},
		{Timestamp: now, Project: projectPath, Type: types.ActivityAgentStop, AgentType: "claude"},
	}

	sessions := DetectSessionBoundaries(activities, projectPath)
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 session, got %d", len(sessions))
	}

	session := sessions[0]

	// Should have 2 chats
	if len(session.Chats) != 2 {
		t.Errorf("Expected 2 chats, got %d", len(session.Chats))
	}

	// Should have 1 commit
	if len(session.Commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(session.Commits))
	}

	// Claude chat should be closed
	var claudeChat *types.AgentChat
	for i := range session.Chats {
		if session.Chats[i].AgentType == "claude" {
			claudeChat = &session.Chats[i]
			break
		}
	}
	if claudeChat == nil {
		t.Fatal("Claude chat not found")
	}
	if claudeChat.IsActive {
		t.Error("Claude chat should be inactive (stopped)")
	}

	// Cursor chat should still be active
	var cursorChat *types.AgentChat
	for i := range session.Chats {
		if session.Chats[i].AgentType == "cursor" {
			cursorChat = &session.Chats[i]
			break
		}
	}
	if cursorChat == nil {
		t.Fatal("Cursor chat not found")
	}
	if !cursorChat.IsActive {
		t.Error("Cursor chat should be active")
	}
}

func TestGetCurrentSession(t *testing.T) {
	now := time.Now()
	projectPath := "/test/project"

	t.Run("active session exists", func(t *testing.T) {
		activities := []types.Activity{
			{Timestamp: now.Add(-10 * time.Minute), Project: projectPath, Type: types.ActivityAgentStart},
		}

		current := GetCurrentSession(activities, projectPath)
		if current == nil {
			t.Error("Expected to find current session")
		}
		if !current.IsActive {
			t.Error("Current session should be active")
		}
	})

	t.Run("no recent activity", func(t *testing.T) {
		activities := []types.Activity{
			{Timestamp: now.Add(-2 * time.Hour), Project: projectPath, Type: types.ActivityAgentStart},
			{Timestamp: now.Add(-90 * time.Minute), Project: projectPath, Type: types.ActivityAgentStop},
		}

		current := GetCurrentSession(activities, projectPath)
		if current != nil {
			t.Error("Expected no current session for old activities")
		}
	})

	t.Run("wrong project", func(t *testing.T) {
		activities := []types.Activity{
			{Timestamp: now.Add(-10 * time.Minute), Project: "/other/project", Type: types.ActivityAgentStart},
		}

		current := GetCurrentSession(activities, projectPath)
		if current != nil {
			t.Error("Expected no current session for wrong project")
		}
	})
}

func TestGroupSessionsByDate(t *testing.T) {
	sessions := []types.ConduitSession{
		{StartTime: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
		{StartTime: time.Date(2025, 1, 1, 14, 0, 0, 0, time.UTC)},
		{StartTime: time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC)},
		{StartTime: time.Date(2025, 1, 3, 11, 0, 0, 0, time.UTC)},
	}

	grouped := GroupSessionsByDate(sessions)

	if len(grouped) != 3 {
		t.Errorf("Expected 3 date groups, got %d", len(grouped))
	}

	if len(grouped["2025-01-01"]) != 2 {
		t.Errorf("Expected 2 sessions on 2025-01-01, got %d", len(grouped["2025-01-01"]))
	}

	if len(grouped["2025-01-02"]) != 1 {
		t.Errorf("Expected 1 session on 2025-01-02, got %d", len(grouped["2025-01-02"]))
	}
}

func TestMergeChatsIntoSessions(t *testing.T) {
	now := time.Now()
	projectPath := "/test/project"

	sessions := []types.ConduitSession{
		{
			ProjectPath: projectPath,
			StartTime:   now.Add(-2 * time.Hour),
			EndTime:     func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
		},
		{
			ProjectPath: projectPath,
			StartTime:   now.Add(-30 * time.Minute),
			IsActive:    true,
		},
	}

	chats := []types.AgentChat{
		{
			ID:          "chat-1",
			ProjectPath: projectPath,
			StartTime:   now.Add(-90 * time.Minute),
			AgentType:   "claude",
		},
		{
			ID:          "chat-2",
			ProjectPath: projectPath,
			StartTime:   now.Add(-20 * time.Minute),
			AgentType:   "claude",
		},
		{
			ID:          "chat-3",
			ProjectPath: "/other/project",
			StartTime:   now.Add(-15 * time.Minute),
			AgentType:   "cursor",
		},
	}

	merged := MergeChatsIntoSessions(sessions, chats)

	// First session should have chat-1
	if len(merged[0].Chats) != 1 {
		t.Errorf("First session expected 1 chat, got %d", len(merged[0].Chats))
	}

	// Second session should have chat-2 (chat-3 is wrong project)
	if len(merged[1].Chats) != 1 {
		t.Errorf("Second session expected 1 chat, got %d", len(merged[1].Chats))
	}
}

func TestGetSessionStats(t *testing.T) {
	now := time.Now()
	end := now.Add(-30 * time.Minute)

	sessions := []types.ConduitSession{
		{
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   &end,
			Chats: []types.AgentChat{
				{AgentType: "claude", StartTime: now.Add(-2 * time.Hour), EndTime: &end},
			},
			Commits: []types.CommitRef{
				{IsAI: true},
				{IsAI: false},
			},
		},
		{
			StartTime: now.Add(-30 * time.Minute),
			IsActive:  true,
			Chats: []types.AgentChat{
				{AgentType: "cursor", StartTime: now.Add(-30 * time.Minute)},
			},
		},
	}

	stats := GetSessionStats(sessions)

	if stats.SessionCount != 2 {
		t.Errorf("SessionCount = %d, want 2", stats.SessionCount)
	}

	if stats.CommitCount != 2 {
		t.Errorf("CommitCount = %d, want 2", stats.CommitCount)
	}

	if stats.AICommitCount != 1 {
		t.Errorf("AICommitCount = %d, want 1", stats.AICommitCount)
	}
}
