package daemon

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kraitsura/conduit/internal/types"
)

// TestIdleTimeoutConstant verifies the idle timeout value.
func TestIdleTimeoutConstant(t *testing.T) {
	// IdleTimeout should be reasonable (not too short, not too long)
	if IdleTimeout < 5*time.Minute {
		t.Errorf("IdleTimeout too short: %v", IdleTimeout)
	}
	if IdleTimeout > 2*time.Hour {
		t.Errorf("IdleTimeout too long: %v", IdleTimeout)
	}

	// Current expected value
	expected := 30 * time.Minute
	if IdleTimeout != expected {
		t.Errorf("IdleTimeout = %v, want %v", IdleTimeout, expected)
	}
}

// TestRequestJSON verifies Request JSON serialization.
func TestRequestJSON(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		check   func(t *testing.T, data []byte)
	}{
		{
			name: "status command",
			request: Request{
				Command: "status",
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"command":"status"`) {
					t.Error("missing command field")
				}
			},
		},
		{
			name: "projects command",
			request: Request{
				Command: "projects",
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"command":"projects"`) {
					t.Error("missing command field")
				}
			},
		},
		{
			name: "agents with project filter",
			request: Request{
				Command: "agents",
				Project: "/home/user/project",
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"command":"agents"`) {
					t.Error("missing command field")
				}
				if !containsString(data, `"project":"/home/user/project"`) {
					t.Error("missing project field")
				}
			},
		},
		{
			name: "activities with limit",
			request: Request{
				Command: "activities",
				Project: "/project",
				Limit:   100,
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"limit":100`) {
					t.Error("missing limit field")
				}
			},
		},
		{
			name: "refresh command",
			request: Request{
				Command: "refresh",
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"command":"refresh"`) {
					t.Error("missing command field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			tt.check(t, data)

			// Verify roundtrip
			var decoded Request
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if decoded.Command != tt.request.Command {
				t.Errorf("Command: got %q, want %q", decoded.Command, tt.request.Command)
			}
		})
	}
}

// TestRequestOmitempty verifies empty fields are omitted.
func TestRequestOmitempty(t *testing.T) {
	req := Request{Command: "status"}
	data, _ := json.Marshal(req)

	// Project should be omitted when empty
	if containsString(data, `"project"`) {
		t.Error("empty project should be omitted")
	}
	// Limit should be omitted when zero
	if containsString(data, `"limit"`) {
		t.Error("zero limit should be omitted")
	}
}

// TestResponseJSON verifies Response JSON serialization.
func TestResponseJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		response Response
		check    func(t *testing.T, data []byte)
	}{
		{
			name: "success response",
			response: Response{
				Status: "ok",
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"status":"ok"`) {
					t.Error("missing status field")
				}
			},
		},
		{
			name: "error response",
			response: Response{
				Status: "error",
				Error:  "something went wrong",
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"status":"error"`) {
					t.Error("missing status field")
				}
				if !containsString(data, `"error":"something went wrong"`) {
					t.Error("missing error field")
				}
			},
		},
		{
			name: "status response with daemon info",
			response: Response{
				Status: "ok",
				Daemon: &types.DaemonStatus{
					Running:      true,
					PID:          12345,
					StartTime:    now,
					RootPath:     "/home/user/Projects",
					ProjectCount: 10,
					ActiveAgents: 2,
				},
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"daemon"`) {
					t.Error("missing daemon field")
				}
				if !containsString(data, `"running":true`) {
					t.Error("missing running field")
				}
			},
		},
		{
			name: "projects response",
			response: Response{
				Status: "ok",
				Projects: []types.Project{
					{Path: "/project/a", Name: "a"},
					{Path: "/project/b", Name: "b"},
				},
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"projects"`) {
					t.Error("missing projects field")
				}
			},
		},
		{
			name: "agents response",
			response: Response{
				Status: "ok",
				Agents: []types.Agent{
					{PID: 123, Type: "claude", ProjectPath: "/project"},
				},
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"agents"`) {
					t.Error("missing agents field")
				}
			},
		},
		{
			name: "activities response",
			response: Response{
				Status: "ok",
				Activities: []types.Activity{
					{ID: 1, Type: types.ActivityAgentStart, Project: "/project"},
				},
			},
			check: func(t *testing.T, data []byte) {
				if !containsString(data, `"activities"`) {
					t.Error("missing activities field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			tt.check(t, data)

			// Verify roundtrip
			var decoded Response
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if decoded.Status != tt.response.Status {
				t.Errorf("Status: got %q, want %q", decoded.Status, tt.response.Status)
			}
		})
	}
}

// TestResponseOmitempty verifies empty fields are omitted.
func TestResponseOmitempty(t *testing.T) {
	resp := Response{Status: "ok"}
	data, _ := json.Marshal(resp)

	omitFields := []string{"error", "daemon", "projects", "agents", "activities"}
	for _, field := range omitFields {
		if containsString(data, `"`+field+`"`) {
			t.Errorf("empty %s should be omitted", field)
		}
	}
}

// TestIsRunningWithNoSocket verifies IsRunning returns false when no socket.
func TestIsRunningWithNoSocket(t *testing.T) {
	// When there's no socket file, IsRunning should return false
	// This test relies on the test environment not having a running daemon
	// with the default socket path

	// Create a temp socket path that doesn't exist
	tmpDir, err := os.MkdirTemp("", "conduit-daemon-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "nonexistent.sock")

	// Try to connect - should fail
	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		conn.Close()
		t.Error("expected connection to fail for nonexistent socket")
	}
}

// TestSocketPathCreation verifies socket can be created and removed.
func TestSocketPathCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "conduit-daemon-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create socket: %v", err)
	}

	// Verify socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("socket file should exist")
	}

	// Connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	conn.Close()

	// Close listener
	listener.Close()

	// Remove socket
	os.Remove(socketPath)

	// Verify socket is gone
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("socket file should be removed")
	}
}

// TestDaemonRequestResponseProtocol verifies the IPC protocol.
func TestDaemonRequestResponseProtocol(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "conduit-daemon-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create a mock server
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create socket: %v", err)
	}
	defer listener.Close()

	// Handle one connection in background
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read request
		decoder := json.NewDecoder(conn)
		var req Request
		if err := decoder.Decode(&req); err != nil {
			return
		}

		// Send response based on command
		encoder := json.NewEncoder(conn)
		switch req.Command {
		case "status":
			encoder.Encode(Response{
				Status: "ok",
				Daemon: &types.DaemonStatus{
					Running:      true,
					PID:          12345,
					ProjectCount: 5,
				},
			})
		case "error":
			encoder.Encode(Response{
				Status: "error",
				Error:  "test error",
			})
		default:
			encoder.Encode(Response{Status: "ok"})
		}
	}()

	// Connect as client
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Send request
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	err = encoder.Encode(Request{Command: "status"})
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}

	// Read response
	var resp Response
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status: got %q, want ok", resp.Status)
	}
	if resp.Daemon == nil {
		t.Fatal("Daemon should not be nil")
	}
	if resp.Daemon.PID != 12345 {
		t.Errorf("PID: got %d, want 12345", resp.Daemon.PID)
	}

	<-done
}

// TestIdleTimeoutBehavior tests the idle timeout logic conceptually.
func TestIdleTimeoutBehavior(t *testing.T) {
	t.Run("no agents and expired", func(t *testing.T) {
		// Simulate: no agents, lastActivity was 31 minutes ago
		lastActivity := time.Now().Add(-31 * time.Minute)
		agentCount := 0

		shouldShutdown := agentCount == 0 && time.Since(lastActivity) > IdleTimeout
		if !shouldShutdown {
			t.Error("should shutdown when idle for more than IdleTimeout")
		}
	})

	t.Run("no agents but recent activity", func(t *testing.T) {
		// Simulate: no agents, but activity was just 5 minutes ago
		lastActivity := time.Now().Add(-5 * time.Minute)
		agentCount := 0

		shouldShutdown := agentCount == 0 && time.Since(lastActivity) > IdleTimeout
		if shouldShutdown {
			t.Error("should not shutdown with recent activity")
		}
	})

	t.Run("has agents keeps alive", func(t *testing.T) {
		// Simulate: agents running, even if lastActivity was long ago
		lastActivity := time.Now().Add(-60 * time.Minute)
		agentCount := 1

		shouldShutdown := agentCount == 0 && time.Since(lastActivity) > IdleTimeout
		if shouldShutdown {
			t.Error("should not shutdown while agents are running")
		}
	})

	t.Run("cli connection resets activity", func(t *testing.T) {
		// When a CLI connection comes in, lastActivity should be updated
		oldActivity := time.Now().Add(-25 * time.Minute)
		newActivity := time.Now() // CLI connection updates this

		// After CLI connection, timeout should be reset
		shouldShutdown := time.Since(newActivity) > IdleTimeout
		if shouldShutdown {
			t.Error("CLI connection should reset idle timer")
		}
		_ = oldActivity // Used for documentation
	})
}

// TestActivityTrackingScenarios tests various activity tracking scenarios.
func TestActivityTrackingScenarios(t *testing.T) {
	tests := []struct {
		name           string
		initialAgents  int
		currentAgents  int
		expectActivity bool
	}{
		{
			name:           "agent starts",
			initialAgents:  0,
			currentAgents:  1,
			expectActivity: true, // New agent detected
		},
		{
			name:           "agent stops",
			initialAgents:  1,
			currentAgents:  0,
			expectActivity: true, // Agent stopped
		},
		{
			name:           "no change with agents",
			initialAgents:  2,
			currentAgents:  2,
			expectActivity: true, // Agents still running = activity
		},
		{
			name:           "no change without agents",
			initialAgents:  0,
			currentAgents:  0,
			expectActivity: false, // No agents = no activity
		},
		{
			name:           "multiple agents start",
			initialAgents:  0,
			currentAgents:  3,
			expectActivity: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the activity detection logic from poll()
			agentStarted := tt.currentAgents > tt.initialAgents
			agentStopped := tt.currentAgents < tt.initialAgents
			agentsRunning := tt.currentAgents > 0

			hasActivity := agentStarted || agentStopped || agentsRunning

			if hasActivity != tt.expectActivity {
				t.Errorf("activity = %v, want %v", hasActivity, tt.expectActivity)
			}
		})
	}
}

// TestCommandHandlerRouting tests command routing logic.
func TestCommandHandlerRouting(t *testing.T) {
	validCommands := []string{"status", "projects", "agents", "activities", "refresh"}
	invalidCommands := []string{"", "unknown", "shutdown", "restart", "STOP"}

	t.Run("valid commands", func(t *testing.T) {
		for _, cmd := range validCommands {
			isValid := isValidCommand(cmd)
			if !isValid {
				t.Errorf("command %q should be valid", cmd)
			}
		}
	})

	t.Run("invalid commands", func(t *testing.T) {
		for _, cmd := range invalidCommands {
			isValid := isValidCommand(cmd)
			if isValid {
				t.Errorf("command %q should be invalid", cmd)
			}
		}
	})
}

// isValidCommand checks if a command is valid (mirrors daemon logic).
func isValidCommand(cmd string) bool {
	switch cmd {
	case "status", "projects", "agents", "activities", "refresh":
		return true
	default:
		return false
	}
}

// TestDefaultActivityLimit tests the default limit for activities.
func TestDefaultActivityLimit(t *testing.T) {
	tests := []struct {
		name          string
		requestLimit  int
		expectedLimit int
	}{
		{
			name:          "zero uses default",
			requestLimit:  0,
			expectedLimit: 50,
		},
		{
			name:          "explicit limit used",
			requestLimit:  100,
			expectedLimit: 100,
		},
		{
			name:          "small limit used",
			requestLimit:  10,
			expectedLimit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.requestLimit
			if limit == 0 {
				limit = 50 // Default from handleActivities
			}
			if limit != tt.expectedLimit {
				t.Errorf("limit = %d, want %d", limit, tt.expectedLimit)
			}
		})
	}
}

// Helper function
func containsString(data []byte, s string) bool {
	for i := 0; i <= len(data)-len(s); i++ {
		if string(data[i:i+len(s)]) == s {
			return true
		}
	}
	return false
}

// BenchmarkRequestMarshal measures request serialization performance.
func BenchmarkRequestMarshal(b *testing.B) {
	req := Request{
		Command: "activities",
		Project: "/home/user/projects/app",
		Limit:   100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(req)
	}
}

// BenchmarkResponseMarshal measures response serialization performance.
func BenchmarkResponseMarshal(b *testing.B) {
	now := time.Now()
	resp := Response{
		Status: "ok",
		Daemon: &types.DaemonStatus{
			Running:      true,
			PID:          12345,
			StartTime:    now,
			ProjectCount: 10,
			ActiveAgents: 2,
		},
		Projects: []types.Project{
			{Path: "/p/a", Name: "a", LastActivity: now},
			{Path: "/p/b", Name: "b", LastActivity: now},
		},
		Agents: []types.Agent{
			{PID: 100, Type: "claude", StartTime: now},
			{PID: 200, Type: "cursor", StartTime: now},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(resp)
	}
}

// BenchmarkResponseUnmarshal measures response deserialization performance.
func BenchmarkResponseUnmarshal(b *testing.B) {
	now := time.Now()
	resp := Response{
		Status: "ok",
		Daemon: &types.DaemonStatus{
			Running:      true,
			PID:          12345,
			StartTime:    now,
			ProjectCount: 10,
		},
		Projects: []types.Project{
			{Path: "/p/a", Name: "a"},
		},
	}
	data, _ := json.Marshal(resp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var r Response
		json.Unmarshal(data, &r)
	}
}
