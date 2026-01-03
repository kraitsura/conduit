package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectContext(t *testing.T) {
	// Create temp directories for testing
	tmpDir := t.TempDir()
	// Resolve symlinks (macOS uses /var -> /private/var)
	tmpDir, _ = filepath.EvalSymlinks(tmpDir)

	projectA := filepath.Join(tmpDir, "project-a")
	projectB := filepath.Join(tmpDir, "project-b")
	subDir := filepath.Join(projectA, "src", "components")

	os.MkdirAll(subDir, 0755)
	os.MkdirAll(projectB, 0755)

	trackedProjects := []string{projectA, projectB}

	tests := []struct {
		name           string
		cwd            string
		wantInProject  bool
		wantPath       string
		wantName       string
	}{
		{
			name:           "in project root",
			cwd:            projectA,
			wantInProject:  true,
			wantPath:       projectA,
			wantName:       "project-a",
		},
		{
			name:           "in project subdirectory",
			cwd:            subDir,
			wantInProject:  true,
			wantPath:       projectA,
			wantName:       "project-a",
		},
		{
			name:           "outside projects",
			cwd:            tmpDir,
			wantInProject:  false,
			wantPath:       "",
			wantName:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to test directory
			oldWd, _ := os.Getwd()
			os.Chdir(tt.cwd)
			defer os.Chdir(oldWd)

			ctx := DetectContext(trackedProjects)

			if ctx.InProject != tt.wantInProject {
				t.Errorf("InProject = %v, want %v", ctx.InProject, tt.wantInProject)
			}
			if ctx.ProjectPath != tt.wantPath {
				t.Errorf("ProjectPath = %q, want %q", ctx.ProjectPath, tt.wantPath)
			}
			if ctx.ProjectName != tt.wantName {
				t.Errorf("ProjectName = %q, want %q", ctx.ProjectName, tt.wantName)
			}
		})
	}
}

func TestIsSubpath(t *testing.T) {
	tests := []struct {
		child  string
		parent string
		want   bool
	}{
		{"/home/user/projects/foo", "/home/user/projects/foo", true},
		{"/home/user/projects/foo/src", "/home/user/projects/foo", true},
		{"/home/user/projects/foo/src/lib", "/home/user/projects/foo", true},
		{"/home/user/projects/foobar", "/home/user/projects/foo", false},
		{"/home/user/projects/bar", "/home/user/projects/foo", false},
		{"/other/path", "/home/user/projects/foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.child+"_in_"+tt.parent, func(t *testing.T) {
			got := isSubpath(tt.child, tt.parent)
			if got != tt.want {
				t.Errorf("isSubpath(%q, %q) = %v, want %v", tt.child, tt.parent, got, tt.want)
			}
		})
	}
}

func TestResolveDisplayMode(t *testing.T) {
	tests := []struct {
		name string
		ctx  *Context
		want DisplayMode
	}{
		{
			name: "global flag overrides everything",
			ctx:  &Context{GlobalFlag: true, InProject: true},
			want: DisplayGlobal,
		},
		{
			name: "explicit target project",
			ctx:  &Context{TargetProject: "foo"},
			want: DisplayProject,
		},
		{
			name: "inside project",
			ctx:  &Context{InProject: true, ProjectPath: "/path/to/project"},
			want: DisplayProject,
		},
		{
			name: "outside projects",
			ctx:  &Context{InProject: false},
			want: DisplayGlobal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveDisplayMode(tt.ctx, nil)
			if got != tt.want {
				t.Errorf("ResolveDisplayMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveProjectPath(t *testing.T) {
	projects := []string{
		"/home/user/projects/conduit",
		"/home/user/projects/grimoire",
		"/home/user/projects/beads",
	}

	tests := []struct {
		name        string
		nameOrPath  string
		want        string
	}{
		{
			name:       "exact name match",
			nameOrPath: "conduit",
			want:       "/home/user/projects/conduit",
		},
		{
			name:       "case insensitive",
			nameOrPath: "CONDUIT",
			want:       "/home/user/projects/conduit",
		},
		{
			name:       "partial match",
			nameOrPath: "grim",
			want:       "/home/user/projects/grimoire",
		},
		{
			name:       "no match",
			nameOrPath: "nonexistent",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveProjectPath(tt.nameOrPath, projects)
			if got != tt.want {
				t.Errorf("ResolveProjectPath(%q) = %q, want %q", tt.nameOrPath, got, tt.want)
			}
		})
	}
}

func TestParseGlobalFlag(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantGlobal   bool
		wantRemaining []string
	}{
		{
			name:         "no flags",
			args:         []string{"status"},
			wantGlobal:   false,
			wantRemaining: []string{"status"},
		},
		{
			name:         "long flag",
			args:         []string{"--global", "status"},
			wantGlobal:   true,
			wantRemaining: []string{"status"},
		},
		{
			name:         "short flag",
			args:         []string{"-g", "status"},
			wantGlobal:   true,
			wantRemaining: []string{"status"},
		},
		{
			name:         "flag at end",
			args:         []string{"status", "--global"},
			wantGlobal:   true,
			wantRemaining: []string{"status"},
		},
		{
			name:         "empty args",
			args:         []string{},
			wantGlobal:   false,
			wantRemaining: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGlobal, gotRemaining := ParseGlobalFlag(tt.args)
			if gotGlobal != tt.wantGlobal {
				t.Errorf("ParseGlobalFlag() global = %v, want %v", gotGlobal, tt.wantGlobal)
			}
			if len(gotRemaining) != len(tt.wantRemaining) {
				t.Errorf("ParseGlobalFlag() remaining = %v, want %v", gotRemaining, tt.wantRemaining)
			}
		})
	}
}
