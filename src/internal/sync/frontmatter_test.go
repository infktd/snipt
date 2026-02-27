package sync

import (
	"strings"
	"testing"

	"github.com/infktd/snipt/src/internal/model"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"HTTP server with middleware", "http-server-with-middleware.md"},
		{"retry with exponential backoff", "retry-with-exponential-backoff.md"},
		{"nix flake dev shell", "nix-flake-dev-shell.md"},
		{"hello world!", "hello-world.md"},
		{"  spaces  around  ", "spaces-around.md"},
		{"UPPERCASE TITLE", "uppercase-title.md"},
		{"special@#$chars", "special-chars.md"},
		{"", ".md"},
	}

	for _, tt := range tests {
		got := Slugify(tt.input)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToFrontmatter(t *testing.T) {
	sn := model.Snippet{
		Title:       "HTTP Server",
		Language:    "go",
		Tags:        []string{"web", "http"},
		Description: "A simple HTTP server",
		Pinned:      true,
		Content:     "package main\nfunc main() {}",
	}

	got := ToFrontmatter(sn)

	if !strings.HasPrefix(got, "---\n") {
		t.Error("expected frontmatter to start with ---")
	}
	if !strings.Contains(got, "title: HTTP Server\n") {
		t.Error("expected title field")
	}
	if !strings.Contains(got, "language: go\n") {
		t.Error("expected language field")
	}
	if !strings.Contains(got, "tags: [web, http]\n") {
		t.Error("expected tags field")
	}
	if !strings.Contains(got, "description: A simple HTTP server\n") {
		t.Error("expected description field")
	}
	if !strings.Contains(got, "pinned: true\n") {
		t.Error("expected pinned field")
	}
	if strings.Contains(got, "pinned: false") {
		t.Error("pinned: false should be omitted")
	}
	if !strings.HasSuffix(got, "package main\nfunc main() {}") {
		t.Error("expected content after frontmatter")
	}
}

func TestToFrontmatter_MinimalFields(t *testing.T) {
	sn := model.Snippet{
		Title:    "Minimal",
		Language: "text",
		Content:  "hello",
	}

	got := ToFrontmatter(sn)

	if strings.Contains(got, "tags:") {
		t.Error("tags should be omitted when empty")
	}
	if strings.Contains(got, "description:") {
		t.Error("description should be omitted when empty")
	}
	if strings.Contains(got, "pinned:") {
		t.Error("pinned should be omitted when false")
	}
}

func TestFromFrontmatter(t *testing.T) {
	content := "---\ntitle: HTTP Server\nlanguage: go\ntags: [web, http]\ndescription: A simple HTTP server\npinned: true\n---\npackage main\nfunc main() {}"

	sn, err := FromFrontmatter("http-server.md", content)
	if err != nil {
		t.Fatalf("FromFrontmatter() error: %v", err)
	}

	if sn.Title != "HTTP Server" {
		t.Errorf("Title = %q, want %q", sn.Title, "HTTP Server")
	}
	if sn.Language != "go" {
		t.Errorf("Language = %q, want %q", sn.Language, "go")
	}
	if len(sn.Tags) != 2 || sn.Tags[0] != "web" || sn.Tags[1] != "http" {
		t.Errorf("Tags = %v, want [web http]", sn.Tags)
	}
	if sn.Description != "A simple HTTP server" {
		t.Errorf("Description = %q, want %q", sn.Description, "A simple HTTP server")
	}
	if !sn.Pinned {
		t.Error("Pinned = false, want true")
	}
	if sn.Content != "package main\nfunc main() {}" {
		t.Errorf("Content = %q, want %q", sn.Content, "package main\nfunc main() {}")
	}
}

func TestFromFrontmatter_NoPinned(t *testing.T) {
	content := "---\ntitle: Simple\nlanguage: text\n---\nhello world"

	sn, err := FromFrontmatter("simple.md", content)
	if err != nil {
		t.Fatalf("FromFrontmatter() error: %v", err)
	}

	if sn.Pinned {
		t.Error("Pinned should default to false")
	}
	if sn.Content != "hello world" {
		t.Errorf("Content = %q, want %q", sn.Content, "hello world")
	}
}

func TestFromFrontmatter_InvalidFormat(t *testing.T) {
	_, err := FromFrontmatter("bad.md", "no frontmatter here")
	if err == nil {
		t.Error("expected error for content without frontmatter")
	}
}

func TestRoundtrip(t *testing.T) {
	original := model.Snippet{
		Title:       "Roundtrip Test",
		Language:    "python",
		Tags:        []string{"test", "roundtrip"},
		Description: "Testing roundtrip conversion",
		Pinned:      true,
		Content:     "print('hello')\nprint('world')",
	}

	md := ToFrontmatter(original)
	parsed, err := FromFrontmatter("roundtrip-test.md", md)
	if err != nil {
		t.Fatalf("FromFrontmatter() error: %v", err)
	}

	if parsed.Title != original.Title {
		t.Errorf("Title = %q, want %q", parsed.Title, original.Title)
	}
	if parsed.Language != original.Language {
		t.Errorf("Language = %q, want %q", parsed.Language, original.Language)
	}
	if len(parsed.Tags) != len(original.Tags) {
		t.Errorf("Tags count = %d, want %d", len(parsed.Tags), len(original.Tags))
	}
	if parsed.Description != original.Description {
		t.Errorf("Description = %q, want %q", parsed.Description, original.Description)
	}
	if parsed.Pinned != original.Pinned {
		t.Errorf("Pinned = %v, want %v", parsed.Pinned, original.Pinned)
	}
	if parsed.Content != original.Content {
		t.Errorf("Content = %q, want %q", parsed.Content, original.Content)
	}
}

func TestComputeHash(t *testing.T) {
	sn := model.Snippet{
		Title:    "Test",
		Language: "go",
		Content:  "hello",
	}

	hash1 := ComputeHash(sn)
	hash2 := ComputeHash(sn)

	if hash1 != hash2 {
		t.Error("same snippet should produce same hash")
	}

	if !strings.HasPrefix(hash1, "sha256:") {
		t.Errorf("hash should start with sha256:, got %q", hash1)
	}

	sn.Content = "changed"
	hash3 := ComputeHash(sn)
	if hash3 == hash1 {
		t.Error("different content should produce different hash")
	}
}
