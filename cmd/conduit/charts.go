package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/kraitsura/conduit/internal/types"
)

// renderActivityCharts displays session and agent activity charts
func renderActivityCharts(sessions []types.ConduitSession) {
	if len(sessions) == 0 {
		return
	}

	fmt.Printf("\n%sACTIVITY (last 7 days)%s\n", colorBold, colorReset)

	// Calculate data
	sessionsPerDay := calculateSessionsPerDay(sessions, 7)
	agentDist := calculateAgentDistribution(sessions)

	// Check if we have any data
	hasData := false
	for _, v := range sessionsPerDay {
		if v > 0 {
			hasData = true
			break
		}
	}

	if !hasData {
		fmt.Printf("\n  %sNo activity data%s\n", colorDim, colorReset)
		return
	}

	// Render sessions per day chart
	fmt.Println()
	graph := asciigraph.Plot(sessionsPerDay,
		asciigraph.Height(5),
		asciigraph.Width(30),
	)
	// Print with indentation and add day labels
	lines := strings.Split(graph, "\n")
	for _, line := range lines {
		fmt.Printf("  %s\n", line)
	}

	// Add day labels below chart (last 7 days: index 0 = 6 days ago, index 6 = today)
	days := []string{"S", "M", "T", "W", "T", "F", "S"}
	now := time.Now()
	todayWeekday := int(now.Weekday()) // Sunday = 0, Monday = 1, etc.

	// Build day label string: start from 6 days ago, end with today
	dayLabels := "  "
	for i := 0; i < 7; i++ {
		// Calculate weekday for (today - 6 + i) days
		daysAgo := 6 - i
		weekday := (todayWeekday - daysAgo%7 + 7) % 7
		dayLabels += fmt.Sprintf("%-4s", days[weekday])
	}
	fmt.Printf("  %s\n", dayLabels)

	// Render agent distribution
	if len(agentDist) > 0 {
		fmt.Println()
		renderAgentDistribution(agentDist)
	}
}

// calculateSessionsPerDay returns session counts for last N days
func calculateSessionsPerDay(sessions []types.ConduitSession, days int) []float64 {
	data := make([]float64, days)
	now := time.Now().Truncate(24 * time.Hour)

	for _, s := range sessions {
		sessionDay := s.StartTime.Truncate(24 * time.Hour)
		daysAgo := int(now.Sub(sessionDay).Hours() / 24)
		if daysAgo >= 0 && daysAgo < days {
			data[days-1-daysAgo]++
		}
	}
	return data
}

// calculateAgentDistribution returns agent type usage percentages
func calculateAgentDistribution(sessions []types.ConduitSession) map[string]float64 {
	counts := make(map[string]time.Duration)
	var total time.Duration

	for _, s := range sessions {
		for _, chat := range s.Chats {
			dur := chat.Duration()
			counts[chat.AgentType] += dur
			total += dur
		}
	}

	if total == 0 {
		return nil
	}

	dist := make(map[string]float64)
	for agent, dur := range counts {
		dist[agent] = float64(dur) / float64(total) * 100
	}
	return dist
}

// agentPct holds agent name and percentage for sorting
type agentPct struct {
	agent string
	pct   float64
}

// renderAgentDistribution renders horizontal bar chart for agent usage
func renderAgentDistribution(dist map[string]float64) {
	const maxBarWidth = 20

	// Sort by percentage descending
	var sorted []agentPct
	for a, p := range dist {
		sorted = append(sorted, agentPct{a, p})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].pct > sorted[j].pct
	})

	for _, ap := range sorted {
		barLen := int(ap.pct / 100 * float64(maxBarWidth))
		if barLen == 0 && ap.pct > 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen) + strings.Repeat("░", maxBarWidth-barLen)
		color := agentColor(ap.agent)
		fmt.Printf("  %s%-8s%s %s  %4.0f%%\n",
			color, ap.agent, colorReset, bar, ap.pct)
	}
}

// agentColor returns color code for agent type
func agentColor(agentType string) string {
	switch strings.ToLower(agentType) {
	case "claude":
		return colorCyan
	case "cursor":
		return colorGreen
	case "aider":
		return colorYellow
	case "copilot":
		return colorBlue
	case "continue":
		return "\033[35m" // Magenta
	case "cody":
		return "\033[95m" // Light magenta
	default:
		return colorDim
	}
}

// renderCommitTrend shows commits over the last N days
func renderCommitTrend(sessions []types.ConduitSession, days int) {
	data := make([]float64, days)
	now := time.Now().Truncate(24 * time.Hour)

	for _, s := range sessions {
		for _, c := range s.Commits {
			commitDay := c.Timestamp.Truncate(24 * time.Hour)
			daysAgo := int(now.Sub(commitDay).Hours() / 24)
			if daysAgo >= 0 && daysAgo < days {
				data[days-1-daysAgo]++
			}
		}
	}

	// Check if we have any commit data
	hasData := false
	for _, v := range data {
		if v > 0 {
			hasData = true
			break
		}
	}

	if !hasData {
		return
	}

	fmt.Printf("\n  %sCommits/day%s\n", colorBold, colorReset)
	graph := asciigraph.Plot(data,
		asciigraph.Height(3),
		asciigraph.Width(25),
	)
	for _, line := range strings.Split(graph, "\n") {
		fmt.Printf("  %s\n", line)
	}
}
