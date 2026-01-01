package project

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/testutil"
	"github.com/kraitsura/conduit/internal/types"
)

// TestIsSubPath verifies path containment detection.
func TestIsSubPath(t *testing.T) {
	tests := []struct {
		name   string
		child  string
		parent string
		want   bool
	}{
		{
			name:   "direct child",
			child:  "/home/user/projects/myapp",
			parent: "/home/user/projects",
			want:   true,
		},
		{
			name:   "nested child",
			child:  "/home/user/projects/myapp/src/main",
			parent: "/home/user/projects",
			want:   true,
		},
		{
			name:   "same path",
			child:  "/home/user/projects",
			parent: "/home/user/projects",
			want:   false, // Same path is not a subpath
		},
		{
			name:   "parent path",
			child:  "/home/user",
			parent: "/home/user/projects",
			want:   false,
		},
		{
			name:   "sibling path",
			child:  "/home/user/documents",
			parent: "/home/user/projects",
			want:   false,
		},
		{
			name:   "unrelated path",
			child:  "/var/log",
			parent: "/home/user/projects",
			want:   false,
		},
		{
			name:   "similar prefix but not subpath",
			child:  "/home/user/projects-backup",
			parent: "/home/user/projects",
			want:   false,
		},
		{
			name:   "root paths",
			child:  "/home",
			parent: "/",
			want:   true,
		},
		{
			name:   "trailing slash handling parent",
			child:  "/home/user/projects/app",
			parent: "/home/user/projects/",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSubPath(tt.child, tt.parent)
			if got != tt.want {
				t.Errorf("isSubPath(%q, %q) = %v, want %v",
					tt.child, tt.parent, got, tt.want)
			}
		})
	}
}

// TestSortByActivity verifies projects are sorted by activity time.
func TestSortByActivity(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		projects []types.Project
		wantOrder []string // Expected order of project names after sorting
	}{
		{
			name:      "empty list",
			projects:  []types.Project{},
			wantOrder: []string{},
		},
		{
			name: "single project",
			projects: []types.Project{
				{Name: "a", LastActivity: now},
			},
			wantOrder: []string{"a"},
		},
		{
			name: "already sorted",
			projects: []types.Project{
				{Name: "recent", LastActivity: now},
				{Name: "old", LastActivity: now.Add(-time.Hour)},
			},
			wantOrder: []string{"recent", "old"},
		},
		{
			name: "reverse order",
			projects: []types.Project{
				{Name: "old", LastActivity: now.Add(-time.Hour)},
				{Name: "recent", LastActivity: now},
			},
			wantOrder: []string{"recent", "old"},
		},
		{
			name: "multiple projects unsorted",
			projects: []types.Project{
				{Name: "b", LastActivity: now.Add(-2 * time.Hour)},
				{Name: "a", LastActivity: now},
				{Name: "c", LastActivity: now.Add(-1 * time.Hour)},
				{Name: "d", LastActivity: now.Add(-3 * time.Hour)},
			},
			wantOrder: []string{"a", "c", "b", "d"},
		},
		{
			name: "same activity time preserves relative order",
			projects: []types.Project{
				{Name: "first", LastActivity: now},
				{Name: "second", LastActivity: now},
				{Name: "third", LastActivity: now},
			},
			wantOrder: []string{"first", "second", "third"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			projects := make([]types.Project, len(tt.projects))
			copy(projects, tt.projects)

			SortByActivity(projects)

			if len(projects) != len(tt.wantOrder) {
				t.Fatalf("got %d projects, want %d", len(projects), len(tt.wantOrder))
			}

			for i, wantName := range tt.wantOrder {
				if projects[i].Name != wantName {
					t.Errorf("position %d: got %q, want %q", i, projects[i].Name, wantName)
				}
			}
		})
	}
}

// TestFilterActive verifies filtering by activity window.
func TestFilterActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		projects []types.Project
		since    time.Duration
		want     []string // Expected project names
	}{
		{
			name:     "empty list",
			projects: []types.Project{},
			since:    time.Hour,
			want:     []string{},
		},
		{
			name: "all active",
			projects: []types.Project{
				{Name: "a", LastActivity: now},
				{Name: "b", LastActivity: now.Add(-30 * time.Minute)},
			},
			since: time.Hour,
			want:  []string{"a", "b"},
		},
		{
			name: "some active",
			projects: []types.Project{
				{Name: "recent", LastActivity: now},
				{Name: "old", LastActivity: now.Add(-2 * time.Hour)},
			},
			since: time.Hour,
			want:  []string{"recent"},
		},
		{
			name: "none active",
			projects: []types.Project{
				{Name: "old1", LastActivity: now.Add(-2 * time.Hour)},
				{Name: "old2", LastActivity: now.Add(-3 * time.Hour)},
			},
			since: time.Hour,
			want:  []string{},
		},
		{
			name: "inactive but has agents",
			projects: []types.Project{
				{
					Name:         "with-agent",
					LastActivity: now.Add(-2 * time.Hour), // Old activity
					ActiveAgents: []types.Agent{{PID: 123, Type: "claude"}},
				},
				{
					Name:         "no-agent",
					LastActivity: now.Add(-2 * time.Hour),
					ActiveAgents: nil,
				},
			},
			since: time.Hour,
			want:  []string{"with-agent"}, // Included because of active agents
		},
		{
			name: "24 hour window",
			projects: []types.Project{
				{Name: "today", LastActivity: now.Add(-12 * time.Hour)},
				{Name: "yesterday", LastActivity: now.Add(-25 * time.Hour)},
				{Name: "last-week", LastActivity: now.Add(-7 * 24 * time.Hour)},
			},
			since: 24 * time.Hour,
			want:  []string{"today"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterActive(tt.projects, tt.since)

			if len(got) != len(tt.want) {
				t.Errorf("got %d projects, want %d", len(got), len(tt.want))
				return
			}

			for i, wantName := range tt.want {
				if got[i].Name != wantName {
					t.Errorf("position %d: got %q, want %q", i, got[i].Name, wantName)
				}
			}
		})
	}
}

// TestGetProjectByPath verifies project lookup by path.
func TestGetProjectByPath(t *testing.T) {
	projects := []types.Project{
		{Path: "/home/user/projects/app1", Name: "app1"},
		{Path: "/home/user/projects/app2", Name: "app2"},
		{Path: "/home/user/projects/lib", Name: "lib"},
	}

	tests := []struct {
		name     string
		path     string
		wantName string
		wantNil  bool
	}{
		{
			name:     "exact match first",
			path:     "/home/user/projects/app1",
			wantName: "app1",
		},
		{
			name:     "exact match middle",
			path:     "/home/user/projects/app2",
			wantName: "app2",
		},
		{
			name:     "exact match last",
			path:     "/home/user/projects/lib",
			wantName: "lib",
		},
		{
			name:    "not found",
			path:    "/home/user/projects/unknown",
			wantNil: true,
		},
		{
			name:    "subpath not matched",
			path:    "/home/user/projects/app1/src",
			wantNil: true,
		},
		{
			name:    "parent path not matched",
			path:    "/home/user/projects",
			wantNil: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProjectByPath(projects, tt.path)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("unexpected nil result")
			}
			if got.Name != tt.wantName {
				t.Errorf("got name %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

// TestGetProjectByName verifies project lookup by name.
func TestGetProjectByName(t *testing.T) {
	projects := []types.Project{
		{Path: "/home/user/projects/app1", Name: "app1"},
		{Path: "/home/user/projects/my-app", Name: "my-app"},
		{Path: "/home/user/projects/lib", Name: "lib"},
	}

	tests := []struct {
		name     string
		search   string
		wantPath string
		wantNil  bool
	}{
		{
			name:     "exact match",
			search:   "app1",
			wantPath: "/home/user/projects/app1",
		},
		{
			name:     "exact match with hyphen",
			search:   "my-app",
			wantPath: "/home/user/projects/my-app",
		},
		{
			name:    "not found",
			search:  "unknown",
			wantNil: true,
		},
		{
			name:    "partial match not supported",
			search:  "app",
			wantNil: true,
		},
		{
			name:    "case sensitive",
			search:  "APP1",
			wantNil: true,
		},
		{
			name:    "empty name",
			search:  "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProjectByName(projects, tt.search)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("unexpected nil result")
			}
			if got.Path != tt.wantPath {
				t.Errorf("got path %q, want %q", got.Path, tt.wantPath)
			}
		})
	}
}

// TestDiscoverProjects tests project discovery with real filesystem.
func TestDiscoverProjects(t *testing.T) {
	// Create temp directory structure
	root := testutil.TempDir(t)

	// Create test directories
	tests := []struct {
		name      string
		setup     func()
		wantCount int
		wantNames []string
	}{
		{
			name: "discovers git repos",
			setup: func() {
				testutil.CreateGitRepo(t, root, "project-a")
				testutil.CreateGitRepo(t, root, "project-b")
			},
			wantCount: 2,
			wantNames: []string{"project-a", "project-b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean and setup
			entries, _ := os.ReadDir(root)
			for _, e := range entries {
				os.RemoveAll(filepath.Join(root, e.Name()))
			}
			tt.setup()

			// Discover
			projects, err := DiscoverProjects(root)
			if err != nil {
				t.Fatalf("DiscoverProjects error: %v", err)
			}

			if len(projects) != tt.wantCount {
				t.Errorf("got %d projects, want %d", len(projects), tt.wantCount)
			}

			// Verify names
			names := make(map[string]bool)
			for _, p := range projects {
				names[p.Name] = true
			}
			for _, wantName := range tt.wantNames {
				if !names[wantName] {
					t.Errorf("missing expected project %q", wantName)
				}
			}
		})
	}
}

// TestDiscoverProjectsSkipsHidden verifies hidden directories are skipped.
func TestDiscoverProjectsSkipsHidden(t *testing.T) {
	root := testutil.TempDir(t)

	// Create visible git repo
	testutil.CreateGitRepo(t, root, "visible-project")

	// Create hidden directory with git repo
	hiddenPath := filepath.Join(root, ".hidden-project")
	testutil.MkdirAll(t, filepath.Join(hiddenPath, ".git"))

	projects, err := DiscoverProjects(root)
	if err != nil {
		t.Fatalf("DiscoverProjects error: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	if len(projects) > 0 && projects[0].Name != "visible-project" {
		t.Errorf("expected visible-project, got %s", projects[0].Name)
	}
}

// TestDiscoverProjectsSkipsNonGit verifies non-git directories are skipped.
func TestDiscoverProjectsSkipsNonGit(t *testing.T) {
	root := testutil.TempDir(t)

	// Create git repo
	testutil.CreateGitRepo(t, root, "git-project")

	// Create regular directory (no .git)
	regularPath := filepath.Join(root, "regular-dir")
	testutil.MkdirAll(t, regularPath)

	projects, err := DiscoverProjects(root)
	if err != nil {
		t.Fatalf("DiscoverProjects error: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	if len(projects) > 0 && projects[0].Name != "git-project" {
		t.Errorf("expected git-project, got %s", projects[0].Name)
	}
}

// TestDiscoverProjectsSkipsFiles verifies files are skipped.
func TestDiscoverProjectsSkipsFiles(t *testing.T) {
	root := testutil.TempDir(t)

	// Create git repo
	testutil.CreateGitRepo(t, root, "git-project")

	// Create a file in root
	testutil.TempFile(t, root, "readme.txt", "content")

	projects, err := DiscoverProjects(root)
	if err != nil {
		t.Fatalf("DiscoverProjects error: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

// TestDiscoverProjectsInvalidRoot tests behavior with invalid root path.
func TestDiscoverProjectsInvalidRoot(t *testing.T) {
	_, err := DiscoverProjects("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestDetectCurrentProject tests project detection based on cwd.
func TestDetectCurrentProject(t *testing.T) {
	projects := []types.Project{
		{Path: "/home/user/projects/app1", Name: "app1"},
		{Path: "/home/user/projects/app2", Name: "app2"},
	}

	// Note: This test would need to mock os.Getwd() for comprehensive testing.
	// Here we test the function behavior with an empty project list.
	result := DetectCurrentProject([]types.Project{})
	if result != nil {
		t.Error("expected nil for empty project list")
	}

	// Test that function returns nil when cwd is not in any project
	// (We can't easily test positive case without mocking os.Getwd)
	_ = projects // Used for documentation purposes
}

// BenchmarkSortByActivity measures sorting performance.
func BenchmarkSortByActivity(b *testing.B) {
	now := time.Now()
	base := make([]types.Project, 100)
	for i := 0; i < 100; i++ {
		base[i] = types.Project{
			Name:         string(rune('a' + i%26)),
			Path:         "/project/" + string(rune('a'+i%26)),
			LastActivity: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Copy to avoid sorting already-sorted data
		projects := make([]types.Project, len(base))
		copy(projects, base)
		SortByActivity(projects)
	}
}

// BenchmarkFilterActive measures filtering performance.
func BenchmarkFilterActive(b *testing.B) {
	now := time.Now()
	projects := make([]types.Project, 100)
	for i := 0; i < 100; i++ {
		projects[i] = types.Project{
			Name:         string(rune('a' + i%26)),
			Path:         "/project/" + string(rune('a'+i%26)),
			LastActivity: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterActive(projects, 24*time.Hour)
	}
}

// BenchmarkIsSubPath measures path containment check performance.
func BenchmarkIsSubPath(b *testing.B) {
	child := "/home/user/projects/myapp/src/main/java"
	parent := "/home/user/projects"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isSubPath(child, parent)
	}
}
