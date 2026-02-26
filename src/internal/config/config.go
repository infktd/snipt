package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds user preferences loaded from config.toml.
type Config struct {
	Editor          string `toml:"editor"`
	DefaultLanguage string `toml:"default_language"`
	Theme           string `toml:"theme"`
}

const defaultConfig = `editor = ""
default_language = "text"
theme = "catppuccin-mocha"
`

// Load reads the config file, creating it with defaults if it doesn't exist.
func Load() (*Config, error) {
	path := ConfigPath()

	cfg := &Config{
		DefaultLanguage: "text",
		Theme:           "catppuccin-mocha",
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(defaultConfig), 0o644); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ResolveEditor returns the editor to use, following the chain:
// config editor -> $VISUAL -> $EDITOR -> vi
func (c *Config) ResolveEditor() string {
	if c.Editor != "" {
		return c.Editor
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "vi"
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	cfgHome := os.Getenv("XDG_CONFIG_HOME")
	if cfgHome == "" {
		home, _ := os.UserHomeDir()
		cfgHome = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgHome, "snipt", "config.toml")
}

// DBPath returns the path to the SQLite database.
// If override is non-empty, it is returned directly.
func DBPath(override string) string {
	if override != "" {
		return override
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "snipt", "snipt.db")
}
