# Conduit Documentation

## Quick Links

| Document | Description |
|----------|-------------|
| [Vision](vision.md) | Mission, problem space, core principles |
| [Features](features.md) | Complete feature list by category |
| [Architecture](architecture.md) | System design and components |
| [Workflows](workflows.md) | User stories and usage patterns |
| [Commands](commands.md) | CLI reference |
| [Config](config.md) | Configuration options |
| [Database](database.md) | SQLite schema |
| [Intelligence](intelligence.md) | Headless Claude integration |
| [Roadmap](roadmap.md) | Implementation phases |

## What is Conduit?

Conduit is an agentic workflow companion that helps humans keep pace with AI-assisted development.

```
OBSERVE → NAVIGATE → CATCH-UP → FOCUS → CAPTURE
```

## Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/kraitsura/conduit/main/install.sh | bash

# Initialize
conduit init

# Check status
conduit status

# See what's happening
conduit feed
```

## Core Commands

```bash
conduit status          # What's happening now
conduit switch <proj>   # Jump to project with context
conduit recap           # AI summary of recent activity
conduit focus 60m       # Start deep work session
conduit note "..."      # Capture an insight
conduit next            # What needs attention
```

## Philosophy

- **Thin veil**: Observe everything, interrupt nothing
- **Human amplifier**: Speed up humans, don't slow machines
- **AI summarizes AI**: Use Claude to digest Claude's output
- **Context is king**: Solve "where was I?" ruthlessly
