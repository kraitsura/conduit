package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the global conduit configuration
type Config struct {
	// RootPath is the projects directory being watched
	RootPath string `json:"root_path"`

	// PollInterval is how often the daemon checks for changes (seconds)
	PollInterval int `json:"poll_interval"`

	// AgentPatterns are process name patterns to detect as AI agents
	AgentPatterns []string `json:"agent_patterns"`

	// SocketPath is the Unix socket for daemon communication
	SocketPath string `json:"socket_path"`

	// DBPath is the SQLite database location
	DBPath string `json:"db_path"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	conduitDir := filepath.Join(home, ".conduit")

	return &Config{
		RootPath:     filepath.Join(home, "Projects"),
		PollInterval: 5,
		AgentPatterns: []string{
			"claude",
			"cursor",
			"Cursor",
			"aider",
			"copilot",
			"continue",
			"cody",
		},
		SocketPath: filepath.Join(conduitDir, "conduit.sock"),
		DBPath:     filepath.Join(conduitDir, "conduit.db"),
	}
}

// ConfigDir returns the conduit config directory path
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".conduit")
}

// ConfigPath returns the config file path
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// Load reads config from disk, or returns defaults if not found
func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Use defaults
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes config to disk
func (c *Config) Save() error {
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// EnsureDirectories creates necessary directories
func EnsureDirectories() error {
	return os.MkdirAll(ConfigDir(), 0755)
}
