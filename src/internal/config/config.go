package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds user preferences loaded from config.toml.
type Config struct {
	Editor          string `toml:"editor"`
	DefaultLanguage string `toml:"default_language"`
	Theme           string `toml:"theme"`

	General GeneralConfig `toml:"general"`
	Find    FindConfig    `toml:"find"`
	Sync    SyncConfig    `toml:"sync"`
}

// GeneralConfig holds general app settings.
type GeneralConfig struct {
	Hotkey string `toml:"hotkey"`
}

// FindConfig holds find palette preferences.
type FindConfig struct {
	Preview         bool   `toml:"preview"`
	Sort            string `toml:"sort"`
	CopyToClipboard bool   `toml:"copy_to_clipboard"`
}

// SyncConfig holds GitHub Gist sync settings.
type SyncConfig struct {
	GistID   string `toml:"gist_id"`
	Token    string `toml:"token"`
	LastSync string `toml:"last_sync"`
	Username string `toml:"username"`
}

const defaultConfig = `editor = ""
default_language = "text"
theme = "catppuccin-mocha"

[general]
hotkey = "cmd+shift+s"

[find]
preview = false
sort = "recent"
copy_to_clipboard = true
`

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultLanguage: "text",
		Theme:           "catppuccin-mocha",
		General: GeneralConfig{
			Hotkey: "cmd+shift+s",
		},
		Find: FindConfig{
			Sort:            "recent",
			CopyToClipboard: true,
		},
	}
}

// Load reads the config file, creating it with defaults if it doesn't exist.
func Load() (*Config, error) {
	path := ConfigPath()
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(defaultConfig), 0o600); err != nil {
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

// Save writes the config to disk at ConfigPath().
func (c *Config) Save() error {
	path := ConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(c); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
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
