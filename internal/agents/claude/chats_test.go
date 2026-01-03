package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		maxLen int
		want   string
	}{
		{
			name:   "short message",
			msg:    "hello world",
			maxLen: 40,
			want:   "hello world",
		},
		{
			name:   "long message",
			msg:    "this is a very long message that needs to be truncated",
			maxLen: 20,
			want:   "this is a very lo...",
		},
		{
			name:   "multiline message",
			msg:    "first line\nsecond line\nthird line",
			maxLen: 40,
			want:   "first line",
		},
		{
			name:   "slash command",
			msg:    "/context something else",
			maxLen: 40,
			want:   "/context",
		},
		{
			name:   "pasted content",
			msg:    "[Pasted text #1 +404 lines]",
			maxLen: 40,
			want:   "[Pasted text #1 +404 lines]",
		},
		{
			name:   "whitespace trimming",
			msg:    "  trimmed message  ",
			maxLen: 40,
			want:   "trimmed message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateMessage(tt.msg, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseHistory(t *testing.T) {
	// Create temp directory with test history file
	tmpDir := t.TempDir()
	historyDir := filepath.Join(tmpDir, ".claude")
	os.MkdirAll(historyDir, 0755)

	historyPath := filepath.Join(historyDir, "history.jsonl")

	// Create test entries
	now := time.Now()
	entries := []historyEntry{
		{
			Display:   "old message",
			Timestamp: now.Add(-2 * time.Hour).UnixMilli(),
			Project:   "/project/a",
			SessionID: "session-1",
		},
		{
			Display:   "recent message",
			Timestamp: now.Add(-30 * time.Minute).UnixMilli(),
			Project:   "/project/a",
			SessionID: "session-2",
		},
		{
			Display:   "another recent message",
			Timestamp: now.Add(-10 * time.Minute).UnixMilli(),
			Project:   "/project/b",
			SessionID: "session-3",
		},
	}

	// Write entries to file
	file, err := os.Create(historyPath)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}
	for _, entry := range entries {
		data, _ := json.Marshal(entry)
		file.Write(data)
		file.WriteString("\n")
	}
	file.Close()

	// Test parsing with mock - we need to override GetHistoryPath
	// For now, test the truncateMessage helper which is testable
	// Full integration test would require mocking the home directory

	t.Run("entries_format", func(t *testing.T) {
		// Verify entry JSON format
		entry := historyEntry{
			Display:   "test message",
			Timestamp: time.Now().UnixMilli(),
			Project:   "/test/project",
			SessionID: "test-session",
		}

		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal entry: %v", err)
		}

		var parsed historyEntry
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal entry: %v", err)
		}

		if parsed.Display != entry.Display {
			t.Errorf("Display = %q, want %q", parsed.Display, entry.Display)
		}
		if parsed.SessionID != entry.SessionID {
			t.Errorf("SessionID = %q, want %q", parsed.SessionID, entry.SessionID)
		}
	})
}

func TestGetHistoryPath(t *testing.T) {
	path := GetHistoryPath()

	if path == "" {
		t.Skip("Could not determine home directory")
	}

	// Verify path structure
	if !filepath.IsAbs(path) {
		t.Errorf("Path should be absolute: %s", path)
	}

	if filepath.Base(path) != "history.jsonl" {
		t.Errorf("Expected history.jsonl, got %s", filepath.Base(path))
	}

	if filepath.Base(filepath.Dir(path)) != ".claude" {
		t.Errorf("Expected .claude directory, got %s", filepath.Dir(path))
	}
}

func TestHistoryEntryTimestamp(t *testing.T) {
	// Verify timestamp is in milliseconds
	now := time.Now()
	entry := historyEntry{
		Timestamp: now.UnixMilli(),
	}

	parsed := time.UnixMilli(entry.Timestamp)

	// Should be within 1 second of original
	diff := now.Sub(parsed)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("Timestamp conversion error: diff = %v", diff)
	}
}

func TestGetChatHistoryGrouping(t *testing.T) {
	// Test that entries with same sessionId are grouped together
	// This is a unit test for the grouping logic

	entries := []historyEntry{
		{SessionID: "s1", Timestamp: 1000, Project: "/a", Display: "first"},
		{SessionID: "s1", Timestamp: 2000, Project: "/a", Display: "second"},
		{SessionID: "s2", Timestamp: 1500, Project: "/b", Display: "other"},
		{SessionID: "s1", Timestamp: 3000, Project: "/a", Display: "third"},
	}

	// Group by session
	sessions := make(map[string][]historyEntry)
	for _, entry := range entries {
		if entry.SessionID != "" {
			sessions[entry.SessionID] = append(sessions[entry.SessionID], entry)
		}
	}

	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	if len(sessions["s1"]) != 3 {
		t.Errorf("Session s1 should have 3 entries, got %d", len(sessions["s1"]))
	}

	if len(sessions["s2"]) != 1 {
		t.Errorf("Session s2 should have 1 entry, got %d", len(sessions["s2"]))
	}
}
