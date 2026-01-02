package session

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// DetectSessionBoundaries groups activities into sessions for a project.
// Sessions are bounded by 30-minute gaps in activity.
func DetectSessionBoundaries(activities []types.Activity, projectPath string) []types.ConduitSession {
	// Filter activities for this project
	var projectActivities []types.Activity
	for _, a := range activities {
		if projectPath == "" || a.Project == projectPath {
			projectActivities = append(projectActivities, a)
		}
	}

	if len(projectActivities) == 0 {
		return nil
	}

	// Sort by timestamp ascending
	sort.Slice(projectActivities, func(i, j int) bool {
		return projectActivities[i].Timestamp.Before(projectActivities[j].Timestamp)
	})

	var sessions []types.ConduitSession
	var currentSession *types.ConduitSession
	var lastActivityTime time.Time

	for _, activity := range projectActivities {
		if currentSession == nil {
			// Start new session
			currentSession = &types.ConduitSession{
				ProjectPath: activity.Project,
				ProjectName: filepath.Base(activity.Project),
				StartTime:   activity.Timestamp,
				IsActive:    true,
			}
			lastActivityTime = activity.Timestamp
		} else {
			// Check if gap exceeds session boundary
			gap := activity.Timestamp.Sub(lastActivityTime)
			if gap > types.SessionGap {
				// End current session and start new one
				currentSession.EndTime = &lastActivityTime
				currentSession.IsActive = false
				sessions = append(sessions, *currentSession)

				currentSession = &types.ConduitSession{
					ProjectPath: activity.Project,
					ProjectName: filepath.Base(activity.Project),
					StartTime:   activity.Timestamp,
					IsActive:    true,
				}
			}
			lastActivityTime = activity.Timestamp
		}

		// Update session with activity
		updateSessionWithActivity(currentSession, activity)
	}

	// Add final session
	if currentSession != nil {
		// Check if session is still active (activity within last 30 min)
		if time.Since(lastActivityTime) > types.SessionGap {
			currentSession.EndTime = &lastActivityTime
			currentSession.IsActive = false
		}
		sessions = append(sessions, *currentSession)
	}

	return sessions
}

// getLastActivityTime returns the latest timestamp for a session
func getLastActivityTime(s *types.ConduitSession) *time.Time {
	latest := s.StartTime

	// Check chats
	for _, chat := range s.Chats {
		if chat.EndTime != nil && chat.EndTime.After(latest) {
			latest = *chat.EndTime
		} else if chat.StartTime.After(latest) {
			latest = chat.StartTime
		}
	}

	// Check commits
	for _, commit := range s.Commits {
		if commit.Timestamp.After(latest) {
			latest = commit.Timestamp
		}
	}

	return &latest
}

// updateSessionWithActivity adds activity data to a session
func updateSessionWithActivity(s *types.ConduitSession, a types.Activity) {
	switch a.Type {
	case types.ActivityAgentStart:
		// Add a chat entry for agent start
		chat := types.AgentChat{
			AgentType:   a.AgentType,
			ProjectPath: a.Project,
			StartTime:   a.Timestamp,
			IsActive:    true,
		}
		s.Chats = append(s.Chats, chat)

	case types.ActivityAgentStop:
		// Find and close the matching chat
		for i := len(s.Chats) - 1; i >= 0; i-- {
			if s.Chats[i].AgentType == a.AgentType && s.Chats[i].IsActive {
				s.Chats[i].EndTime = &a.Timestamp
				s.Chats[i].IsActive = false
				break
			}
		}

	case types.ActivityCommit:
		// Add commit reference
		commit := types.CommitRef{
			Timestamp: a.Timestamp,
			// Data field may contain commit info as JSON
		}
		s.Commits = append(s.Commits, commit)
	}
}

// GetCurrentSession returns the active session for a project (if any)
// by looking at recent activities
func GetCurrentSession(activities []types.Activity, projectPath string) *types.ConduitSession {
	// Filter to last 30 minutes
	cutoff := time.Now().Add(-types.SessionGap)

	var recentActivities []types.Activity
	for _, a := range activities {
		if a.Timestamp.After(cutoff) && (projectPath == "" || a.Project == projectPath) {
			recentActivities = append(recentActivities, a)
		}
	}

	if len(recentActivities) == 0 {
		return nil
	}

	sessions := DetectSessionBoundaries(recentActivities, projectPath)
	if len(sessions) == 0 {
		return nil
	}

	// Return the most recent session if it's active
	lastSession := &sessions[len(sessions)-1]
	if lastSession.IsActive {
		return lastSession
	}

	return nil
}

// MergeChatsIntoSessions correlates agent chats with sessions
func MergeChatsIntoSessions(sessions []types.ConduitSession, chats []types.AgentChat) []types.ConduitSession {
	for i := range sessions {
		session := &sessions[i]

		for _, chat := range chats {
			// Check if chat belongs to this session's time range
			if chat.ProjectPath != session.ProjectPath {
				continue
			}

			// Chat starts within session time range
			sessionEnd := session.StartTime.Add(session.Duration())
			if session.EndTime != nil {
				sessionEnd = *session.EndTime
			}

			if chat.StartTime.After(session.StartTime.Add(-time.Minute)) &&
				chat.StartTime.Before(sessionEnd.Add(time.Minute)) {
				// Avoid duplicates
				found := false
				for _, existing := range session.Chats {
					if existing.ID == chat.ID {
						found = true
						break
					}
				}
				if !found {
					session.Chats = append(session.Chats, chat)
				}
			}
		}

		// Sort chats by start time
		sort.Slice(session.Chats, func(a, b int) bool {
			return session.Chats[a].StartTime.Before(session.Chats[b].StartTime)
		})
	}

	return sessions
}

// GroupSessionsByDate groups sessions by calendar date
func GroupSessionsByDate(sessions []types.ConduitSession) map[string][]types.ConduitSession {
	grouped := make(map[string][]types.ConduitSession)

	for _, session := range sessions {
		dateKey := session.StartTime.Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], session)
	}

	return grouped
}

// GetSessionStats calculates aggregate statistics for a set of sessions
func GetSessionStats(sessions []types.ConduitSession) SessionStats {
	stats := SessionStats{
		agentTypes: make(map[string]int),
	}

	for _, session := range sessions {
		stats.SessionCount++
		stats.TotalDuration += session.Duration()
		stats.TotalChatTime += session.TotalChatTime()
		stats.CommitCount += len(session.Commits)

		// Count AI commits
		for _, commit := range session.Commits {
			if commit.IsAI {
				stats.AICommitCount++
			}
		}

		// Track unique agents
		for _, chat := range session.Chats {
			stats.agentTypes[chat.AgentType]++
		}
	}

	return stats
}

// SessionStats holds aggregate session statistics
type SessionStats struct {
	SessionCount   int
	TotalDuration  time.Duration
	TotalChatTime  time.Duration
	CommitCount    int
	AICommitCount  int
	agentTypes     map[string]int
}

// init initializes the agent types map
func init() {
	// This will be called when package is imported
}

// AgentCounts returns a map of agent type to usage count
func (s *SessionStats) AgentCounts() map[string]int {
	if s.agentTypes == nil {
		return make(map[string]int)
	}
	return s.agentTypes
}
