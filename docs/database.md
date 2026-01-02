# Conduit Database Schema

## Overview

SQLite database at `~/.conduit/conduit.db`

All timestamps are Unix epoch (seconds).

---

## Core Tables

### projects
Tracked git repositories.

```sql
CREATE TABLE projects (
    path TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    git_remote TEXT,
    last_activity INTEGER,
    created_at INTEGER NOT NULL
);
```

### activities
Event log for all tracked events.

```sql
CREATE TABLE activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    project_path TEXT NOT NULL,
    type TEXT NOT NULL,  -- agent_start, agent_stop, commit, branch, file_change
    agent_type TEXT,
    data TEXT,           -- JSON blob
    FOREIGN KEY (project_path) REFERENCES projects(path)
);

CREATE INDEX idx_activities_project ON activities(project_path);
CREATE INDEX idx_activities_timestamp ON activities(timestamp);
```

---

## Session Tables

### conduit_sessions
Work windows bounded by 30-minute gaps.

```sql
CREATE TABLE conduit_sessions (
    id TEXT PRIMARY KEY,
    project_path TEXT NOT NULL,
    project_name TEXT NOT NULL,
    branch TEXT,
    start_time INTEGER NOT NULL,
    end_time INTEGER,
    is_active INTEGER DEFAULT 1,
    FOREIGN KEY (project_path) REFERENCES projects(path)
);
```

### agent_chats
Individual conversations with AI agents.

```sql
CREATE TABLE agent_chats (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    name TEXT,
    project_path TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    end_time INTEGER,
    message_count INTEGER DEFAULT 0,
    is_active INTEGER DEFAULT 1,
    FOREIGN KEY (session_id) REFERENCES conduit_sessions(id)
);
```

### session_commits
Links commits to sessions.

```sql
CREATE TABLE session_commits (
    session_id TEXT NOT NULL,
    commit_hash TEXT NOT NULL,
    message TEXT,
    timestamp INTEGER NOT NULL,
    is_ai INTEGER DEFAULT 0,
    PRIMARY KEY (session_id, commit_hash)
);
```

---

## Thread Tables

### threads
Named work streams spanning multiple sessions.

```sql
CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    project_path TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT DEFAULT 'active',  -- active, stalled, completed
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (project_path) REFERENCES projects(path)
);

CREATE INDEX idx_threads_project ON threads(project_path);
CREATE INDEX idx_threads_status ON threads(status);
```

### thread_sessions
Links sessions to threads.

```sql
CREATE TABLE thread_sessions (
    thread_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    PRIMARY KEY (thread_id, session_id),
    FOREIGN KEY (thread_id) REFERENCES threads(id),
    FOREIGN KEY (session_id) REFERENCES conduit_sessions(id)
);
```

---

## Capture Tables

### insights
Notes, bugs, and ideas.

```sql
CREATE TABLE insights (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,          -- note, bug, idea
    content TEXT NOT NULL,
    project_path TEXT,
    thread_id TEXT,
    session_id TEXT,
    file_path TEXT,
    line_number INTEGER,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (project_path) REFERENCES projects(path),
    FOREIGN KEY (thread_id) REFERENCES threads(id)
);

CREATE INDEX idx_insights_project ON insights(project_path);
CREATE INDEX idx_insights_type ON insights(type);
CREATE INDEX idx_insights_created ON insights(created_at);
```

### checkpoints
Human-verified markers.

```sql
CREATE TABLE checkpoints (
    id TEXT PRIMARY KEY,
    project_path TEXT NOT NULL,
    commit_hash TEXT,
    branch TEXT,
    message TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (project_path) REFERENCES projects(path)
);

CREATE INDEX idx_checkpoints_project ON checkpoints(project_path);
```

### reflections
Session reflections.

```sql
CREATE TABLE reflections (
    id TEXT PRIMARY KEY,
    project_path TEXT,
    went_well TEXT,
    unclear TEXT,
    priority TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (project_path) REFERENCES projects(path)
);
```

---

## Intelligence Tables

### summaries
AI-generated summaries.

```sql
CREATE TABLE summaries (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,          -- recap, digest, diff, anomaly
    project_path TEXT,
    session_id TEXT,
    thread_id TEXT,
    content TEXT NOT NULL,
    model TEXT,
    tokens_used INTEGER,
    cost REAL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_summaries_project ON summaries(project_path);
CREATE INDEX idx_summaries_type ON summaries(type);
```

### anomalies
Detected agent anomalies.

```sql
CREATE TABLE anomalies (
    id TEXT PRIMARY KEY,
    project_path TEXT NOT NULL,
    agent_type TEXT,
    severity TEXT NOT NULL,      -- low, medium, high
    description TEXT NOT NULL,
    resolved INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (project_path) REFERENCES projects(path)
);
```

---

## Focus Tables

### focus_sessions
Deep work sessions.

```sql
CREATE TABLE focus_sessions (
    id TEXT PRIMARY KEY,
    project_path TEXT,
    task TEXT,
    duration_planned INTEGER,    -- seconds
    duration_actual INTEGER,
    started_at INTEGER NOT NULL,
    ended_at INTEGER,
    notes_count INTEGER DEFAULT 0,
    FOREIGN KEY (project_path) REFERENCES projects(path)
);
```

---

## Configuration Tables

### workspaces
Project groupings.

```sql
CREATE TABLE workspaces (
    name TEXT PRIMARY KEY,
    projects TEXT NOT NULL,      -- JSON array of paths
    created_at INTEGER NOT NULL
);
```

### hooks
Event hooks.

```sql
CREATE TABLE hooks (
    id TEXT PRIMARY KEY,
    event TEXT NOT NULL,
    command TEXT NOT NULL,
    enabled INTEGER DEFAULT 1,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_hooks_event ON hooks(event);
```

---

## Analytics Tables

### costs
API cost tracking.

```sql
CREATE TABLE costs (
    id TEXT PRIMARY KEY,
    project_path TEXT,
    session_id TEXT,
    summary_id TEXT,
    operation TEXT,              -- recap, digest, diff
    model TEXT NOT NULL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    cost REAL NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_costs_project ON costs(project_path);
CREATE INDEX idx_costs_created ON costs(created_at);
```

### outcomes
Session outcomes for tracking success.

```sql
CREATE TABLE outcomes (
    session_id TEXT PRIMARY KEY,
    outcome TEXT NOT NULL,       -- merged, reverted, abandoned, in_progress
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES conduit_sessions(id)
);
```

---

## ID Generation

All IDs are 12-character hex strings generated from 6 random bytes:

```go
func generateID() string {
    b := make([]byte, 6)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```
