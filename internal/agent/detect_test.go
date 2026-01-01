package agent

import (
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// TestNewDetector verifies detector initialization with various pattern inputs.
func TestNewDetector(t *testing.T) {
	tests := []struct {
		name         string
		patterns     []string
		wantPatterns int
	}{
		{
			name:         "empty patterns",
			patterns:     []string{},
			wantPatterns: 0,
		},
		{
			name:         "single pattern",
			patterns:     []string{"claude"},
			wantPatterns: 1,
		},
		{
			name:         "multiple patterns",
			patterns:     []string{"claude", "cursor", "aider"},
			wantPatterns: 3,
		},
		{
			name:         "patterns with special chars are quoted",
			patterns:     []string{"c++", "node.js"},
			wantPatterns: 2, // Should compile with QuoteMeta
		},
		{
			name:         "duplicate patterns",
			patterns:     []string{"claude", "claude"},
			wantPatterns: 2, // Duplicates are not deduplicated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector(tt.patterns)
			if d == nil {
				t.Fatal("NewDetector returned nil")
			}
			if len(d.patterns) != tt.wantPatterns {
				t.Errorf("got %d patterns, want %d", len(d.patterns), tt.wantPatterns)
			}
		})
	}
}

// TestMatchAgentType verifies pattern matching against process commands.
func TestMatchAgentType(t *testing.T) {
	detector := NewDetector([]string{
		"claude",
		"cursor",
		"aider",
		"copilot",
		"continue",
		"cody",
	})

	tests := []struct {
		name    string
		command string
		want    string // Empty string means no match
	}{
		// Claude variants
		{
			name:    "claude code cli",
			command: "/usr/local/bin/claude --code",
			want:    "claude",
		},
		{
			name:    "claude with path",
			command: "/home/user/.local/bin/claude",
			want:    "claude",
		},
		{
			name:    "claude uppercase",
			command: "CLAUDE --version",
			want:    "claude", // Case insensitive, returns lowercase
		},

		// Cursor variants
		{
			name:    "cursor app macOS",
			command: "/Applications/Cursor.app/Contents/MacOS/Cursor",
			want:    "cursor",
		},
		{
			name:    "cursor helper",
			command: "Cursor Helper (Renderer)",
			want:    "cursor",
		},
		{
			name:    "cursor linux",
			command: "/usr/bin/cursor",
			want:    "cursor",
		},

		// Aider variants
		{
			name:    "aider python",
			command: "python3 -m aider",
			want:    "aider",
		},
		{
			name:    "aider with args",
			command: "aider --model gpt-4",
			want:    "aider",
		},

		// Copilot variants
		{
			name:    "copilot process",
			command: "copilot-agent",
			want:    "copilot",
		},

		// Continue variants
		{
			name:    "continue extension",
			command: "continue-server",
			want:    "continue",
		},

		// Cody variants
		{
			name:    "cody process",
			command: "cody-agent",
			want:    "cody",
		},

		// Non-matches
		{
			name:    "regular vim",
			command: "vim /path/to/file.go",
			want:    "",
		},
		{
			name:    "vscode",
			command: "code /project",
			want:    "",
		},
		{
			name:    "node process",
			command: "node server.js",
			want:    "",
		},
		{
			name:    "python process",
			command: "python3 app.py",
			want:    "",
		},
		{
			name:    "shell",
			command: "/bin/zsh",
			want:    "",
		},
		{
			name:    "empty command",
			command: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.matchAgentType(tt.command)
			if got != tt.want {
				t.Errorf("matchAgentType(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

// TestMatchAgentTypeCaseInsensitive ensures matching is case-insensitive.
func TestMatchAgentTypeCaseInsensitive(t *testing.T) {
	detector := NewDetector([]string{"Claude", "CURSOR"})

	cases := []struct {
		command string
		matches bool
	}{
		{"claude", true},
		{"CLAUDE", true},
		{"Claude", true},
		{"ClAuDe", true},
		{"cursor", true},
		{"CURSOR", true},
		{"Cursor", true},
		{"vim", false},
	}

	for _, c := range cases {
		result := detector.matchAgentType(c.command)
		matched := result != ""
		if matched != c.matches {
			t.Errorf("matchAgentType(%q): got match=%v, want %v", c.command, matched, c.matches)
		}
	}
}

// TestExtractProcessName verifies process name extraction from commands.
func TestExtractProcessName(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{
			name:    "simple command",
			command: "claude",
			want:    "claude",
		},
		{
			name:    "command with args",
			command: "claude --code",
			want:    "claude",
		},
		{
			name:    "absolute path",
			command: "/usr/local/bin/claude",
			want:    "claude",
		},
		{
			name:    "absolute path with args",
			command: "/usr/local/bin/claude --version",
			want:    "claude",
		},
		{
			name:    "macOS app path",
			command: "/Applications/Cursor.app/Contents/MacOS/Cursor --flag",
			want:    "Cursor",
		},
		{
			name:    "home dir path",
			command: "/home/user/.local/bin/aider model gpt-4",
			want:    "aider",
		},
		{
			name:    "python module",
			command: "python3 -m aider",
			want:    "python3",
		},
		{
			name:    "empty command",
			command: "",
			want:    "",
		},
		{
			name:    "only spaces",
			command: "   ",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProcessName(tt.command)
			if got != tt.want {
				t.Errorf("extractProcessName(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}

// TestGroupAgentsByProject verifies grouping agents by their project paths.
func TestGroupAgentsByProject(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		agents []types.Agent
		want   map[string]int // project path -> expected count
	}{
		{
			name:   "empty agents",
			agents: []types.Agent{},
			want:   map[string]int{},
		},
		{
			name: "single agent single project",
			agents: []types.Agent{
				{PID: 1, Type: "claude", ProjectPath: "/home/user/project-a", StartTime: now},
			},
			want: map[string]int{"/home/user/project-a": 1},
		},
		{
			name: "multiple agents single project",
			agents: []types.Agent{
				{PID: 1, Type: "claude", ProjectPath: "/home/user/project-a", StartTime: now},
				{PID: 2, Type: "cursor", ProjectPath: "/home/user/project-a", StartTime: now},
			},
			want: map[string]int{"/home/user/project-a": 2},
		},
		{
			name: "multiple agents multiple projects",
			agents: []types.Agent{
				{PID: 1, Type: "claude", ProjectPath: "/home/user/project-a", StartTime: now},
				{PID: 2, Type: "cursor", ProjectPath: "/home/user/project-b", StartTime: now},
				{PID: 3, Type: "aider", ProjectPath: "/home/user/project-a", StartTime: now},
			},
			want: map[string]int{
				"/home/user/project-a": 2,
				"/home/user/project-b": 1,
			},
		},
		{
			name: "agents with empty project path excluded",
			agents: []types.Agent{
				{PID: 1, Type: "claude", ProjectPath: "/home/user/project-a", StartTime: now},
				{PID: 2, Type: "cursor", ProjectPath: "", StartTime: now}, // No project path
				{PID: 3, Type: "aider", ProjectPath: "/home/user/project-b", StartTime: now},
			},
			want: map[string]int{
				"/home/user/project-a": 1,
				"/home/user/project-b": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GroupAgentsByProject(tt.agents)

			if len(got) != len(tt.want) {
				t.Errorf("got %d project groups, want %d", len(got), len(tt.want))
			}

			for path, wantCount := range tt.want {
				gotAgents := got[path]
				if len(gotAgents) != wantCount {
					t.Errorf("project %q: got %d agents, want %d", path, len(gotAgents), wantCount)
				}
			}
		})
	}
}

// TestDetectorPatternEscaping ensures special regex chars in patterns are properly escaped.
func TestDetectorPatternEscaping(t *testing.T) {
	// Patterns with regex special characters should be treated as literals
	patterns := []string{
		"c++",        // + is regex special
		"node.js",    // . is regex special
		"(process)",  // () are regex special
		"[tool]",     // [] are regex special
		"foo|bar",    // | is regex special
		"path\\file", // \ is regex special
	}

	detector := NewDetector(patterns)

	// All patterns should compile successfully
	if len(detector.patterns) != len(patterns) {
		t.Errorf("expected %d patterns, got %d", len(patterns), len(detector.patterns))
	}

	// Test that patterns match literally
	tests := []struct {
		pattern string
		input   string
		matches bool
	}{
		{"c++", "c++ compiler", true},
		{"c++", "cpp compiler", false},
		{"node.js", "node.js runtime", true},
		{"node.js", "nodexjs", false},
		{"(process)", "my (process) here", true},
		{"[tool]", "use [tool] now", true},
		{"foo|bar", "foo|bar baz", true},
		{"foo|bar", "foo only", false},
	}

	for _, tt := range tests {
		d := NewDetector([]string{tt.pattern})
		result := d.matchAgentType(tt.input)
		matched := result != ""
		if matched != tt.matches {
			t.Errorf("pattern %q against %q: got %v, want %v",
				tt.pattern, tt.input, matched, tt.matches)
		}
	}
}

// TestParseProcessOutput simulates parsing ps command output.
func TestParseProcessOutput(t *testing.T) {
	// Simulate ps -eo pid,command output lines (after header)
	tests := []struct {
		name    string
		line    string
		wantPID int
		wantCmd string
		valid   bool
	}{
		{
			name:    "normal process",
			line:    "  1234 /usr/bin/claude --code",
			wantPID: 1234,
			wantCmd: "/usr/bin/claude --code",
			valid:   true,
		},
		{
			name:    "high pid",
			line:    "99999 node server.js",
			wantPID: 99999,
			wantCmd: "node server.js",
			valid:   true,
		},
		{
			name:    "pid with leading spaces",
			line:    "     1 /sbin/init",
			wantPID: 1,
			wantCmd: "/sbin/init",
			valid:   true,
		},
		{
			name:    "command with spaces",
			line:    "  500 /Applications/Cursor.app/Contents/MacOS/Cursor Helper (Renderer)",
			wantPID: 500,
			wantCmd: "/Applications/Cursor.app/Contents/MacOS/Cursor Helper (Renderer)",
			valid:   true,
		},
		{
			name:  "empty line",
			line:  "",
			valid: false,
		},
		{
			name:  "only spaces",
			line:  "      ",
			valid: false,
		},
		{
			name:  "no command part",
			line:  "1234",
			valid: false,
		},
		{
			name:  "invalid pid",
			line:  "abc /bin/command",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the parsing logic from getProcesses
			line := tt.line
			if len(line) > 0 {
				// Trim leading/trailing space
				trimmed := trimSpace(line)
				if trimmed == "" {
					if tt.valid {
						t.Error("expected valid but got empty after trim")
					}
					return
				}

				// Split into PID and command
				parts := splitN2(trimmed, " ", 2)
				if len(parts) < 2 {
					if tt.valid {
						t.Error("expected valid but couldn't split")
					}
					return
				}

				pid := parseInt(parts[0])
				if pid == 0 && parts[0] != "0" {
					if tt.valid {
						t.Error("expected valid but invalid PID")
					}
					return
				}

				cmd := parts[1]

				if !tt.valid {
					t.Error("expected invalid but parsed successfully")
					return
				}

				if pid != tt.wantPID {
					t.Errorf("PID: got %d, want %d", pid, tt.wantPID)
				}
				if cmd != tt.wantCmd {
					t.Errorf("Command: got %q, want %q", cmd, tt.wantCmd)
				}
			} else if tt.valid {
				t.Error("expected valid but line is empty")
			}
		})
	}
}

// Helper functions for test parsing
func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func splitN2(s, sep string, n int) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func parseInt(s string) int {
	result := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + int(c-'0')
	}
	return result
}

// BenchmarkMatchAgentType measures pattern matching performance.
func BenchmarkMatchAgentType(b *testing.B) {
	detector := NewDetector([]string{
		"claude", "cursor", "aider", "copilot", "continue", "cody",
	})

	commands := []string{
		"/usr/local/bin/claude --code",
		"/Applications/Cursor.app/Contents/MacOS/Cursor",
		"python3 -m aider --model gpt-4",
		"vim /path/to/file.go",                       // Non-match
		"/bin/bash",                                  // Non-match
		"node --inspect app.js",                      // Non-match
		"/usr/lib/systemd/systemd --user",            // Non-match
		"/System/Library/Frameworks/AppKit.framework", // Non-match
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := commands[i%len(commands)]
		detector.matchAgentType(cmd)
	}
}

// BenchmarkGroupAgentsByProject measures grouping performance.
func BenchmarkGroupAgentsByProject(b *testing.B) {
	now := time.Now()
	agents := make([]types.Agent, 100)
	for i := 0; i < 100; i++ {
		agents[i] = types.Agent{
			PID:         i,
			Type:        "claude",
			ProjectPath: "/home/user/project-" + string(rune('a'+i%10)),
			StartTime:   now,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GroupAgentsByProject(agents)
	}
}
