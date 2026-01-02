# Conduit Architecture

## System Overview

```
┌────────────────────────────────────────────────────────────────┐
│                         CONDUIT                                │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  USER LAYER                                                    │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  CLI (bin/conduit)          Desktop App (future)        │  │
│  └─────────────────────────────────────────────────────────┘  │
│                              │                                 │
│                              ▼                                 │
│  INTELLIGENCE LAYER                                            │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  Headless Claude (sandboxed, read-only, summaries)      │  │
│  └─────────────────────────────────────────────────────────┘  │
│                              │                                 │
│                              ▼                                 │
│  SERVICE LAYER                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  Daemon (bin/conduitd)                                  │  │
│  │  • Process detection    • Session management            │  │
│  │  • Git monitoring       • Chat history sync             │  │
│  └─────────────────────────────────────────────────────────┘  │
│                              │                                 │
│                              ▼                                 │
│  DATA LAYER                                                    │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │  SQLite (~/.conduit/conduit.db)                         │  │
│  │  Config (~/.conduit/config.json)                        │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

## Components

### CLI Client (`bin/conduit`)

- Lightweight command dispatcher
- Communicates with daemon via Unix socket
- Formats and displays output
- No heavy processing

### Daemon (`bin/conduitd`)

- Background process monitoring
- Git repository watching
- Session boundary detection
- Chat history synchronization
- Auto-start on first CLI use
- Auto-shutdown after 30min idle

### Unix Socket Protocol

Location: `~/.conduit/conduit.sock`

Request format:
```json
{
  "command": "status|projects|agents|sessions|...",
  "project": "optional/project/path",
  "args": {}
}
```

### Headless Claude

- Spawned for summary generation
- Strict sandbox: read-only + conduit CLI only
- Writes results via `conduit db` commands
- Authenticated via OAuth token or API key

## Data Flow

```
Agent Activity → Daemon Detection → SQLite Store
                                         │
User Command → CLI → Socket → Daemon → Response
                                         │
                              Headless Claude (if summary needed)
```

## Directory Structure

```
~/.conduit/
├── config.json      # User configuration
├── conduit.db       # SQLite database
├── conduit.sock     # Unix socket (runtime)
└── conduit.pid      # Daemon PID (runtime)
```

## Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/daemon` | Background service |
| `internal/agent` | Process detection |
| `internal/git` | Git operations |
| `internal/project` | Project discovery |
| `internal/store` | SQLite persistence |
| `internal/config` | Configuration |
| `internal/types` | Shared data types |
