# Conduit Intelligence Layer

## Overview

Conduit uses headless Claude sessions to generate summaries and insights. This closes the comprehension gap - AI summarizes AI output so humans can keep up.

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│               HEADLESS CLAUDE PIPELINE                     │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  TRIGGER                                                   │
│  User runs: conduit recap / digest / diff / next           │
│                         │                                  │
│                         ▼                                  │
│  GATHER CONTEXT (Daemon)                                   │
│  • Git log, diffs, uncommitted changes                     │
│  • Session logs from database                              │
│  • Existing notes, bugs, ideas                             │
│  • File contents (if needed)                               │
│                         │                                  │
│                         ▼                                  │
│  SPAWN HEADLESS CLAUDE                                     │
│  • Sandboxed environment                                   │
│  • Read-only access to git/files                           │
│  • Write access only via conduit CLI                       │
│                         │                                  │
│                         ▼                                  │
│  GENERATE & STORE                                          │
│  • Claude analyzes context                                 │
│  • Writes summary via: conduit db add-summary              │
│  • Detects anomalies via: conduit db add-anomaly           │
│                         │                                  │
│                         ▼                                  │
│  DISPLAY                                                   │
│  • CLI reads from database                                 │
│  • Formats and displays to user                            │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## Authentication

### Option 1: OAuth Token (Recommended)

From Claude Code's OAuth flow. Stored in `~/.claude/config.json`.

```json
{
  "claude": {
    "auth_type": "oauth",
    "oauth_token": "sk-ant-oaut..."
  }
}
```

### Option 2: API Key

Direct Anthropic API key from console.anthropic.com.

```json
{
  "claude": {
    "auth_type": "api_key",
    "api_key": "sk-ant-api..."
  }
}
```

### Setup Command

```bash
conduit auth

🔐 Claude Authentication
──────────────────────────────────

[1] OAuth Token (from Claude Code)
[2] API Key (from Anthropic console)

> 1

Paste token: ****
✓ Validated and saved
```

---

## Sandbox Permissions

Headless Claude runs with strict permissions:

| Permission | Status |
|------------|--------|
| Read git history | ✅ Allowed |
| Read file contents | ✅ Allowed |
| Read conduit database | ✅ Allowed |
| Write via conduit CLI | ✅ Allowed |
| Network access | ❌ Blocked |
| File writes | ❌ Blocked |
| Shell commands | ❌ Blocked |
| Other tools | ❌ Blocked |

---

## CLI Commands for Claude

Internal commands used by headless Claude to write results:

```bash
# Add a summary
conduit db add-summary \
  --type <recap|digest|diff> \
  --project <path> \
  --content "<summary text>"

# Add an anomaly
conduit db add-anomaly \
  --project <path> \
  --severity <low|medium|high> \
  --content "<description>"

# Add risk assessment
conduit db add-risk \
  --project <path> \
  --file <filepath> \
  --level <low|medium|high> \
  --reason "<explanation>"
```

---

## Summary Types

### Recap
Generated on-demand via `conduit recap`.

**Input**: Activity since last user interaction
**Output**: Natural language summary of what happened

```
Claude completed 2 sessions. Finished OAuth Google
integration, started GitHub but hit rate limit.
Tests passing. One TODO left in oauth.go:127.
```

### Digest
Generated daily/weekly via `conduit digest` or scheduled.

**Input**: All activity in time period
**Output**: Comprehensive summary with highlights

### Semantic Diff
Generated via `conduit diff <range>`.

**Input**: Git diff for commit range
**Output**: Categorized changes (behavior, refactor, tests, risk)

### Attention Queue
Generated via `conduit next`.

**Input**: All projects, sessions, anomalies
**Output**: Prioritized list of items needing human attention

---

## Anomaly Detection

Claude analyzes patterns to detect:

| Anomaly | Detection |
|---------|-----------|
| Spinning | Same error appears 5+ times |
| Undo loop | Agent reverting own changes |
| Stuck | Long duration with no commits |
| High churn | Excessive file modifications |
| Risk | Security-sensitive files touched |

Anomalies are stored and surfaced in `conduit status` and `conduit next`.

---

## Cost Management

### Tracking

All API calls are logged:

```sql
INSERT INTO costs (
  project_path, operation, model,
  input_tokens, output_tokens, cost
) VALUES (?, ?, ?, ?, ?, ?);
```

### Limits

Optional daily cost limit in config:

```json
{
  "claude": {
    "max_cost_per_day": 5.00
  }
}
```

### Model Selection

Default: `claude-sonnet-4-20250514` (fast, cheap)

For complex analysis: `claude-opus-4-20250514`

```json
{
  "claude": {
    "model": "claude-sonnet-4-20250514",
    "model_complex": "claude-opus-4-20250514"
  }
}
```

---

## Prompt Templates

### Recap Prompt

```
You are analyzing developer activity for a project summary.

Project: {project_name}
Time range: {since} to now

Recent commits:
{commits}

Session activity:
{sessions}

User notes:
{notes}

Summarize what happened in 2-3 sentences. Focus on:
- What was accomplished
- What's in progress
- What needs attention

Be concise. Write for a developer context-switching back to this project.
```

### Diff Analysis Prompt

```
Analyze this git diff and categorize the changes.

Diff:
{diff}

Categorize into:
1. Behavior changes (new features, bug fixes)
2. Refactors (no behavior change)
3. Tests (new or modified)
4. Config/docs

Assess risk level (low/medium/high) based on:
- Authentication/security code
- Database schema changes
- Public API changes
- Deletion of tests

Output in structured format.
```

---

## Implementation Notes

### Spawning Headless Claude

```go
func spawnHeadlessClaude(context string) error {
    cmd := exec.Command("claude",
        "--headless",
        "--sandbox",
        "--allowed-tools", "conduit",
    )
    cmd.Stdin = strings.NewReader(context)
    return cmd.Run()
}
```

### Context Size Management

- Truncate large diffs to most recent changes
- Summarize long session logs
- Limit file content to relevant sections
- Target: <50k tokens per request

### Caching

- Cache summaries by content hash
- Skip regeneration if inputs unchanged
- TTL: 1 hour for recap, 24 hours for digest
