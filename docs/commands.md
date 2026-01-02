# Conduit CLI Reference

## Command Overview

```
conduit <command> [args] [flags]
```

Running `conduit` with no args shows context-aware status.

---

## OBSERVE

### `conduit status [project]`
Show current status. If in a git repo, shows project status. Otherwise shows global overview.

```
Flags:
  -g, --global    Force global view
```

### `conduit feed`
Live stream of activity across all projects.

```
Flags:
  --notify    Enable desktop notifications
  --project   Filter to specific project
```

### `conduit agents`
List all active AI agents.

```
Output: PID, type, project, duration, status
```

### `conduit health [project]`
Show project health indicators.

```
Output: test status, build status, agent status, uncommitted count
```

### `conduit watch`
Persistent live mode. Stays running, updates in real-time.

---

## NAVIGATE

### `conduit switch <project>`
Switch to project with full context restoration.

```
Shows: branch, active agents, activity since last visit, threads, notes
```

### `conduit dashboard`
Birds-eye view of all tracked projects.

```
Shows: activity bars, agent status, last activity time
```

### `conduit threads [project]`
List work threads for a project.

```
Flags:
  --all       Include completed threads
  --active    Active threads only (default)
```

### `conduit thread new <name>`
Create a new named thread.

### `conduit thread continue <id>`
Resume a thread with context.

### `conduit workspace <name>`
Switch to a workspace (group of projects).

### `conduit cd <project>`
Print project path (for shell integration).

```bash
# Usage: cd $(conduit cd myapp)
```

---

## CATCH-UP

### `conduit recap [project]`
AI-generated summary since last interaction.

```
Flags:
  --since     Custom time range (e.g., "2 days")
```

### `conduit digest`
Generate or show daily/weekly digest.

```
Flags:
  --daily     Generate daily digest
  --weekly    Generate weekly digest
```

### `conduit next`
Prioritized attention queue.

```
Shows: blocked items, risky changes, stale unreviewed work
```

### `conduit blocked`
Projects where agents are waiting on human input.

### `conduit replay <session-id>`
Step through what an agent did in a session.

### `conduit diff <range>`
Semantic diff analysis.

```
Example: conduit diff HEAD~5..HEAD
Output: behavior changes, refactors, test changes, risk level
```

---

## FOCUS

### `conduit focus [duration]`
Start a deep work session.

```
Flags:
  --task      Description of focus task

Example: conduit focus 90m --task "review auth changes"
```

### `conduit focus status`
Show current focus session status.

### `conduit focus end`
End focus session early.

### `conduit pomo [cycles]`
Start pomodoro timer.

```
Default: 4 cycles (25min focus, 5min break, 15min long break)
```

### `conduit block`
Pause notifications.

### `conduit pause [project]`
Mark project as "human mode".

### `conduit resume [project]`
End pause, resume agent awareness.

---

## CAPTURE

### `conduit note "<text>"`
Capture a quick note.

```
Auto-links to: current project, branch, thread, session
```

### `conduit bug "<text>"`
Log a bug.

### `conduit idea "<text>"`
Capture an idea.

### `conduit checkpoint [message]`
Mark current state as human-verified.

```
Creates trust marker at current commit.
```

### `conduit reflect`
Interactive session reflection.

### `conduit notes [project]`
Review captured insights.

```
Flags:
  --type      Filter by type (note, bug, idea)
  --today     Today only
  --week      Last 7 days
```

### `conduit search <query>`
Search notes, sessions, summaries.

---

## ANALYTICS

### `conduit stats`
Activity statistics.

```
Flags:
  -t, --today    Today only
  -w, --week     Last 7 days
  -m, --month    Last 30 days
```

### `conduit costs [project]`
API cost tracking.

```
Requires: cost tracking enabled in config
```

### `conduit outcomes`
Session outcome tracking (merged, reverted, abandoned).

### `conduit trends`
Pattern analysis over time.

---

## CONFIG

### `conduit init`
Initialize conduit in current directory or globally.

### `conduit config`
View or edit configuration.

```
Subcommands:
  conduit config show
  conduit config set <key> <value>
  conduit config edit
```

### `conduit auth`
Set up Claude authentication for summaries.

### `conduit daemon`
Manage background daemon.

```
Subcommands:
  conduit daemon start
  conduit daemon stop
  conduit daemon status
  conduit daemon logs
```

### `conduit on <event> <command>`
Register an event hook.

```
Events: session-start, session-end, commit, checkpoint, anomaly, focus-end
```

### `conduit hooks`
List registered hooks.

### `conduit export [format]`
Export data.

```
Formats: json, csv
```

---

## META

### `conduit help [command]`
Show help for a command.

### `conduit version`
Show version information.

---

## Global Flags

```
-g, --global     Force global context
-p, --project    Specify project
-q, --quiet      Minimal output
--json           JSON output
```
