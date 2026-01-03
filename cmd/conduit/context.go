package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kraitsura/conduit/internal/types"
)

// Re-export for convenience
type DisplayMode = types.DisplayMode

const (
	DisplayGlobal  = types.DisplayGlobal
	DisplayProject = types.DisplayProject
)

// Context holds information about the CLI execution environment
type Context struct {
	// InProject indicates if CWD is inside a tracked project
	InProject bool

	// ProjectPath is the absolute path if inside a project
	ProjectPath string

	// ProjectName is the project's display name (basename)
	ProjectName string

	// GlobalFlag indicates --global was passed
	GlobalFlag bool

	// TargetProject is an explicit project argument (name or path)
	TargetProject string
}

// DetectContext determines the execution context based on CWD and tracked projects
func DetectContext(trackedProjects []string) *Context {
	ctx := &Context{}

	cwd, err := os.Getwd()
	if err != nil {
		return ctx
	}

	// Check if cwd is inside any tracked project
	for _, projectPath := range trackedProjects {
		if isSubpath(cwd, projectPath) {
			ctx.InProject = true
			ctx.ProjectPath = projectPath
			ctx.ProjectName = filepath.Base(projectPath)
			break
		}
	}

	return ctx
}

// isSubpath checks if child is inside or equal to parent
func isSubpath(child, parent string) bool {
	// Normalize paths
	child = filepath.Clean(child)
	parent = filepath.Clean(parent)

	// Exact match
	if child == parent {
		return true
	}

	// Check if child starts with parent + separator
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	// If relative path starts with "..", it's not a subpath
	return !strings.HasPrefix(rel, "..")
}

// ResolveDisplayMode determines how output should be rendered
func ResolveDisplayMode(ctx *Context, args []string) DisplayMode {
	// --global flag forces global view
	if ctx.GlobalFlag {
		return DisplayGlobal
	}

	// Explicit project argument forces project view
	if ctx.TargetProject != "" {
		return DisplayProject
	}

	// If inside a project, show project view
	if ctx.InProject {
		return DisplayProject
	}

	// Default to global view
	return DisplayGlobal
}

// ResolveProjectPath resolves a project name or path to an absolute path
func ResolveProjectPath(nameOrPath string, trackedProjects []string) string {
	// If it looks like a path, try to resolve it directly
	if strings.Contains(nameOrPath, string(filepath.Separator)) || nameOrPath == "." {
		absPath, err := filepath.Abs(nameOrPath)
		if err == nil {
			for _, p := range trackedProjects {
				if p == absPath {
					return p
				}
			}
		}
		return ""
	}

	// Try to match by name (case-insensitive)
	nameLower := strings.ToLower(nameOrPath)
	for _, projectPath := range trackedProjects {
		projectName := strings.ToLower(filepath.Base(projectPath))
		if projectName == nameLower {
			return projectPath
		}
	}

	// Try partial match
	for _, projectPath := range trackedProjects {
		projectName := strings.ToLower(filepath.Base(projectPath))
		if strings.Contains(projectName, nameLower) {
			return projectPath
		}
	}

	return ""
}

// ParseGlobalFlag checks args for --global or -g flag and returns remaining args
func ParseGlobalFlag(args []string) (bool, []string) {
	var remaining []string
	global := false

	for _, arg := range args {
		switch arg {
		case "--global", "-g":
			global = true
		default:
			remaining = append(remaining, arg)
		}
	}

	return global, remaining
}
