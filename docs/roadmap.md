# Conduit Roadmap

## Implementation Phases

---

## Phase 1: Foundation
*Build on existing infrastructure*

### Already Complete ✅
- [x] Daemon architecture
- [x] Agent process detection
- [x] Session tracking (30-min gaps)
- [x] Git integration (commits, branches, AI detection)
- [x] SQLite persistence
- [x] Basic CLI commands (status, agents, sessions, stats)
- [x] Claude chat history sync

### To Build
- [ ] **Thread system**
  - Create/list/continue threads
  - Link sessions to threads
  - Thread status (active, stalled, completed)

- [ ] **Insight capture**
  - `note`, `bug`, `idea` commands
  - Context auto-linking (project, branch, thread)
  - `notes` listing command

- [ ] **Checkpoint system**
  - `checkpoint` command
  - Store verified commit markers
  - Track verified vs unverified work

- [ ] **Database migrations**
  - Add new tables (threads, insights, checkpoints)
  - Migration tooling

---

## Phase 2: Navigation
*Context switching and awareness*

### To Build
- [ ] **Enhanced `switch` command**
  - Show activity since last interaction
  - Surface relevant threads and notes
  - Display current agent status

- [ ] **Dashboard view**
  - All projects overview
  - Activity heat indicators
  - Agent status per project

- [ ] **Workspace grouping**
  - Config for workspace definitions
  - `workspace` command
  - Unified view across grouped projects

- [ ] **Improved status display**
  - Health indicators
  - Uncommitted changes
  - Active thread info

---

## Phase 3: Intelligence
*AI-powered summaries and insights*

### To Build
- [ ] **Claude authentication**
  - OAuth token support
  - API key support
  - `conduit auth` command
  - Config storage

- [ ] **Headless Claude sandbox**
  - Spawn mechanism
  - Permission restrictions
  - Output capture

- [ ] **`recap` command**
  - Gather context (git, sessions, notes)
  - Generate summary via Claude
  - Store and display

- [ ] **`next` attention queue**
  - Identify blocked items
  - Flag risky changes
  - Prioritize stale work
  - Risk scoring

- [ ] **Anomaly detection**
  - Spinning detection
  - Undo loop detection
  - Store and surface anomalies

- [ ] **Semantic diffs**
  - `diff` command with analysis
  - Categorize changes
  - Risk assessment

---

## Phase 4: Focus
*Deep work enablement*

### To Build
- [ ] **Focus sessions**
  - `focus` command with timer
  - Task tracking
  - Session logging

- [ ] **Pomodoro timer**
  - `pomo` command
  - Configurable durations
  - Cycle tracking

- [ ] **Notification blocking**
  - macOS DND integration
  - `block` command
  - Auto-unblock on session end

- [ ] **Pause/resume**
  - `pause` command
  - "Human mode" indicator
  - `resume` command

---

## Phase 5: Polish
*Analytics, hooks, integrations*

### To Build
- [ ] **Cost tracking**
  - Log API usage
  - `costs` command
  - Daily/weekly/monthly views

- [ ] **Outcome tracking**
  - Tag sessions (merged, reverted, abandoned)
  - `outcomes` command
  - Success rate metrics

- [ ] **Hooks system**
  - `on` command for registration
  - Event dispatch
  - `hooks` listing

- [ ] **Export/import**
  - JSON export
  - CSV export
  - Data portability

- [ ] **Desktop app integration**
  - Shared data layer
  - IPC between CLI and desktop
  - Unified experience

---

## Future Considerations

### Team Features
- Multi-user visibility
- Shared annotations
- Conflict detection

### Extended Agents
- VS Code extension integration
- JetBrains integration
- Terminal multiplexer awareness

### Advanced Analytics
- Productivity patterns
- Agent efficiency comparison
- Project health scoring

---

## Milestones

| Milestone | Target | Key Deliverable |
|-----------|--------|-----------------|
| v0.2 | Phase 1 | Threads + Insights |
| v0.3 | Phase 2 | Context switching |
| v0.4 | Phase 3 | AI summaries |
| v0.5 | Phase 4 | Focus mode |
| v1.0 | Phase 5 | Full feature set |

---

## Technical Priorities

1. **Stability** - Daemon reliability, graceful failures
2. **Performance** - Fast CLI response, efficient polling
3. **Data integrity** - Safe migrations, backup support
4. **UX** - Clear output, helpful errors, discoverability
