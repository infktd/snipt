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

func TestLoad_NewSections(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "snipt")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
editor = "nvim"
default_language = "go"
theme = "catppuccin-mocha"

[general]
hotkey = "cmd+shift+s"

[find]
preview = true
sort = "alpha"
copy_to_clipboard = false
`), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.General.Hotkey != "cmd+shift+s" {
		t.Errorf("expected hotkey=cmd+shift+s, got %q", cfg.General.Hotkey)
	}
	if cfg.Find.Preview != true {
		t.Error("expected find.preview=true")
	}
	if cfg.Find.Sort != "alpha" {
		t.Errorf("expected find.sort=alpha, got %q", cfg.Find.Sort)
	}
	if cfg.Find.CopyToClipboard != false {
		t.Error("expected find.copy_to_clipboard=false")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.General.Hotkey != "cmd+shift+s" {
		t.Errorf("expected default hotkey=cmd+shift+s, got %q", cfg.General.Hotkey)
	}
	if cfg.Find.Sort != "recent" {
		t.Errorf("expected default sort=recent, got %q", cfg.Find.Sort)
	}
	if cfg.Find.CopyToClipboard != true {
		t.Error("expected default copy_to_clipboard=true")
	}
}

func TestLoad_SyncSection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfgDir := filepath.Join(dir, "snipt")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(`
editor = "nvim"

[sync]
gist_id = "abc123"
token = "ghp_test"
last_sync = "2026-02-27T12:00:00Z"
username = "testuser"
`), 0o644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Sync.GistID != "abc123" {
		t.Errorf("expected gist_id=abc123, got %q", cfg.Sync.GistID)
	}
	if cfg.Sync.Token != "ghp_test" {
		t.Errorf("expected token=ghp_test, got %q", cfg.Sync.Token)
	}
	if cfg.Sync.LastSync != "2026-02-27T12:00:00Z" {
		t.Errorf("expected last_sync=2026-02-27T12:00:00Z, got %q", cfg.Sync.LastSync)
	}
	if cfg.Sync.Username != "testuser" {
		t.Errorf("expected username=testuser, got %q", cfg.Sync.Username)
	}
}

func TestSave_SyncSection(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := DefaultConfig()
	cfg.Sync.GistID = "def456"
	cfg.Sync.Token = "ghp_save"
	cfg.Sync.LastSync = "2026-02-27T14:00:00Z"
	cfg.Sync.Username = "saveuser"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Sync.GistID != "def456" {
		t.Errorf("expected gist_id=def456 after save, got %q", loaded.Sync.GistID)
	}
	if loaded.Sync.Token != "ghp_save" {
		t.Errorf("expected token=ghp_save after save, got %q", loaded.Sync.Token)
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := DefaultConfig()
	cfg.Editor = "code"
	cfg.Find.Sort = "alpha"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Editor != "code" {
		t.Errorf("expected editor=code after save, got %q", loaded.Editor)
	}
	if loaded.Find.Sort != "alpha" {
		t.Errorf("expected find.sort=alpha after save, got %q", loaded.Find.Sort)
	}
	if loaded.Find.CopyToClipboard != true {
		t.Error("expected copy_to_clipboard preserved after save")
	}
}
