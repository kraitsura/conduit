# Conduit Configuration

## Config File

Location: `~/.conduit/config.json`

---

## Full Schema

```jsonc
{
  // ─────────────────────────────────────────────────────────
  // CORE
  // ─────────────────────────────────────────────────────────

  // Root directory to scan for projects
  "root_path": "~/Projects",

  // Daemon polling interval (seconds)
  "poll_interval": 5,

  // Idle timeout before daemon auto-shutdown (minutes)
  "idle_timeout": 30,

  // ─────────────────────────────────────────────────────────
  // AGENT DETECTION
  // ─────────────────────────────────────────────────────────

  // Process name patterns to detect as AI agents
  "agent_patterns": [
    "claude",
    "cursor",
    "aider",
    "copilot",
    "continue",
    "cody"
  ],

  // ─────────────────────────────────────────────────────────
  // CLAUDE INTEGRATION (for AI summaries)
  // ─────────────────────────────────────────────────────────

  "claude": {
    // Authentication type: "oauth" or "api_key"
    "auth_type": "oauth",

    // OAuth token (from Claude Code ~/.claude/config.json)
    "oauth_token": "",

    // API key (alternative to OAuth)
    "api_key": "",

    // Model for summaries (cost-effective)
    "model": "claude-sonnet-4-20250514",

    // Model for complex analysis
    "model_complex": "claude-opus-4-20250514",

    // Run in sandbox mode (recommended)
    "sandbox": true,

    // Max daily spend on API calls (optional)
    "max_cost_per_day": 5.00
  },

  // ─────────────────────────────────────────────────────────
  // SUMMARIES
  // ─────────────────────────────────────────────────────────

  "summaries": {
    // Auto-generate recap on catch-up commands
    "auto_recap": true,

    // Auto-generate digest: "daily", "weekly", or "none"
    "auto_digest": "daily",

    // Time to generate daily digest (24h format)
    "digest_time": "09:00"
  },

  // ─────────────────────────────────────────────────────────
  // FOCUS MODE
  // ─────────────────────────────────────────────────────────

  "focus": {
    // Block OS notifications during focus
    "block_notifications": true,

    // Pomodoro durations (minutes)
    "pomodoro_focus": 25,
    "pomodoro_break": 5,
    "pomodoro_long_break": 15,

    // Cycles before long break
    "pomodoro_cycles": 4
  },

  // ─────────────────────────────────────────────────────────
  // PATHS
  // ─────────────────────────────────────────────────────────

  // Unix socket for CLI-daemon communication
  "socket_path": "~/.conduit/conduit.sock",

  // SQLite database
  "db_path": "~/.conduit/conduit.db",

  // ─────────────────────────────────────────────────────────
  // DISPLAY
  // ─────────────────────────────────────────────────────────

  "display": {
    // Use colors in output
    "colors": true,

    // Date format for timestamps
    "date_format": "2006-01-02 15:04",

    // Timezone (empty = local)
    "timezone": ""
  }
}
```

---

## Minimal Config

Most settings have sensible defaults. Minimal config:

```json
{
  "root_path": "~/Projects"
}
```

---

## Environment Variables

Override config with environment variables:

| Variable | Config Key |
|----------|------------|
| `CONDUIT_ROOT` | `root_path` |
| `CONDUIT_POLL_INTERVAL` | `poll_interval` |
| `CONDUIT_DB_PATH` | `db_path` |
| `ANTHROPIC_API_KEY` | `claude.api_key` |

---

## Workspaces

Define project groups in `~/.conduit/workspaces.yaml`:

```yaml
# Group related projects
client-x:
  - ~/Projects/client-x-frontend
  - ~/Projects/client-x-backend
  - ~/Projects/client-x-infra

personal:
  - ~/Projects/blog
  - ~/Projects/dotfiles
```

Usage:
```bash
conduit workspace client-x
```

---

## Commands

```bash
# View current config
conduit config show

# Set a value
conduit config set poll_interval 10

# Edit in $EDITOR
conduit config edit

# Reset to defaults
conduit config reset
```

---

## Directory Structure

```
~/.conduit/
├── config.json       # Main configuration
├── workspaces.yaml   # Workspace definitions
├── conduit.db        # SQLite database
├── conduit.sock      # Unix socket (runtime)
├── conduit.pid       # Daemon PID (runtime)
└── logs/             # Log files (if enabled)
```
