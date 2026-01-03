package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/kraitsura/conduit/internal/git"
	"github.com/kraitsura/conduit/internal/types"
)

// showSessions displays session history
func showSessions(args []string) {
	globalFlag, args := ParseGlobalFlag(args)

	// Parse time filters
	var since time.Time
	limit := 20

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--today", "-t":
			since = time.Now().Truncate(24 * time.Hour)
		case "--week", "-w":
			since = time.Now().AddDate(0, 0, -7)
		case "--limit", "-n":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &limit)
				i++
			}
		}
	}

	// Get project context
	resp, err := sendRequest(Request{Command: "projects"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var trackedProjects []string
	for _, p := range resp.Projects {
		trackedProjects = append(trackedProjects, p.Path)
	}

	projectPath := ""
	if !globalFlag {
		ctx := DetectContext(trackedProjects)
		if ctx.InProject {
			projectPath = ctx.ProjectPath
		}

		// Check for explicit project argument
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				resolved := ResolveProjectPath(arg, trackedProjects)
				if resolved != "" {
					projectPath = resolved
					break
				}
			}
		}
	}

	// Get sessions
	req := Request{Command: "sessions", Project: projectPath, Limit: limit}
	sessResp, err := sendSessionRequest(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if sessResp == nil || len(sessResp.Sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	// Filter by time if needed
	var filtered []types.ConduitSession
	for _, s := range sessResp.Sessions {
		if !since.IsZero() && s.StartTime.Before(since) {
			continue
		}
		filtered = append(filtered, s)
	}

	if len(filtered) == 0 {
		fmt.Println("No sessions in the specified time range")
		return
	}

	// Display header
	if projectPath != "" {
		fmt.Printf("%sSessions for %s%s\n\n", colorBold, filepath.Base(projectPath), colorReset)
	} else {
		fmt.Printf("%sSessions (all projects)%s\n\n", colorBold, colorReset)
	}

	// Group by date
	currentDate := ""
	for _, s := range filtered {
		dateStr := s.StartTime.Format("Mon, Jan 2")
		if dateStr != currentDate {
			if currentDate != "" {
				fmt.Println()
			}
			fmt.Printf("%s%s%s\n", colorDim, dateStr, colorReset)
			currentDate = dateStr
		}

		projectName := filepath.Base(s.ProjectPath)
		timeStr := s.StartTime.Format("3:04 PM")
		duration := s.Duration().Round(time.Minute)

		// Collect agents
		agentTypes := make(map[string]int)
		for _, chat := range s.Chats {
			agentTypes[chat.AgentType]++
		}
		var agentStrs []string
		for t, c := range agentTypes {
			if c > 1 {
				agentStrs = append(agentStrs, fmt.Sprintf("%s×%d", t, c))
			} else {
				agentStrs = append(agentStrs, t)
			}
		}
		agentsStr := strings.Join(agentStrs, ", ")
		if agentsStr == "" {
			agentsStr = "-"
		}

		activeMarker := ""
		if s.IsActive {
			activeMarker = fmt.Sprintf(" %s●%s", colorGreen, colorReset)
		}

		if globalFlag || projectPath == "" {
			fmt.Printf("  %-12s %-8s %8s  %-15s%s\n",
				projectName, timeStr, formatDuration(duration), agentsStr, activeMarker)
		} else {
			fmt.Printf("  %-8s %8s  %-20s%s\n",
				timeStr, formatDuration(duration), agentsStr, activeMarker)
		}
	}

	fmt.Println()
}

// searchSessions searches for sessions by name/content
func searchSessions(query string, args []string) {
	globalFlag, _ := ParseGlobalFlag(args)

	// Get project context
	resp, err := sendRequest(Request{Command: "projects"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var trackedProjects []string
	for _, p := range resp.Projects {
		trackedProjects = append(trackedProjects, p.Path)
	}

	projectPath := ""
	if !globalFlag {
		ctx := DetectContext(trackedProjects)
		if ctx.InProject {
			projectPath = ctx.ProjectPath
		}
	}

	// Get sessions
	sessResp, err := sendSessionRequest(Request{Command: "sessions", Project: projectPath, Limit: 100})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if sessResp == nil || len(sessResp.Sessions) == 0 {
		fmt.Println("No sessions found")
		return
	}

	// Search
	queryLower := strings.ToLower(query)
	var matches []types.ConduitSession

	for _, s := range sessResp.Sessions {
		matched := false

		// Search in chat names
		for _, chat := range s.Chats {
			if strings.Contains(strings.ToLower(chat.Name), queryLower) {
				matched = true
				break
			}
		}

		if matched {
			matches = append(matches, s)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No sessions matching \"%s\"\n", query)
		return
	}

	fmt.Printf("%sSearch results for \"%s\"%s (%d matches)\n\n", colorBold, query, colorReset, len(matches))

	for _, s := range matches {
		projectName := filepath.Base(s.ProjectPath)
		ago := time.Since(s.StartTime)
		duration := s.Duration().Round(time.Minute)

		// Find matching chat name
		chatName := ""
		for _, chat := range s.Chats {
			if strings.Contains(strings.ToLower(chat.Name), queryLower) {
				chatName = chat.Name
				break
			}
		}

		fmt.Printf("  %-12s %s ago  %s\n", projectName, formatDuration(ago), formatDuration(duration))
		if chatName != "" {
			// Highlight match
			highlighted := highlightMatch(chatName, query)
			fmt.Printf("    %s\"%s\"%s\n", colorDim, highlighted, colorReset)
		}
	}

	fmt.Println()
}

// highlightMatch highlights the query in text
func highlightMatch(text, query string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(query))
	if idx == -1 {
		return text
	}
	return text[:idx] + colorYellow + text[idx:idx+len(query)] + colorReset + colorDim + text[idx+len(query):]
}

// showStats displays statistics
func showStats(args []string) {
	globalFlag, args := ParseGlobalFlag(args)

	// Parse time filter
	since := time.Now().AddDate(0, 0, -7) // Default: last week
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--today", "-t":
			since = time.Now().Truncate(24 * time.Hour)
		case "--week", "-w":
			since = time.Now().AddDate(0, 0, -7)
		case "--month", "-m":
			since = time.Now().AddDate(0, -1, 0)
		case "--all":
			since = time.Time{}
		}
	}

	// Get project context
	resp, err := sendRequest(Request{Command: "projects"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var trackedProjects []string
	for _, p := range resp.Projects {
		trackedProjects = append(trackedProjects, p.Path)
	}

	projectPath := ""
	if !globalFlag {
		ctx := DetectContext(trackedProjects)
		if ctx.InProject {
			projectPath = ctx.ProjectPath
		}

		// Check for explicit project argument
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				resolved := ResolveProjectPath(arg, trackedProjects)
				if resolved != "" {
					projectPath = resolved
					break
				}
			}
		}
	}

	// Get sessions
	sessResp, err := sendSessionRequest(Request{Command: "sessions", Project: projectPath, Limit: 1000})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Calculate stats
	var totalDuration time.Duration
	var totalChatTime time.Duration
	sessionCount := 0
	commitCount := 0
	aiCommitCount := 0
	agentCounts := make(map[string]int)

	if sessResp != nil {
		for _, s := range sessResp.Sessions {
			if !since.IsZero() && s.StartTime.Before(since) {
				continue
			}

			sessionCount++
			totalDuration += s.Duration()
			totalChatTime += s.TotalChatTime()
			commitCount += len(s.Commits)

			for _, c := range s.Commits {
				if c.IsAI {
					aiCommitCount++
				}
			}

			for _, chat := range s.Chats {
				agentCounts[chat.AgentType]++
			}
		}
	}

	// Also get commit stats from git
	gitCommits := 0
	gitAICommits := 0
	if projectPath != "" {
		if commits, _ := git.GetRecentCommits(projectPath, since); commits != nil {
			gitCommits = len(commits)
			for _, c := range commits {
				if c.IsAI {
					gitAICommits++
				}
			}
		}
	}

	// Display
	timeRange := "Last 7 days"
	if since.IsZero() {
		timeRange = "All time"
	} else if since.After(time.Now().Truncate(24 * time.Hour)) {
		timeRange = "Today"
	} else if since.After(time.Now().AddDate(0, -1, 0)) {
		timeRange = "Last 7 days"
	} else {
		timeRange = "Last 30 days"
	}

	if projectPath != "" {
		fmt.Printf("%sStats for %s%s (%s)\n", colorBold, filepath.Base(projectPath), colorReset, timeRange)
	} else {
		fmt.Printf("%sGlobal Stats%s (%s)\n", colorBold, colorReset, timeRange)
	}

	fmt.Println(strings.Repeat("─", 40))

	fmt.Printf("\n  Sessions:      %d\n", sessionCount)
	fmt.Printf("  Agent time:    %s\n", formatDuration(totalDuration))

	if gitCommits > 0 {
		fmt.Printf("  Commits:       %d", gitCommits)
		if gitAICommits > 0 {
			fmt.Printf(" (%d AI-authored, %.0f%%)", gitAICommits, float64(gitAICommits)/float64(gitCommits)*100)
		}
		fmt.Println()
	} else if commitCount > 0 {
		fmt.Printf("  Commits:       %d", commitCount)
		if aiCommitCount > 0 {
			fmt.Printf(" (%d AI)", aiCommitCount)
		}
		fmt.Println()
	}

	if len(agentCounts) > 0 {
		fmt.Printf("\n  %sAgent breakdown:%s\n", colorDim, colorReset)
		for agent, count := range agentCounts {
			fmt.Printf("    %s %-10s %d sessions\n", agentIcon(agent), agent, count)
		}
	}

	fmt.Println()
}

// printProjectPath prints the path for a project (for shell integration)
func printProjectPath(nameOrPath string) {
	resp, err := sendRequest(Request{Command: "projects"})
	if err != nil {
		return
	}

	var trackedProjects []string
	for _, p := range resp.Projects {
		trackedProjects = append(trackedProjects, p.Path)
	}

	resolved := ResolveProjectPath(nameOrPath, trackedProjects)
	if resolved != "" {
		fmt.Println(resolved)
	}
}
