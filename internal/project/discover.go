package project

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kraitsura/conduit/internal/git"
	"github.com/kraitsura/conduit/internal/types"
)

// DiscoverProjects finds all git repositories under a root path
func DiscoverProjects(rootPath string) ([]types.Project, error) {
	var projects []types.Project

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if entry.Name()[0] == '.' {
			continue
		}

		projectPath := filepath.Join(rootPath, entry.Name())

		// Check if it's a git repo
		if !git.IsGitRepo(projectPath) {
			continue
		}

		project := types.Project{
			Path:      projectPath,
			Name:      entry.Name(),
			GitRemote: git.GetRemoteURL(projectPath),
		}

		// Get last activity from git
		if lastCommit, err := git.GetLastCommit(projectPath); err == nil && lastCommit != nil {
			project.LastActivity = lastCommit.Timestamp
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// EnrichProject adds statistics and agent info to a project
func EnrichProject(p *types.Project) error {
	// Get commit stats
	commitsToday, _ := git.GetCommitsToday(p.Path)
	commitsWeek, _ := git.GetCommitsThisWeek(p.Path)
	lastCommit, _ := git.GetLastCommit(p.Path)
	uncommitted := git.GetUncommittedChanges(p.Path)

	stats := &types.ProjectStats{
		CommitsToday:    len(commitsToday),
		CommitsThisWeek: len(commitsWeek),
		FilesChanged:    uncommitted,
	}

	if lastCommit != nil {
		stats.LastCommitTime = lastCommit.Timestamp
		stats.LastCommitAuthor = lastCommit.Author
	}

	p.Stats = stats
	return nil
}

// GetProjectByPath finds a project by its path
func GetProjectByPath(projects []types.Project, path string) *types.Project {
	for i := range projects {
		if projects[i].Path == path {
			return &projects[i]
		}
	}
	return nil
}

// GetProjectByName finds a project by name (case-insensitive partial match)
func GetProjectByName(projects []types.Project, name string) *types.Project {
	// Exact match first
	for i := range projects {
		if projects[i].Name == name {
			return &projects[i]
		}
	}

	// Partial match
	for i := range projects {
		if filepath.Base(projects[i].Name) == name {
			return &projects[i]
		}
	}

	return nil
}

// DetectCurrentProject checks if cwd is inside a tracked project
func DetectCurrentProject(projects []types.Project) *types.Project {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	// Check if cwd is or is under any project path
	for i := range projects {
		if cwd == projects[i].Path || isSubPath(cwd, projects[i].Path) {
			return &projects[i]
		}
	}

	return nil
}

// isSubPath checks if child is under parent
func isSubPath(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	// If rel doesn't start with "..", child is under parent
	return len(rel) > 0 && rel[0] != '.'
}

// SortByActivity sorts projects by last activity (most recent first)
func SortByActivity(projects []types.Project) {
	for i := 0; i < len(projects)-1; i++ {
		for j := i + 1; j < len(projects); j++ {
			if projects[j].LastActivity.After(projects[i].LastActivity) {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}
}

// FilterActive returns projects with activity in the last duration
func FilterActive(projects []types.Project, since time.Duration) []types.Project {
	cutoff := time.Now().Add(-since)
	var active []types.Project
	for _, p := range projects {
		if p.LastActivity.After(cutoff) || len(p.ActiveAgents) > 0 {
			active = append(active, p)
		}
	}
	return active
}
