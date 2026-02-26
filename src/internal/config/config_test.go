package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Editor != "" {
		t.Errorf("expected empty editor default, got %q", cfg.Editor)
	}
	if cfg.DefaultLanguage != "text" {
		t.Errorf("expected default_language=text, got %q", cfg.DefaultLanguage)
	}
	if cfg.Theme != "catppuccin-mocha" {
		t.Errorf("expected theme=catppuccin-mocha, got %q", cfg.Theme)
	}

	cfgFile := filepath.Join(dir, "snipt", "config.toml")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		t.Error("expected config file to be created")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "snipt")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
editor = "code"
default_language = "go"
theme = "dracula"
`), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Editor != "code" {
		t.Errorf("expected editor=code, got %q", cfg.Editor)
	}
	if cfg.DefaultLanguage != "go" {
		t.Errorf("expected default_language=go, got %q", cfg.DefaultLanguage)
	}
}

func TestResolveEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	cfg := &Config{Editor: "nvim"}
	if got := cfg.ResolveEditor(); got != "nvim" {
		t.Errorf("expected nvim from config, got %q", got)
	}

	cfg.Editor = ""
	t.Setenv("VISUAL", "code")
	if got := cfg.ResolveEditor(); got != "code" {
		t.Errorf("expected code from $VISUAL, got %q", got)
	}

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nano")
	if got := cfg.ResolveEditor(); got != "nano" {
		t.Errorf("expected nano from $EDITOR, got %q", got)
	}

	t.Setenv("EDITOR", "")
	if got := cfg.ResolveEditor(); got != "vi" {
		t.Errorf("expected vi fallback, got %q", got)
	}
}

func TestDBPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)

	path := DBPath("")
	expected := filepath.Join(dir, "snipt", "snipt.db")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestDBPath_Override(t *testing.T) {
	override := "/tmp/custom.db"
	path := DBPath(override)
	if path != override {
		t.Errorf("expected override %q, got %q", override, path)
	}
}

func TestConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path := ConfigPath()
	expected := filepath.Join(dir, "snipt", "config.toml")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
