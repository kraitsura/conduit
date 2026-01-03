package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// historyEntry represents a single entry in ~/.claude/history.jsonl
type historyEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"` // milliseconds
	Project   string `json:"project"`
	SessionID string `json:"sessionId"`
}

// GetHistoryPath returns the path to Claude's history file
func GetHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "history.jsonl")
}

// ParseHistory reads and parses the Claude history file
func ParseHistory(since time.Time) ([]historyEntry, error) {
	historyPath := GetHistoryPath()
	if historyPath == "" {
		return nil, nil
	}

	file, err := os.Open(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	sinceMs := since.UnixMilli()
	var entries []historyEntry

	scanner := bufio.NewScanner(file)
	// Increase buffer size for potentially long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var entry historyEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed entries
		}

		// Filter by time
		if entry.Timestamp < sinceMs {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

// GetChatHistory returns chat history grouped by session
func GetChatHistory(projectPath string, since time.Time) ([]types.AgentChat, error) {
	entries, err := ParseHistory(since)
	if err != nil {
		return nil, err
	}

	// Group entries by sessionId
	sessions := make(map[string][]historyEntry)
	for _, entry := range entries {
		// Filter by project if specified
		if projectPath != "" && entry.Project != projectPath {
			continue
		}
		if entry.SessionID != "" {
			sessions[entry.SessionID] = append(sessions[entry.SessionID], entry)
		}
	}

	var chats []types.AgentChat
	for sessionID, entries := range sessions {
		if len(entries) == 0 {
			continue
		}

		// Sort entries by timestamp
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp < entries[j].Timestamp
		})

		// Create chat from session
		chat := types.AgentChat{
			ID:           sessionID,
			AgentType:    "claude",
			ProjectPath:  entries[0].Project,
			StartTime:    time.UnixMilli(entries[0].Timestamp),
			MessageCount: len(entries),
		}

		// Use first message as name (truncated)
		chat.Name = truncateMessage(entries[0].Display, 40)

		// Set end time from last entry
		lastTime := time.UnixMilli(entries[len(entries)-1].Timestamp)
		chat.EndTime = &lastTime

		chats = append(chats, chat)
	}

	// Sort by start time descending (most recent first)
	sort.Slice(chats, func(i, j int) bool {
		return chats[i].StartTime.After(chats[j].StartTime)
	})

	return chats, nil
}

// GetActiveChats returns currently active Claude chats by checking running processes
// and correlating with recent history
func GetActiveChats(projectPath string, runningPIDs []int) ([]types.AgentChat, error) {
	// Get recent history (last 2 hours)
	since := time.Now().Add(-2 * time.Hour)
	entries, err := ParseHistory(since)
	if err != nil {
		return nil, err
	}

	// Filter by project
	var projectEntries []historyEntry
	for _, entry := range entries {
		if projectPath == "" || entry.Project == projectPath {
			projectEntries = append(projectEntries, entry)
		}
	}

	// Group by sessionId
	sessions := make(map[string][]historyEntry)
	for _, entry := range projectEntries {
		if entry.SessionID != "" {
			sessions[entry.SessionID] = append(sessions[entry.SessionID], entry)
		}
	}

	// Find sessions with recent activity (within 30 minutes)
	recentCutoff := time.Now().Add(-30 * time.Minute).UnixMilli()

	var activeChats []types.AgentChat
	for sessionID, entries := range sessions {
		if len(entries) == 0 {
			continue
		}

		// Sort by timestamp
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp < entries[j].Timestamp
		})

		lastEntry := entries[len(entries)-1]

		// Check if session has recent activity
		if lastEntry.Timestamp < recentCutoff {
			continue
		}

		chat := types.AgentChat{
			ID:           sessionID,
			AgentType:    "claude",
			ProjectPath:  entries[0].Project,
			StartTime:    time.UnixMilli(entries[0].Timestamp),
			MessageCount: len(entries),
			Name:         truncateMessage(entries[0].Display, 40),
			IsActive:     true,
		}

		activeChats = append(activeChats, chat)
	}

	// Sort by start time descending
	sort.Slice(activeChats, func(i, j int) bool {
		return activeChats[i].StartTime.After(activeChats[j].StartTime)
	})

	return activeChats, nil
}

// GetRecentSessions returns unique session IDs from recent history
func GetRecentSessions(since time.Time) (map[string]time.Time, error) {
	entries, err := ParseHistory(since)
	if err != nil {
		return nil, err
	}

	// Map session ID to last activity time
	sessions := make(map[string]time.Time)
	for _, entry := range entries {
		if entry.SessionID == "" {
			continue
		}
		entryTime := time.UnixMilli(entry.Timestamp)
		if existing, ok := sessions[entry.SessionID]; !ok || entryTime.After(existing) {
			sessions[entry.SessionID] = entryTime
		}
	}

	return sessions, nil
}

// truncateMessage truncates a message to maxLen characters
func truncateMessage(msg string, maxLen int) string {
	// Clean up the message
	msg = strings.TrimSpace(msg)

	// Handle pasted content markers
	if strings.HasPrefix(msg, "[Pasted") {
		return msg
	}

	// Handle slash commands
	if strings.HasPrefix(msg, "/") {
		parts := strings.SplitN(msg, " ", 2)
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Take first line only
	if idx := strings.Index(msg, "\n"); idx != -1 {
		msg = msg[:idx]
	}

	// Truncate
	if len(msg) > maxLen {
		return msg[:maxLen-3] + "..."
	}

	return msg
}

// GetProjectsFromHistory returns unique project paths from history
func GetProjectsFromHistory(since time.Time) ([]string, error) {
	entries, err := ParseHistory(since)
	if err != nil {
		return nil, err
	}

	projectSet := make(map[string]bool)
	for _, entry := range entries {
		if entry.Project != "" {
			projectSet[entry.Project] = true
		}
	}

	var projects []string
	for project := range projectSet {
		projects = append(projects, project)
	}

	sort.Strings(projects)
	return projects, nil
}
