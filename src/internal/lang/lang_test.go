package lang

import "testing"

func TestFromExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"script.py", "python"},
		{"app.ts", "typescript"},
		{"index.js", "javascript"},
		{"lib.rs", "rust"},
		{"init.lua", "lua"},
		{"deploy.sh", "bash"},
		{"setup.bash", "bash"},
		{"query.sql", "sql"},
		{"flake.nix", "nix"},
		{"Gemfile.rb", "ruby"},
		{"Main.java", "java"},
		{"util.c", "c"},
		{"util.cpp", "cpp"},
		{"README.md", "markdown"},
		{"config.toml", "toml"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"data.json", "json"},
		{"noext", ""},
		{"weird.xyz", ""},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := FromExtension(tt.filename)
			if got != tt.want {
				t.Errorf("FromExtension(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
