package main

import (
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kraitsura/conduit/internal/config"
	"github.com/kraitsura/conduit/internal/daemon"
	"github.com/kraitsura/conduit/internal/git"
	"github.com/kraitsura/conduit/internal/types"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
)

// SessionsResponse matches daemon's SessionsResponse
type SessionsResponse struct {
	Status         string                 `json:"status"`
	Error          string                 `json:"error,omitempty"`
	Sessions       []types.ConduitSession `json:"sessions,omitempty"`
	CurrentSession *types.ConduitSession  `json:"current_session,omitempty"`
	ProjectSummary *types.ProjectSummary  `json:"project_summary,omitempty"`
}

// showContextAwareStatus displays status based on context
func showContextAwareStatus(args []string) {
	// Parse global flag and remaining args
	globalFlag, args := ParseGlobalFlag(args)

	// Get list of tracked projects from daemon
	resp, err := sendRequest(Request{Command: "projects"})
	if err != nil {
		showLocalStatus()
		return
	}

	// Get project paths
	var trackedProjects []string
	for _, p := range resp.Projects {
		trackedProjects = append(trackedProjects, p.Path)
	}

	// Detect context
	ctx := DetectContext(trackedProjects)
	ctx.GlobalFlag = globalFlag

	// Handle explicit project argument
	if len(args) > 0 {
		ctx.TargetProject = ResolveProjectPath(args[0], trackedProjects)
		if ctx.TargetProject == "" {
			fmt.Printf("Project not found: %s\n", args[0])
			return
		}
	}

	// Resolve display mode
	mode := ResolveDisplayMode(ctx, args)

	switch mode {
	case DisplayGlobal:
		showGlobalView()
	case DisplayProject:
		projectPath := ctx.ProjectPath
		if ctx.TargetProject != "" {
			projectPath = ctx.TargetProject
		}
		showProjectView(projectPath)
	}
}

// showProjectView displays detailed view for a single project
func showProjectView(projectPath string) {
	// Get project summary from daemon
	summaryResp, err := sendSessionRequest(Request{Command: "project_summary", Project: projectPath})
	if err != nil || summaryResp.ProjectSummary == nil {
		// Fallback to basic view
		showBasicProjectView(projectPath)
		return
	}

	summary := summaryResp.ProjectSummary

	// Header: project name and branch
	branch := git.GetCurrentBranch(projectPath)
	branchStr := ""
	if branch != "" {
		branchStr = fmt.Sprintf("%s%s%s", colorDim, branch, colorReset)
	}

	fmt.Printf("%s%s/%s %s%s\n", colorBold, summary.Name, colorReset, branchStr, strings.Repeat("─", 40))

	// Current session section
	if summary.CurrentSession != nil {
		sess := summary.CurrentSession
		duration := sess.Duration().Round(time.Minute)
		fmt.Printf("\n%s▶ CURRENT SESSION%s (%s)\n", colorBold, colorReset, formatDuration(duration))

		// Show chats
		if len(sess.Chats) > 0 {
			fmt.Println("\n  Chats:")
			for _, chat := range sess.Chats {
				icon := agentIcon(chat.AgentType)
				name := truncate(chat.Name, 35)
				if name == "" {
					name = "(unnamed)"
				}

				chatDuration := chat.Duration().Round(time.Minute)
				activeMarker := ""
				timeStr := formatDuration(chatDuration) + " ago"
				if chat.IsActive {
					activeMarker = fmt.Sprintf(" %s●%s", colorGreen, colorReset)
					timeStr = formatDuration(chatDuration)
				}

				// Format message count
				msgStr := ""
				if chat.MessageCount > 0 {
					msgStr = fmt.Sprintf("  %s%d msgs%s", colorDim, chat.MessageCount, colorReset)
				}

				fmt.Printf("    %s %-8s %s\"%s\"%s  %-8s%s%s\n",
					icon, chat.AgentType, colorDim, name, colorReset, timeStr, msgStr, activeMarker)
			}
		}

		// Show detailed commits section
		uncommitted := git.GetUncommittedChanges(projectPath)
		weekAgo := time.Now().AddDate(0, 0, -7)
		recentCommits, _ := git.GetRecentCommits(projectPath, weekAgo)

		if len(recentCommits) > 0 {
			fmt.Println("\n  Recent Commits:")
			maxCommits := 5
			if len(recentCommits) < maxCommits {
				maxCommits = len(recentCommits)
			}
			for i := 0; i < maxCommits; i++ {
				c := recentCommits[i]
				shortHash := c.Hash
				if len(shortHash) > 7 {
					shortHash = shortHash[:7]
				}
				msg := truncate(c.Message, 40)
				ago := formatTimeAgo(c.Timestamp)
				aiTag := ""
				if c.IsAI {
					aiTag = fmt.Sprintf(" %s[AI]%s", colorYellow, colorReset)
				}
				fmt.Printf("    %s%s%s  %-40s  %s%s\n",
					colorDim, shortHash, colorReset, msg, ago, aiTag)
			}
			if len(recentCommits) > 5 {
				fmt.Printf("    %s... and %d more%s\n", colorDim, len(recentCommits)-5, colorReset)
			}
		}

		// Show summary line
		aiCount := 0
		for _, c := range recentCommits {
			if c.IsAI {
				aiCount++
			}
		}

		fmt.Printf("\n  %d commits this week", len(recentCommits))
		if aiCount > 0 {
			fmt.Printf(" (%d AI)", aiCount)
		}
		if uncommitted > 0 {
			fmt.Printf(" · %s%d uncommitted%s", colorYellow, uncommitted, colorReset)
		}
		fmt.Println()

	} else if len(summary.ActiveAgents) > 0 {
		// No session but have active agents
		fmt.Printf("\n%s▶ ACTIVE%s\n", colorBold, colorReset)
		for _, agent := range summary.ActiveAgents {
			duration := agent.Duration().Round(time.Minute)
			fmt.Printf("  %s %s running (%s)\n", agentIcon(agent.AgentType), agent.AgentType, formatDuration(duration))
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 55))

	// Get sessions for charts and history
	sessResp, _ := sendSessionRequest(Request{Command: "sessions", Project: projectPath, Limit: 50})

	// Render activity charts
	if sessResp != nil && len(sessResp.Sessions) > 0 {
		renderActivityCharts(sessResp.Sessions)
		fmt.Println()
		fmt.Println(strings.Repeat("─", 55))
	}

	// Session history
	if sessResp != nil && len(sessResp.Sessions) > 0 {
		// Group by date
		today := time.Now().Truncate(24 * time.Hour)
		weekAgo := today.AddDate(0, 0, -7)

		var todaySessions, weekSessions []types.ConduitSession
		for _, s := range sessResp.Sessions {
			if summary.CurrentSession != nil && s.ID == summary.CurrentSession.ID {
				continue // Skip current
			}
			if s.StartTime.After(today) {
				todaySessions = append(todaySessions, s)
			} else if s.StartTime.After(weekAgo) {
				weekSessions = append(weekSessions, s)
			}
		}

		if len(todaySessions) > 0 {
			fmt.Printf("\n%s■ TODAY%s\n", colorBold, colorReset)
			for _, s := range todaySessions {
				printSessionLine(s)
			}
		}

		if len(weekSessions) > 0 {
			fmt.Printf("\n%s■ THIS WEEK%s\n", colorBold, colorReset)
			for _, s := range weekSessions[:min(5, len(weekSessions))] {
				printSessionLine(s)
			}
		}

		// Summary stats
		totalTime := time.Duration(0)
		totalCommits := 0
		aiCommits := 0
		for _, s := range sessResp.Sessions {
			totalTime += s.Duration()
			for _, c := range s.Commits {
				totalCommits++
				if c.IsAI {
					aiCommits++
				}
			}
		}

		aiPct := 0
		if totalCommits > 0 {
			aiPct = aiCommits * 100 / totalCommits
		}

		fmt.Printf("\n  %sTotal: %s agent time · %d commits",
			colorDim, formatDuration(totalTime), totalCommits)
		if aiPct > 0 {
			fmt.Printf(" · %d%% AI-authored", aiPct)
		}
		fmt.Printf("%s\n", colorReset)
	}

	fmt.Println()
}

// showGlobalView displays overview of all projects
func showGlobalView() {
	const width = 58

	// Get status from daemon
	resp, err := sendRequest(Request{Command: "status"})
	if err != nil {
		showLocalStatus()
		return
	}

	// Count live sessions
	sessResp, _ := sendSessionRequest(Request{Command: "sessions", Limit: 50})
	liveSessions := 0
	var activeSessions []types.ConduitSession
	if sessResp != nil {
		for _, s := range sessResp.Sessions {
			if s.IsActive {
				liveSessions++
				activeSessions = append(activeSessions, s)
			}
		}
	}

	// Header box - calculate padding for centered content
	headerContent := fmt.Sprintf("%d projects  %d active agents  %d sessions",
		resp.Daemon.ProjectCount, resp.Daemon.ActiveAgents, liveSessions)
	// Box structure: "│  " (3 chars) + content + padding + " │" (2 chars) = width total
	headerPadding := width - 5 - len(headerContent)
	if headerPadding < 0 {
		headerPadding = 0
	}

	fmt.Println()
	// Top: "╭─ CONDUIT " (11 visible chars) + dashes + "╮" (1 char) = width
	fmt.Printf("╭─%s CONDUIT %s%s╮\n", colorBold, colorReset, strings.Repeat("─", width-12))
	fmt.Printf("│  %s%s │\n", headerContent, strings.Repeat(" ", headerPadding))
	fmt.Printf("╰%s╯\n", strings.Repeat("─", width-2))

	// Active agents section
	if len(resp.Agents) > 0 {
		fmt.Printf("\n%s▶ ACTIVE NOW%s\n\n", colorBold, colorReset)

		for _, a := range resp.Agents {
			projectName := filepath.Base(a.ProjectPath) + "/"
			duration := time.Since(a.StartTime).Round(time.Minute)

			fmt.Printf("  %-18s %-8s %s\n", projectName, a.Type, formatDuration(duration))
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("─", width))

	// Projects section with activity bars
	fmt.Printf("\n%s■ PROJECTS%s %s%s%s\n\n",
		colorBold, colorReset, colorDim, padRight("sessions today", width-14), colorReset)

	// Sort projects by activity
	sortedProjects := make([]types.Project, len(resp.Projects))
	copy(sortedProjects, resp.Projects)
	sort.Slice(sortedProjects, func(i, j int) bool {
		return sortedProjects[i].LastActivity.After(sortedProjects[j].LastActivity)
	})

	// Get session counts per project
	sessionsPerProject := make(map[string]int)
	today := time.Now().Truncate(24 * time.Hour)
	if sessResp != nil {
		for _, s := range sessResp.Sessions {
			if s.StartTime.After(today) {
				sessionsPerProject[s.ProjectPath]++
			}
		}
	}

	maxSessions := 0
	for _, count := range sessionsPerProject {
		if count > maxSessions {
			maxSessions = count
		}
	}
	if maxSessions == 0 {
		maxSessions = 1
	}

	for i, p := range sortedProjects {
		if i >= 6 {
			break
		}

		agentCount := countAgentsInProject(resp.Agents, p.Path)
		agentStr := fmt.Sprintf("%s-       %s", colorDim, colorReset)
		if agentCount > 0 {
			if agentCount == 1 {
				agentStr = fmt.Sprintf("%s%d agent %s", colorGreen, agentCount, colorReset)
			} else {
				agentStr = fmt.Sprintf("%s%d agents%s", colorGreen, agentCount, colorReset)
			}
		}

		// Activity bar
		sessionCount := sessionsPerProject[p.Path]
		barWidth := (sessionCount * 8) / maxSessions
		if sessionCount > 0 && barWidth == 0 {
			barWidth = 1
		}
		bar := strings.Repeat("█", barWidth) + strings.Repeat("░", 8-barWidth)

		sessionsStr := fmt.Sprintf("%d sessions", sessionCount)
		if sessionCount == 0 {
			ago := time.Since(p.LastActivity)
			if ago > 24*time.Hour {
				sessionsStr = fmt.Sprintf("%squiet (%s)%s", colorDim, formatDuration(ago), colorReset)
			} else {
				sessionsStr = fmt.Sprintf("%squiet%s", colorDim, colorReset)
			}
		}

		fmt.Printf("  %-16s %-12s %s  %s\n", p.Name, agentStr, bar, sessionsStr)
	}

	// Recent sessions (only show if there are any)
	hasRecentSessions := false
	if sessResp != nil && len(sessResp.Sessions) > 0 {
		for _, s := range sessResp.Sessions {
			if s.StartTime.After(time.Now().Add(-7 * 24 * time.Hour)) {
				hasRecentSessions = true
				break
			}
		}
	}

	if hasRecentSessions {
		fmt.Printf("\n%s\n", strings.Repeat("─", width))
		fmt.Printf("\n%s■ RECENT SESSIONS%s\n\n", colorBold, colorReset)

		shown := 0
		for _, s := range sessResp.Sessions {
			if shown >= 5 {
				break
			}

			projectName := s.ProjectName
			if projectName == "" {
				projectName = filepath.Base(s.ProjectPath)
			}

			timeStr := "now"
			if !s.IsActive {
				ago := time.Since(s.StartTime)
				timeStr = formatDuration(ago) + " ago"
			}

			// Get first chat name
			chatName := ""
			if len(s.Chats) > 0 {
				chatName = truncate(s.Chats[0].Name, 30)
			}

			if chatName != "" {
				fmt.Printf("  %-12s %-10s %s\"%s\"%s\n",
					projectName, timeStr, colorDim, chatName, colorReset)
			} else {
				fmt.Printf("  %-12s %-10s\n", projectName, timeStr)
			}

			shown++
		}
	}

	// Warning section for uncommitted changes
	var uncommittedProjects []string
	for _, p := range resp.Projects {
		uncommitted := git.GetUncommittedChanges(p.Path)
		if uncommitted > 0 {
			uncommittedProjects = append(uncommittedProjects,
				fmt.Sprintf("%s (%d)", p.Name, uncommitted))
		}
	}

	if len(uncommittedProjects) > 0 {
		fmt.Printf("\n%s\n", strings.Repeat("─", width))
		fmt.Printf("\n%s!%s %d projects with uncommitted changes\n",
			colorYellow, colorReset, len(uncommittedProjects))
		fmt.Printf("  %s%s%s\n", colorDim, strings.Join(uncommittedProjects[:min(3, len(uncommittedProjects))], "  "), colorReset)
	}

	fmt.Println()
}

// padRight pads a string to the right with spaces
func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(" ", length-len(s)) + s
}

// showBasicProjectView is a fallback when session data isn't available
func showBasicProjectView(projectPath string) {
	name := filepath.Base(projectPath)
	branch := git.GetCurrentBranch(projectPath)

	fmt.Printf("%s%s/%s", colorBold, name, colorReset)
	if branch != "" {
		fmt.Printf(" %s%s%s", colorDim, branch, colorReset)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 55))

	// Git status
	uncommitted := git.GetUncommittedChanges(projectPath)
	if uncommitted > 0 {
		fmt.Printf("  %sUncommitted: %d files%s\n", colorYellow, uncommitted, colorReset)
	}

	// Show recent commits
	weekAgo := time.Now().AddDate(0, 0, -7)
	commits, _ := git.GetRecentCommits(projectPath, weekAgo)
	if len(commits) > 0 {
		fmt.Println("\n  Recent Commits:")
		maxCommits := 5
		if len(commits) < maxCommits {
			maxCommits = len(commits)
		}
		for i := 0; i < maxCommits; i++ {
			c := commits[i]
			shortHash := c.Hash
			if len(shortHash) > 7 {
				shortHash = shortHash[:7]
			}
			msg := truncate(c.Message, 40)
			ago := formatTimeAgo(c.Timestamp)
			aiTag := ""
			if c.IsAI {
				aiTag = fmt.Sprintf(" %s[AI]%s", colorYellow, colorReset)
			}
			fmt.Printf("    %s%s%s  %-40s  %s%s\n",
				colorDim, shortHash, colorReset, msg, ago, aiTag)
		}
		if len(commits) > 5 {
			fmt.Printf("    %s... and %d more%s\n", colorDim, len(commits)-5, colorReset)
		}

		// Summary
		aiCount := 0
		for _, c := range commits {
			if c.IsAI {
				aiCount++
			}
		}
		fmt.Printf("\n  %d commits this week", len(commits))
		if aiCount > 0 {
			fmt.Printf(" (%d AI-authored)", aiCount)
		}
		fmt.Println()
	}

	fmt.Println()
}

// printSessionLine formats a session for the history list
func printSessionLine(s types.ConduitSession) {
	timeStr := s.StartTime.Format("3:04 PM")
	duration := s.Duration().Round(time.Minute)

	// Collect agent types
	agentTypes := make(map[string]int)
	for _, chat := range s.Chats {
		agentTypes[chat.AgentType]++
	}

	var agentStrs []string
	for agentType, count := range agentTypes {
		if count > 1 {
			agentStrs = append(agentStrs, fmt.Sprintf("%s ×%d", agentType, count))
		} else {
			agentStrs = append(agentStrs, agentType)
		}
	}
	agentsStr := strings.Join(agentStrs, ", ")
	if agentsStr == "" {
		agentsStr = "no chats"
	}

	commitStr := ""
	if len(s.Commits) > 0 {
		commitStr = fmt.Sprintf("   %d commits", len(s.Commits))
	}

	fmt.Printf("  %-10s %8s   %-18s%s\n", timeStr, formatDuration(duration), agentsStr, commitStr)
}

// sendSessionRequest sends a request and decodes as SessionsResponse
func sendSessionRequest(req Request) (*SessionsResponse, error) {
	daemon.EnsureRunning()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", cfg.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to daemon")
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(req); err != nil {
		return nil, err
	}

	var resp SessionsResponse
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Status == "error" {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return &resp, nil
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatTimeAgo returns human-readable time difference
func formatTimeAgo(t time.Time) string {
	ago := time.Since(t)
	if ago < time.Minute {
		return "just now"
	}
	if ago < time.Hour {
		return fmt.Sprintf("%dm ago", int(ago.Minutes()))
	}
	if ago < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(ago.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(ago.Hours()/24))
}
