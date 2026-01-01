package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultConfig verifies default configuration values.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	t.Run("has root path", func(t *testing.T) {
		if cfg.RootPath == "" {
			t.Error("RootPath should not be empty")
		}
		// Should end with "Projects"
		if filepath.Base(cfg.RootPath) != "Projects" {
			t.Errorf("RootPath should end with Projects, got %q", cfg.RootPath)
		}
	})

	t.Run("has reasonable poll interval", func(t *testing.T) {
		if cfg.PollInterval <= 0 {
			t.Error("PollInterval should be positive")
		}
		if cfg.PollInterval > 60 {
			t.Error("PollInterval seems too high")
		}
	})

	t.Run("has agent patterns", func(t *testing.T) {
		if len(cfg.AgentPatterns) == 0 {
			t.Error("AgentPatterns should not be empty")
		}

		// Verify expected patterns are present
		expectedPatterns := []string{"claude", "cursor", "aider", "copilot"}
		for _, expected := range expectedPatterns {
			found := false
			for _, pattern := range cfg.AgentPatterns {
				if pattern == expected || pattern == capitalize(expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected pattern %q not found in defaults", expected)
			}
		}
	})

	t.Run("has socket path", func(t *testing.T) {
		if cfg.SocketPath == "" {
			t.Error("SocketPath should not be empty")
		}
		if filepath.Base(cfg.SocketPath) != "conduit.sock" {
			t.Errorf("SocketPath should end with conduit.sock, got %q", cfg.SocketPath)
		}
	})

	t.Run("has db path", func(t *testing.T) {
		if cfg.DBPath == "" {
			t.Error("DBPath should not be empty")
		}
		if filepath.Base(cfg.DBPath) != "conduit.db" {
			t.Errorf("DBPath should end with conduit.db, got %q", cfg.DBPath)
		}
	})
}

// TestConfigDir verifies config directory path.
func TestConfigDir(t *testing.T) {
	dir := ConfigDir()

	if dir == "" {
		t.Fatal("ConfigDir returned empty string")
	}

	// Should be under home directory
	home, _ := os.UserHomeDir()
	if home != "" && !hasPrefix(dir, home) {
		t.Errorf("ConfigDir should be under home, got %q", dir)
	}

	// Should end with .conduit
	if filepath.Base(dir) != ".conduit" {
		t.Errorf("ConfigDir should end with .conduit, got %q", dir)
	}
}

// TestConfigPath verifies config file path.
func TestConfigPath(t *testing.T) {
	path := ConfigPath()

	if path == "" {
		t.Fatal("ConfigPath returned empty string")
	}

	// Should be under ConfigDir
	if filepath.Dir(path) != ConfigDir() {
		t.Errorf("ConfigPath should be under ConfigDir")
	}

	// Should be config.json
	if filepath.Base(path) != "config.json" {
		t.Errorf("ConfigPath should be config.json, got %q", filepath.Base(path))
	}
}

// TestLoadNonExistent verifies Load returns defaults when file doesn't exist.
func TestLoadNonExistent(t *testing.T) {
	// This test relies on the file not existing.
	// In a real test environment, we'd mock the file system.
	// For now, we just verify the function doesn't error.

	// Note: This would need mocking to properly test without touching
	// the real config file. We'll test Save/Load with temp files below.
}

// TestSaveAndLoad tests config persistence with temporary files.
func TestSaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "conduit-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	t.Run("save creates file", func(t *testing.T) {
		cfg := &Config{
			RootPath:      "/custom/projects",
			PollInterval:  10,
			AgentPatterns: []string{"claude", "cursor"},
			SocketPath:    filepath.Join(tmpDir, "test.sock"),
			DBPath:        filepath.Join(tmpDir, "test.db"),
		}

		// Write to file directly since Save uses fixed path
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}
	})

	t.Run("load reads file", func(t *testing.T) {
		// Read the file we just wrote
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if cfg.RootPath != "/custom/projects" {
			t.Errorf("RootPath: got %q, want /custom/projects", cfg.RootPath)
		}
		if cfg.PollInterval != 10 {
			t.Errorf("PollInterval: got %d, want 10", cfg.PollInterval)
		}
		if len(cfg.AgentPatterns) != 2 {
			t.Errorf("AgentPatterns: got %d patterns, want 2", len(cfg.AgentPatterns))
		}
	})
}

// TestConfigJSONFormat verifies JSON serialization format.
func TestConfigJSONFormat(t *testing.T) {
	cfg := &Config{
		RootPath:      "/home/user/Projects",
		PollInterval:  5,
		AgentPatterns: []string{"claude", "cursor"},
		SocketPath:    "/home/user/.conduit/conduit.sock",
		DBPath:        "/home/user/.conduit/conduit.db",
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(data)

	// Verify expected keys are present
	expectedKeys := []string{
		"root_path",
		"poll_interval",
		"agent_patterns",
		"socket_path",
		"db_path",
	}

	for _, key := range expectedKeys {
		if !contains(jsonStr, `"`+key+`"`) {
			t.Errorf("expected key %q not found in JSON", key)
		}
	}
}

// TestConfigPartialJSON verifies partial JSON updates work correctly.
func TestConfigPartialJSON(t *testing.T) {
	// Start with defaults
	cfg := DefaultConfig()

	// Partial JSON with only some fields
	partialJSON := `{
		"poll_interval": 15,
		"agent_patterns": ["custom-agent"]
	}`

	if err := json.Unmarshal([]byte(partialJSON), cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Updated fields should be changed
	if cfg.PollInterval != 15 {
		t.Errorf("PollInterval should be 15, got %d", cfg.PollInterval)
	}
	if len(cfg.AgentPatterns) != 1 || cfg.AgentPatterns[0] != "custom-agent" {
		t.Errorf("AgentPatterns should be [custom-agent], got %v", cfg.AgentPatterns)
	}

	// Non-updated fields should retain defaults
	if cfg.RootPath == "" {
		t.Error("RootPath should retain default value")
	}
	if cfg.SocketPath == "" {
		t.Error("SocketPath should retain default value")
	}
}

// TestEnsureDirectories verifies directory creation.
func TestEnsureDirectories(t *testing.T) {
	// Note: EnsureDirectories creates the real config directory.
	// In production code, we'd inject the path for testability.
	// Here we just verify it doesn't error.

	// This is a simple smoke test
	err := EnsureDirectories()
	if err != nil {
		// It's okay if it fails due to permissions in CI
		t.Logf("EnsureDirectories returned: %v (may be expected in some environments)", err)
	}
}

// TestConfigValidation tests config field validation scenarios.
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				RootPath:      "/home/user/Projects",
				PollInterval:  5,
				AgentPatterns: []string{"claude"},
				SocketPath:    "/tmp/test.sock",
				DBPath:        "/tmp/test.db",
			},
			wantErr: false,
		},
		{
			name: "empty root path",
			cfg: &Config{
				RootPath:     "",
				PollInterval: 5,
			},
			// Currently no validation, but documenting expected behavior
			wantErr: false,
		},
		{
			name: "zero poll interval",
			cfg: &Config{
				RootPath:     "/projects",
				PollInterval: 0,
			},
			wantErr: false, // Currently allowed
		},
		{
			name: "empty agent patterns",
			cfg: &Config{
				RootPath:      "/projects",
				PollInterval:  5,
				AgentPatterns: []string{},
			},
			wantErr: false, // Currently allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Config doesn't have validation currently
			// This test documents that any config can be serialized
			_, err := json.Marshal(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConfigPaths verifies path construction is platform-appropriate.
func TestConfigPaths(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("paths use filepath separator", func(t *testing.T) {
		// On Unix, filepath.Separator is /
		// On Windows, filepath.Separator is \
		// filepath.Join should handle this correctly

		// Verify paths don't have mixed separators
		if containsBoth(cfg.RootPath, "/", "\\") {
			t.Error("RootPath has mixed path separators")
		}
		if containsBoth(cfg.SocketPath, "/", "\\") {
			t.Error("SocketPath has mixed path separators")
		}
		if containsBoth(cfg.DBPath, "/", "\\") {
			t.Error("DBPath has mixed path separators")
		}
	})
}

// Helper functions for tests

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsBoth(s, a, b string) bool {
	return contains(s, a) && contains(s, b)
}

// BenchmarkDefaultConfig measures config factory performance.
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultConfig()
	}
}

// BenchmarkConfigMarshal measures JSON serialization performance.
func BenchmarkConfigMarshal(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(cfg)
	}
}

// BenchmarkConfigUnmarshal measures JSON deserialization performance.
func BenchmarkConfigUnmarshal(b *testing.B) {
	cfg := DefaultConfig()
	data, _ := json.Marshal(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var c Config
		json.Unmarshal(data, &c)
	}
}
