# Conduit Workflows

## User Stories & Scenarios

---

## 1. Morning Start

**Scenario**: Developer starts their day, wants to know what happened overnight.

```bash
$ conduit recap

📋 Recap (last 12 hours)
──────────────────────────────────

myapp:
  Claude completed 2 sessions overnight. Finished
  OAuth Google integration, started on GitHub
  provider but hit rate limit. Tests passing.

conduit:
  No activity since you left.

⚠️ Attention:
  • myapp: GitHub OAuth blocked (rate limit)
```

**Flow**:
1. User runs `conduit recap`
2. Daemon gathers recent activity
3. Headless Claude generates summary
4. Summary displayed and stored

---

## 2. Context Switch

**Scenario**: Developer needs to jump from project A to project B.

```bash
$ conduit switch api-core

⚡ api-core
──────────────────────────────────
Branch: main
Last activity: 3 days ago

Since you left:
  • 0 agent sessions
  • 2 manual commits (documentation)

No active threads.

Your notes:
  💡 "consider caching layer for /users endpoint"
```

**Flow**:
1. User runs `conduit switch <project>`
2. Daemon loads project context
3. Shows activity since last interaction
4. Surfaces relevant notes and threads

---

## 3. Deep Work Session

**Scenario**: Developer needs focused time to review changes.

```bash
$ conduit focus 60m --task "review oauth implementation"

🎯 Deep Work Session
──────────────────────────────────
Duration: 60 minutes
Task: review oauth implementation

✓ Notifications paused
✓ Timer started

$ conduit note "refresh token logic looks solid"
📝 Note captured

$ conduit bug "no error handling for revoked tokens"
🐛 Bug captured

$ conduit focus end

🎯 Session Complete (47:32)
Notes: 1 | Bugs: 1
```

**Flow**:
1. User starts focus session
2. System blocks notifications
3. User captures insights as they work
4. Session ends, summary saved

---

## 4. Quick Insight Capture

**Scenario**: Developer spots something while working, wants to note it fast.

```bash
$ conduit bug "login fails silently when session expires"
🐛 Bug captured
Linked to: myapp, branch feature/auth

$ conduit idea "could batch these API calls"
💡 Idea captured
```

**Flow**:
1. User runs capture command
2. System auto-links to current context (project, branch, thread)
3. Stored for later retrieval

---

## 5. Attention Triage

**Scenario**: Developer has limited time, wants to know what needs attention most.

```bash
$ conduit next

📥 Attention Queue
──────────────────────────────────

1. [blocked] myapp: GitHub OAuth rate limited
   └─ Agent waiting, needs retry or manual intervention

2. [risky] myapp: auth/oauth.go changed (+127 lines)
   └─ Authentication code, unreviewed

3. [stale] conduit: thread system (3 days unreviewed)
   └─ 4 commits, +312 lines

$ conduit checkpoint "reviewed oauth.go, looks good"
✓ Checkpoint created at a4f2c89
```

**Flow**:
1. System analyzes: blocked agents, risky changes, stale work
2. Prioritizes by urgency and risk
3. User reviews and creates checkpoints

---

## 6. Thread Management

**Scenario**: Developer working on a multi-session feature.

```bash
$ conduit thread new "implement caching layer"
🧵 Thread created: implement caching layer

# ... work happens over days ...

$ conduit threads myapp

🧵 Threads:
1. [active] implement caching layer (2 sessions, 5 commits)
2. [stalled] fix memory leak (blocked 4 days)
3. [completed] oauth integration (merged)

$ conduit thread continue 1

🧵 Resuming: implement caching layer

Last activity: yesterday
Your notes: 3
Status: Redis client added, needs config

Notes from this thread:
  💡 "consider TTL based on endpoint"
  📝 "redis vs memcached - went with redis"
```

**Flow**:
1. User creates named thread
2. Sessions and commits auto-link to active thread
3. Thread persists across days/sessions
4. Context restored on continue

---

## 7. End of Day Reflection

**Scenario**: Developer wrapping up, wants to capture state.

```bash
$ conduit reflect

📝 Session Reflection
──────────────────────────────────
Today:
  • 2 projects touched
  • 3 agent sessions (4.2 hours)
  • 1 focus session (47 min)
  • 8 commits, 4 notes captured

What went well?
> caching implementation cleaner than expected

What's still unclear?
> not sure about cache invalidation strategy

Priority tomorrow?
> finish cache invalidation, then deploy

Saved.
```

---

## 8. Team Handoff (Future)

**Scenario**: Handing work to a teammate.

```bash
$ conduit handoff --to sarah

📤 Preparing handoff for sarah
──────────────────────────────────

Thread: implement caching layer
Status: Redis client done, needs invalidation logic

Your notes:
  • "TTL based on endpoint type"
  • "watch out for race condition in /users"

Files touched:
  • internal/cache/redis.go
  • internal/api/users.go

Handoff summary copied to clipboard.
```

---

## Workflow Patterns

| Pattern | Commands | When |
|---------|----------|------|
| Morning catchup | `recap` → `next` | Start of day |
| Project switch | `switch` → `threads` | Changing context |
| Deep work | `focus` → `note/bug/idea` → `checkpoint` | Focused review |
| Quick capture | `note/bug/idea` | Anytime |
| Triage | `next` → `checkpoint` | Limited time |
| End of day | `reflect` | Wrapping up |
