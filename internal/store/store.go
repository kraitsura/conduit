package store

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kraitsura/conduit/internal/types"
)

// Store handles persistence to SQLite
type Store struct {
	db *sql.DB
}

// Open opens or creates the database
func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close closes the database
func (s *Store) Close() error {
	return s.db.Close()
}

// init creates the schema
func (s *Store) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		path TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		git_remote TEXT,
		last_activity INTEGER,
		created_at INTEGER DEFAULT (strftime('%s', 'now'))
	);

	CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		project TEXT NOT NULL,
		type TEXT NOT NULL,
		agent_type TEXT,
		data TEXT,
		FOREIGN KEY (project) REFERENCES projects(path)
	);

	CREATE TABLE IF NOT EXISTS agent_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pid INTEGER NOT NULL,
		project TEXT NOT NULL,
		agent_type TEXT NOT NULL,
		start_time INTEGER NOT NULL,
		end_time INTEGER,
		FOREIGN KEY (project) REFERENCES projects(path)
	);

	CREATE INDEX IF NOT EXISTS idx_activities_project ON activities(project);
	CREATE INDEX IF NOT EXISTS idx_activities_timestamp ON activities(timestamp);
	CREATE INDEX IF NOT EXISTS idx_sessions_project ON agent_sessions(project);

	-- Conduit sessions (work windows bounded by 30-min gaps)
	CREATE TABLE IF NOT EXISTS conduit_sessions (
		id TEXT PRIMARY KEY,
		project_path TEXT NOT NULL,
		start_time INTEGER NOT NULL,
		end_time INTEGER,
		branch TEXT,
		FOREIGN KEY (project_path) REFERENCES projects(path)
	);

	-- Agent chats within sessions
	CREATE TABLE IF NOT EXISTS agent_chats (
		id TEXT PRIMARY KEY,
		conduit_session_id TEXT,
		agent_type TEXT NOT NULL,
		name TEXT,
		project_path TEXT NOT NULL,
		start_time INTEGER NOT NULL,
		end_time INTEGER,
		message_count INTEGER DEFAULT 0,
		FOREIGN KEY (conduit_session_id) REFERENCES conduit_sessions(id),
		FOREIGN KEY (project_path) REFERENCES projects(path)
	);

	-- Link commits to sessions
	CREATE TABLE IF NOT EXISTS session_commits (
		session_id TEXT NOT NULL,
		commit_hash TEXT NOT NULL,
		PRIMARY KEY (session_id, commit_hash),
		FOREIGN KEY (session_id) REFERENCES conduit_sessions(id)
	);

	-- Indexes for session queries
	CREATE INDEX IF NOT EXISTS idx_conduit_sessions_project ON conduit_sessions(project_path);
	CREATE INDEX IF NOT EXISTS idx_conduit_sessions_time ON conduit_sessions(start_time);
	CREATE INDEX IF NOT EXISTS idx_agent_chats_project ON agent_chats(project_path);
	CREATE INDEX IF NOT EXISTS idx_agent_chats_session ON agent_chats(conduit_session_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

// SaveProject upserts a project
func (s *Store) SaveProject(p *types.Project) error {
	_, err := s.db.Exec(`
		INSERT INTO projects (path, name, git_remote, last_activity)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			git_remote = excluded.git_remote,
			last_activity = excluded.last_activity
	`, p.Path, p.Name, p.GitRemote, p.LastActivity.Unix())
	return err
}

// GetProjects returns all tracked projects
func (s *Store) GetProjects() ([]types.Project, error) {
	rows, err := s.db.Query(`
		SELECT path, name, git_remote, last_activity FROM projects
		ORDER BY last_activity DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []types.Project
	for rows.Next() {
		var p types.Project
		var lastActivity int64
		var gitRemote sql.NullString

		if err := rows.Scan(&p.Path, &p.Name, &gitRemote, &lastActivity); err != nil {
			continue
		}

		p.LastActivity = time.Unix(lastActivity, 0)
		if gitRemote.Valid {
			p.GitRemote = gitRemote.String
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// LogActivity records an activity event
func (s *Store) LogActivity(a *types.Activity) error {
	_, err := s.db.Exec(`
		INSERT INTO activities (timestamp, project, type, agent_type, data)
		VALUES (?, ?, ?, ?, ?)
	`, a.Timestamp.Unix(), a.Project, a.Type, a.AgentType, a.Data)
	return err
}

// GetActivities returns recent activities
func (s *Store) GetActivities(projectPath string, limit int) ([]types.Activity, error) {
	query := `
		SELECT id, timestamp, project, type, agent_type, data
		FROM activities
		WHERE (? = '' OR project = ?)
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, projectPath, projectPath, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []types.Activity
	for rows.Next() {
		var a types.Activity
		var timestamp int64
		var agentType, data sql.NullString

		if err := rows.Scan(&a.ID, &timestamp, &a.Project, &a.Type, &agentType, &data); err != nil {
			continue
		}

		a.Timestamp = time.Unix(timestamp, 0)
		if agentType.Valid {
			a.AgentType = agentType.String
		}
		if data.Valid {
			a.Data = data.String
		}
		activities = append(activities, a)
	}

	return activities, nil
}

// StartAgentSession records an agent starting
func (s *Store) StartAgentSession(agent *types.Agent) error {
	_, err := s.db.Exec(`
		INSERT INTO agent_sessions (pid, project, agent_type, start_time)
		VALUES (?, ?, ?, ?)
	`, agent.PID, agent.ProjectPath, agent.Type, agent.StartTime.Unix())
	return err
}

// EndAgentSession records an agent stopping
func (s *Store) EndAgentSession(pid int) error {
	_, err := s.db.Exec(`
		UPDATE agent_sessions
		SET end_time = ?
		WHERE pid = ? AND end_time IS NULL
	`, time.Now().Unix(), pid)
	return err
}

// GetActiveAgentSessions returns sessions without end times
func (s *Store) GetActiveAgentSessions() ([]types.Agent, error) {
	rows, err := s.db.Query(`
		SELECT pid, project, agent_type, start_time
		FROM agent_sessions
		WHERE end_time IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []types.Agent
	for rows.Next() {
		var a types.Agent
		var startTime int64

		if err := rows.Scan(&a.PID, &a.ProjectPath, &a.Type, &startTime); err != nil {
			continue
		}

		a.StartTime = time.Unix(startTime, 0)
		agents = append(agents, a)
	}

	return agents, nil
}

// GetProjectStats returns aggregated stats for a project
func (s *Store) GetProjectStats(projectPath string, since time.Time) (*types.ProjectStats, error) {
	stats := &types.ProjectStats{}

	// Total agent time
	row := s.db.QueryRow(`
		SELECT COALESCE(SUM(COALESCE(end_time, strftime('%s', 'now')) - start_time), 0)
		FROM agent_sessions
		WHERE project = ? AND start_time >= ?
	`, projectPath, since.Unix())

	var totalSeconds int64
	if err := row.Scan(&totalSeconds); err == nil {
		stats.TotalAgentTime = time.Duration(totalSeconds) * time.Second
	}

	return stats, nil
}

// ActivityData is a helper for encoding activity data
type ActivityData map[string]interface{}

// Encode serializes activity data to JSON
func (d ActivityData) Encode() string {
	data, _ := json.Marshal(d)
	return string(data)
}
