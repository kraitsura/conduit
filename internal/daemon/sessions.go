package daemon

import (
	"path/filepath"
	"time"

	"github.com/kraitsura/conduit/internal/agents/claude"
	"github.com/kraitsura/conduit/internal/session"
	"github.com/kraitsura/conduit/internal/types"
)

// SessionsRequest extends Request for session queries
type SessionsRequest struct {
	Request
	Since      int64 `json:"since,omitempty"`      // Unix timestamp
	Until      int64 `json:"until,omitempty"`      // Unix timestamp
	ActiveOnly bool  `json:"active_only,omitempty"`
}

// SessionsResponse extends Response for session data
type SessionsResponse struct {
	Response
	Sessions       []types.ConduitSession `json:"sessions,omitempty"`
	CurrentSession *types.ConduitSession  `json:"current_session,omitempty"`
	ProjectSummary *types.ProjectSummary  `json:"project_summary,omitempty"`
}

// ensureActiveSession creates or returns the active session for a project
func (d *Daemon) ensureActiveSession(projectPath string) *types.ConduitSession {
	// Check if there's already an active session
	currentSession, err := d.store.GetCurrentSession(projectPath)
	if err == nil && currentSession != nil {
		return currentSession
	}

	// Create new session
	newSession := &types.ConduitSession{
		ProjectPath: projectPath,
		ProjectName: filepath.Base(projectPath),
		StartTime:   time.Now(),
		IsActive:    true,
	}

	if err := d.store.SaveConduitSession(newSession); err != nil {
		return nil
	}

	return newSession
}

// checkSessionTimeouts closes sessions that have been idle for 30+ minutes
func (d *Daemon) checkSessionTimeouts() {
	// Get all active sessions
	sessions, err := d.store.GetConduitSessions(types.SessionFilter{
		ActiveOnly: true,
	})
	if err != nil {
		return
	}

	now := time.Now()
	for _, session := range sessions {
		// Check if session has any active agents
		hasActiveAgents := false
		for _, a := range d.agents {
			if a.ProjectPath == session.ProjectPath {
				hasActiveAgents = true
				break
			}
		}

		if !hasActiveAgents {
			// Check if last activity was > 30 min ago
			lastActivity := session.StartTime
			chats, _ := d.store.GetChatsForSession(session.ID)
			for _, chat := range chats {
				if chat.EndTime != nil && chat.EndTime.After(lastActivity) {
					lastActivity = *chat.EndTime
				} else if chat.StartTime.After(lastActivity) {
					lastActivity = chat.StartTime
				}
			}

			if now.Sub(lastActivity) > types.SessionGap {
				d.store.EndConduitSession(session.ID)
			}
		}
	}
}

// syncClaudeChats updates chat records from Claude history
func (d *Daemon) syncClaudeChats() {
	since := time.Now().Add(-24 * time.Hour) // Look back 24 hours

	chats, err := claude.GetChatHistory("", since)
	if err != nil {
		return
	}

	for _, chat := range chats {
		// Check if this chat is already in the database
		existing, _ := d.store.GetAgentChats(chat.ProjectPath, chat.StartTime.Add(-time.Minute), 1)
		if len(existing) > 0 {
			// Update existing
			for _, e := range existing {
				if e.ID == chat.ID {
					continue // Already have this one
				}
			}
		}

		// Ensure there's a session for this chat's project
		currentSession, _ := d.store.GetCurrentSession(chat.ProjectPath)
		if currentSession == nil {
			// Try to find or create a session that covers this chat's time
			sessions, _ := d.store.GetConduitSessions(types.SessionFilter{
				ProjectPath: chat.ProjectPath,
				Since:       chat.StartTime.Add(-types.SessionGap),
				Limit:       1,
			})
			if len(sessions) == 0 {
				// Create a new session for this chat
				newSession := &types.ConduitSession{
					ProjectPath: chat.ProjectPath,
					ProjectName: filepath.Base(chat.ProjectPath),
					StartTime:   chat.StartTime,
					IsActive:    chat.IsActive,
				}
				d.store.SaveConduitSession(newSession)
			}
		}

		// Save the chat
		d.store.SaveAgentChat(&chat)
	}
}

// handleCurrentSession returns the active session for a project
func (d *Daemon) handleCurrentSession(projectPath string) SessionsResponse {
	currentSession, err := d.store.GetCurrentSession(projectPath)
	if err != nil {
		return SessionsResponse{Response: Response{Status: "error", Error: err.Error()}}
	}

	// Enrich with chats
	if currentSession != nil {
		chats, _ := d.store.GetChatsForSession(currentSession.ID)
		currentSession.Chats = chats
	}

	return SessionsResponse{
		Response:       Response{Status: "ok"},
		CurrentSession: currentSession,
	}
}

// handleSessions returns session history
func (d *Daemon) handleSessions(projectPath string, since, until int64, activeOnly bool, limit int) SessionsResponse {
	filter := types.SessionFilter{
		ProjectPath: projectPath,
		ActiveOnly:  activeOnly,
		Limit:       limit,
	}

	if since > 0 {
		filter.Since = time.Unix(since, 0)
	}
	if until > 0 {
		filter.Until = time.Unix(until, 0)
	}

	sessions, err := d.store.GetSessionsWithChats(filter)
	if err != nil {
		return SessionsResponse{Response: Response{Status: "error", Error: err.Error()}}
	}

	return SessionsResponse{
		Response: Response{Status: "ok"},
		Sessions: sessions,
	}
}

// handleProjectSummary returns an enriched project summary
func (d *Daemon) handleProjectSummary(projectPath string) SessionsResponse {
	summary := types.ProjectSummary{
		Path: projectPath,
		Name: filepath.Base(projectPath),
	}

	// Get active agents for this project
	for _, a := range d.agents {
		if a.ProjectPath == projectPath {
			summary.ActiveAgents = append(summary.ActiveAgents, types.AgentChat{
				AgentType:   a.Type,
				ProjectPath: a.ProjectPath,
				StartTime:   a.StartTime,
				IsActive:    true,
			})
		}
	}

	// Get current session
	currentSession, _ := d.store.GetCurrentSession(projectPath)
	if currentSession != nil {
		chats, _ := d.store.GetChatsForSession(currentSession.ID)
		currentSession.Chats = chats
		summary.CurrentSession = currentSession
	}

	// Get today's stats
	today := time.Now().Truncate(24 * time.Hour)
	todaySessions, _ := d.store.GetConduitSessions(types.SessionFilter{
		ProjectPath: projectPath,
		Since:       today,
	})

	summary.SessionsToday = len(todaySessions)

	// Calculate agent time today
	for _, sess := range todaySessions {
		summary.AgentTimeToday += sess.Duration()
	}

	// Get today's activities for commit count
	activities, _ := d.store.GetActivities(projectPath, 100)
	for _, a := range activities {
		if a.Timestamp.After(today) && a.Type == types.ActivityCommit {
			summary.CommitsToday++
		}
	}

	// Get last activity time
	if len(activities) > 0 {
		summary.LastActivity = activities[0].Timestamp
	}

	return SessionsResponse{
		Response:       Response{Status: "ok"},
		ProjectSummary: &summary,
	}
}

// BuildProjectSummaries creates summaries for all tracked projects
func (d *Daemon) BuildProjectSummaries() []types.ProjectSummary {
	var summaries []types.ProjectSummary

	for _, p := range d.projects {
		resp := d.handleProjectSummary(p.Path)
		if resp.ProjectSummary != nil {
			summaries = append(summaries, *resp.ProjectSummary)
		}
	}

	return summaries
}

// GetAllActiveSessions returns active sessions across all projects
func (d *Daemon) GetAllActiveSessions() []types.ConduitSession {
	sessions, err := d.store.GetConduitSessions(types.SessionFilter{
		ActiveOnly: true,
	})
	if err != nil {
		return nil
	}

	// Enrich with chats
	for i := range sessions {
		chats, _ := d.store.GetChatsForSession(sessions[i].ID)
		sessions[i].Chats = chats
	}

	return sessions
}

// GetRecentSessionsAllProjects returns recent sessions across all projects
func (d *Daemon) GetRecentSessionsAllProjects(limit int) []types.ConduitSession {
	sessions, err := d.store.GetConduitSessions(types.SessionFilter{
		Limit: limit,
	})
	if err != nil {
		return nil
	}

	return sessions
}

// ComputeGlobalStats returns aggregate stats across all projects
func (d *Daemon) ComputeGlobalStats() session.SessionStats {
	// Get all sessions from today
	today := time.Now().Truncate(24 * time.Hour)
	sessions, _ := d.store.GetConduitSessions(types.SessionFilter{
		Since: today,
	})

	return session.GetSessionStats(sessions)
}
