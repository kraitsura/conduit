# Conduit

Agentic Developer Observability — Track AI coding agents across your projects.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/kraitsura/conduit/main/install.sh | bash
```

Or with Go:
```bash
go install github.com/kraitsura/conduit/cmd/conduit@latest
```

## Usage

```bash
conduit              # Status overview
conduit agents       # Active AI agents
conduit projects     # List projects
conduit log          # Activity history
```

That's it. The daemon auto-starts when you run any command and auto-stops after 30 minutes of no activity.

## What It Does

Tracks AI coding agents (Claude, Cursor, Aider, etc.) running in your projects:

```
$ conduit agents
Active Agents:
  ● claude  PID 86361  delphi      2h
  ● claude  PID 10480  grimoire    45m
  ● cursor  PID 31137  my-saas     12m
```

```
$ conduit
 CONDUIT
Watching: ~/Projects
Projects: 14 | Active Agents: 3

Current: grimoire
  ● claude running (45m)
  Branch: main
  Today: 8 commits
```

## Configuration

First run creates `~/.conduit/config.json`:

```json
{
  "root_path": "~/Projects",
  "poll_interval": 5,
  "agent_patterns": ["claude", "cursor", "aider", "copilot", "continue", "cody"]
}
```

Change `root_path` to your projects directory.

## Development

```bash
make build      # Build
make link       # Symlink to /usr/local/bin for live dev
make build      # Rebuild anytime, `conduit` uses latest
```

## How It Works

- **Auto-start**: First `conduit` command spawns daemon in background
- **Auto-stop**: Daemon exits after 30 min with no agents detected
- **Tracking**: Polls for AI agent processes, logs sessions to SQLite
- **Zero config**: Works out of the box, daemon is invisible

## License

MIT
