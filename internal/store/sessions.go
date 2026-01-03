package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// generateID creates a short random ID
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)[:12]
}

// SaveConduitSession creates or updates a conduit session
func (s *Store) SaveConduitSession(session *types.ConduitSession) error {
	if session.ID == "" {
		session.ID = generateID()
	}

	var endTime *int64
	if session.EndTime != nil {
		t := session.EndTime.Unix()
		endTime = &t
	}

	_, err := s.db.Exec(`
		INSERT INTO conduit_sessions (id, project_path, start_time, end_time, branch)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			end_time = excluded.end_time,
			branch = excluded.branch
	`, session.ID, session.ProjectPath, session.StartTime.Unix(), endTime, session.Branch)

	return err
}

// GetConduitSession retrieves a session by ID
func (s *Store) GetConduitSession(id string) (*types.ConduitSession, error) {
	row := s.db.QueryRow(`
		SELECT id, project_path, start_time, end_time, branch
		FROM conduit_sessions
		WHERE id = ?
	`, id)

	var session types.ConduitSession
	var startTime int64
	var endTime sql.NullInt64
	var branch sql.NullString

	if err := row.Scan(&session.ID, &session.ProjectPath, &startTime, &endTime, &branch); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	session.StartTime = time.Unix(startTime, 0)
	if endTime.Valid {
		t := time.Unix(endTime.Int64, 0)
		session.EndTime = &t
	}
	if branch.Valid {
		session.Branch = branch.String
	}
	session.IsActive = session.EndTime == nil

	return &session, nil
}

// GetConduitSessions retrieves sessions matching the filter
func (s *Store) GetConduitSessions(filter types.SessionFilter) ([]types.ConduitSession, error) {
	query := `
		SELECT id, project_path, start_time, end_time, branch
		FROM conduit_sessions
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.ProjectPath != "" {
		query += " AND project_path = ?"
		args = append(args, filter.ProjectPath)
	}

	if !filter.Since.IsZero() {
		query += " AND start_time >= ?"
		args = append(args, filter.Since.Unix())
	}

	if !filter.Until.IsZero() {
		query += " AND start_time <= ?"
		args = append(args, filter.Until.Unix())
	}

	if filter.ActiveOnly {
		query += " AND end_time IS NULL"
	}

	query += " ORDER BY start_time DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []types.ConduitSession
	for rows.Next() {
		var session types.ConduitSession
		var startTime int64
		var endTime sql.NullInt64
		var branch sql.NullString

		if err := rows.Scan(&session.ID, &session.ProjectPath, &startTime, &endTime, &branch); err != nil {
			continue
		}

		session.StartTime = time.Unix(startTime, 0)
		if endTime.Valid {
			t := time.Unix(endTime.Int64, 0)
			session.EndTime = &t
		}
		if branch.Valid {
			session.Branch = branch.String
		}
		session.IsActive = session.EndTime == nil
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetCurrentSession returns the active session for a project (if any)
func (s *Store) GetCurrentSession(projectPath string) (*types.ConduitSession, error) {
	sessions, err := s.GetConduitSessions(types.SessionFilter{
		ProjectPath: projectPath,
		ActiveOnly:  true,
		Limit:       1,
	})
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

// EndConduitSession closes an active session
func (s *Store) EndConduitSession(sessionID string) error {
	_, err := s.db.Exec(`
		UPDATE conduit_sessions
		SET end_time = ?
		WHERE id = ? AND end_time IS NULL
	`, time.Now().Unix(), sessionID)
	return err
}

// SaveAgentChat creates or updates an agent chat
func (s *Store) SaveAgentChat(chat *types.AgentChat) error {
	if chat.ID == "" {
		chat.ID = generateID()
	}

	var endTime *int64
	if chat.EndTime != nil {
		t := chat.EndTime.Unix()
		endTime = &t
	}

	// Find current session for this project
	var sessionID *string
	session, _ := s.GetCurrentSession(chat.ProjectPath)
	if session != nil {
		sessionID = &session.ID
	}

	_, err := s.db.Exec(`
		INSERT INTO agent_chats (id, conduit_session_id, agent_type, name, project_path, start_time, end_time, message_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			conduit_session_id = excluded.conduit_session_id,
			name = excluded.name,
			end_time = excluded.end_time,
			message_count = excluded.message_count
	`, chat.ID, sessionID, chat.AgentType, chat.Name, chat.ProjectPath, chat.StartTime.Unix(), endTime, chat.MessageCount)

	return err
}

// GetAgentChats retrieves chats for a project
func (s *Store) GetAgentChats(projectPath string, since time.Time, limit int) ([]types.AgentChat, error) {
	query := `
		SELECT id, conduit_session_id, agent_type, name, project_path, start_time, end_time, message_count
		FROM agent_chats
		WHERE project_path = ?
	`
	args := []interface{}{projectPath}

	if !since.IsZero() {
		query += " AND start_time >= ?"
		args = append(args, since.Unix())
	}

	query += " ORDER BY start_time DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []types.AgentChat
	for rows.Next() {
		var chat types.AgentChat
		var sessionID sql.NullString
		var name sql.NullString
		var startTime int64
		var endTime sql.NullInt64

		if err := rows.Scan(&chat.ID, &sessionID, &chat.AgentType, &name, &chat.ProjectPath, &startTime, &endTime, &chat.MessageCount); err != nil {
			continue
		}

		chat.StartTime = time.Unix(startTime, 0)
		if endTime.Valid {
			t := time.Unix(endTime.Int64, 0)
			chat.EndTime = &t
		}
		if name.Valid {
			chat.Name = name.String
		}
		chat.IsActive = chat.EndTime == nil
		chats = append(chats, chat)
	}

	return chats, nil
}

// GetActiveChats returns all chats without end times
func (s *Store) GetActiveChats() ([]types.AgentChat, error) {
	rows, err := s.db.Query(`
		SELECT id, conduit_session_id, agent_type, name, project_path, start_time, message_count
		FROM agent_chats
		WHERE end_time IS NULL
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []types.AgentChat
	for rows.Next() {
		var chat types.AgentChat
		var sessionID, name sql.NullString
		var startTime int64

		if err := rows.Scan(&chat.ID, &sessionID, &chat.AgentType, &name, &chat.ProjectPath, &startTime, &chat.MessageCount); err != nil {
			continue
		}

		chat.StartTime = time.Unix(startTime, 0)
		if name.Valid {
			chat.Name = name.String
		}
		chat.IsActive = true
		chats = append(chats, chat)
	}

	return chats, nil
}

// EndAgentChat closes an active chat
func (s *Store) EndAgentChat(chatID string) error {
	_, err := s.db.Exec(`
		UPDATE agent_chats
		SET end_time = ?
		WHERE id = ? AND end_time IS NULL
	`, time.Now().Unix(), chatID)
	return err
}

// LinkCommitToSession associates a commit with a session
func (s *Store) LinkCommitToSession(sessionID, commitHash string) error {
	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO session_commits (session_id, commit_hash)
		VALUES (?, ?)
	`, sessionID, commitHash)
	return err
}

// GetSessionCommits retrieves commits linked to a session
func (s *Store) GetSessionCommits(sessionID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT commit_hash FROM session_commits WHERE session_id = ?
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			continue
		}
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// GetSessionsWithChats retrieves sessions with their associated chats populated
func (s *Store) GetSessionsWithChats(filter types.SessionFilter) ([]types.ConduitSession, error) {
	sessions, err := s.GetConduitSessions(filter)
	if err != nil {
		return nil, err
	}

	// Populate chats for each session
	for i := range sessions {
		chats, err := s.GetChatsForSession(sessions[i].ID)
		if err == nil {
			sessions[i].Chats = chats
		}
	}

	return sessions, nil
}

// GetChatsForSession retrieves all chats belonging to a session
func (s *Store) GetChatsForSession(sessionID string) ([]types.AgentChat, error) {
	rows, err := s.db.Query(`
		SELECT id, agent_type, name, project_path, start_time, end_time, message_count
		FROM agent_chats
		WHERE conduit_session_id = ?
		ORDER BY start_time ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []types.AgentChat
	for rows.Next() {
		var chat types.AgentChat
		var name sql.NullString
		var startTime int64
		var endTime sql.NullInt64

		if err := rows.Scan(&chat.ID, &chat.AgentType, &name, &chat.ProjectPath, &startTime, &endTime, &chat.MessageCount); err != nil {
			continue
		}

		chat.StartTime = time.Unix(startTime, 0)
		if endTime.Valid {
			t := time.Unix(endTime.Int64, 0)
			chat.EndTime = &t
		}
		if name.Valid {
			chat.Name = name.String
		}
		chat.IsActive = chat.EndTime == nil
		chats = append(chats, chat)
	}

	return chats, nil
}
