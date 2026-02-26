package lang

import (
	"path/filepath"
	"strings"
)

var extMap = map[string]string{
	".go":   "go",
	".py":   "python",
	".ts":   "typescript",
	".js":   "javascript",
	".rs":   "rust",
	".lua":  "lua",
	".sh":   "bash",
	".bash": "bash",
	".sql":  "sql",
	".nix":  "nix",
	".rb":   "ruby",
	".java": "java",
	".c":    "c",
	".cpp":  "cpp",
	".md":   "markdown",
	".toml": "toml",
	".yaml": "yaml",
	".yml":  "yaml",
	".json": "json",
}

// FromExtension returns the language for a filename based on its extension.
// Returns empty string if the extension is not recognized.
func FromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return extMap[ext]
}
