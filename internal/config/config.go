package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultServerURL = "http://localhost:1888"
	ConfigFileName   = "config.json"
	ConfigDirName    = "certvaultclix"
)

// Config holds the application configuration.
type Config struct {
	ServerURL string `json:"server_url"`
	Session   string `json:"session,omitempty"`
}

var configPath string

func init() {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	configPath = filepath.Join(dir, ConfigDirName, ConfigFileName)
}

// Load reads config from disk (or returns defaults).
func Load() (*Config, error) {
	cfg := &Config{ServerURL: DefaultServerURL}

	// Environment variable overrides
	if url := os.Getenv("CERTVAULT_URL"); url != "" {
		cfg.ServerURL = url
	}
	if session := os.Getenv("CERTVAULT_SESSION"); session != "" {
		cfg.Session = session
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	// Environment variable overrides take priority over file
	if url := os.Getenv("CERTVAULT_URL"); url != "" {
		cfg.ServerURL = url
	}
	if session := os.Getenv("CERTVAULT_SESSION"); session != "" {
		cfg.Session = session
	}

	return cfg, nil
}

// Save writes config to disk.
func Save(cfg *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// Path returns the configuration file path.
func Path() string {
	return configPath
}
