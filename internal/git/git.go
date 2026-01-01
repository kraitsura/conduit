package git

import (
	"bufio"
	"bytes"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// AI author patterns - commits likely made by AI
var aiPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)claude`),
	regexp.MustCompile(`(?i)cursor`),
	regexp.MustCompile(`(?i)copilot`),
	regexp.MustCompile(`(?i)aider`),
	regexp.MustCompile(`(?i)\bai\b`),
	regexp.MustCompile(`(?i)assistant`),
	regexp.MustCompile(`(?i)automated`),
	regexp.MustCompile(`(?i)bot@`),
	regexp.MustCompile(`(?i)noreply@`),
}

// IsGitRepo checks if a directory is a git repository
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	cmd := exec.Command("test", "-d", gitDir)
	return cmd.Run() == nil
}

// GetRemoteURL gets the primary remote URL for a repo
func GetRemoteURL(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetCurrentBranch gets the current branch name
func GetCurrentBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetRecentCommits gets commits since a given time
func GetRecentCommits(repoPath string, since time.Time) ([]types.Commit, error) {
	// Format: hash|author|email|timestamp|message
	format := "%H|%an|%ae|%at|%s"
	sinceStr := since.Format(time.RFC3339)

	cmd := exec.Command("git", "-C", repoPath, "log",
		"--format="+format,
		"--since="+sinceStr,
	)
	output, err := cmd.Output()
	if err != nil {
		// May fail if no commits match
		return nil, nil
	}

	var commits []types.Commit
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		timestamp, _ := strconv.ParseInt(parts[3], 10, 64)
		commit := types.Commit{
			Hash:      parts[0],
			Author:    parts[1],
			Email:     parts[2],
			Timestamp: time.Unix(timestamp, 0),
			Message:   parts[4],
			IsAI:      isAICommit(parts[1], parts[2], parts[4]),
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

// GetCommitsToday returns commits made today
func GetCommitsToday(repoPath string) ([]types.Commit, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return GetRecentCommits(repoPath, today)
}

// GetCommitsThisWeek returns commits made this week
func GetCommitsThisWeek(repoPath string) ([]types.Commit, error) {
	weekAgo := time.Now().AddDate(0, 0, -7)
	return GetRecentCommits(repoPath, weekAgo)
}

// GetLastCommit returns the most recent commit
func GetLastCommit(repoPath string) (*types.Commit, error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "-1", "--format=%H|%an|%ae|%at|%s")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	line := strings.TrimSpace(string(output))
	parts := strings.SplitN(line, "|", 5)
	if len(parts) < 5 {
		return nil, nil
	}

	timestamp, _ := strconv.ParseInt(parts[3], 10, 64)
	return &types.Commit{
		Hash:      parts[0],
		Author:    parts[1],
		Email:     parts[2],
		Timestamp: time.Unix(timestamp, 0),
		Message:   parts[4],
		IsAI:      isAICommit(parts[1], parts[2], parts[4]),
	}, nil
}

// GetUncommittedChanges returns count of modified files
func GetUncommittedChanges(repoPath string) int {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	return count
}

// GetDiffStat returns a summary of changes since a commit
func GetDiffStat(repoPath string, sinceCommit string) (added, deleted, files int) {
	cmd := exec.Command("git", "-C", repoPath, "diff", "--stat", sinceCommit)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0
	}

	// Parse the summary line at the bottom
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "files changed") ||
			strings.Contains(line, "file changed") {
			// Parse: "X files changed, Y insertions(+), Z deletions(-)"
			re := regexp.MustCompile(`(\d+) files? changed`)
			if m := re.FindStringSubmatch(line); len(m) > 1 {
				files, _ = strconv.Atoi(m[1])
			}
			re = regexp.MustCompile(`(\d+) insertions?`)
			if m := re.FindStringSubmatch(line); len(m) > 1 {
				added, _ = strconv.Atoi(m[1])
			}
			re = regexp.MustCompile(`(\d+) deletions?`)
			if m := re.FindStringSubmatch(line); len(m) > 1 {
				deleted, _ = strconv.Atoi(m[1])
			}
			break
		}
	}
	return
}

// isAICommit checks if a commit was likely made by AI
func isAICommit(author, email, message string) bool {
	combined := author + " " + email + " " + message
	for _, re := range aiPatterns {
		if re.MatchString(combined) {
			return true
		}
	}
	return false
}
