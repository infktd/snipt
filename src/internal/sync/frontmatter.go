package sync

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/infktd/snipt/src/internal/model"
)

var nonAlphanumericSlug = regexp.MustCompile(`[^a-z0-9\s-]`)
var whitespaceRun = regexp.MustCompile(`[-\s]+`)

// Slugify converts a title into a filename-safe slug with .md extension.
func Slugify(title string) string {
	s := strings.ToLower(title)
	s = nonAlphanumericSlug.ReplaceAllString(s, " ")
	s = whitespaceRun.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s + ".md"
}

// ToFrontmatter converts a Snippet into a frontmatter markdown string.
func ToFrontmatter(sn model.Snippet) string {
	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("title: %s\n", sn.Title))
	buf.WriteString(fmt.Sprintf("language: %s\n", sn.Language))
	if len(sn.Tags) > 0 {
		buf.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(sn.Tags, ", ")))
	}
	if sn.Description != "" {
		buf.WriteString(fmt.Sprintf("description: %s\n", sn.Description))
	}
	if sn.Pinned {
		buf.WriteString("pinned: true\n")
	}
	buf.WriteString("---\n")
	buf.WriteString(sn.Content)
	return buf.String()
}

// FromFrontmatter parses a frontmatter markdown string into a Snippet.
func FromFrontmatter(filename string, raw string) (model.Snippet, error) {
	if !strings.HasPrefix(raw, "---\n") {
		return model.Snippet{}, fmt.Errorf("%s: missing frontmatter header", filename)
	}

	rest := raw[4:]
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		return model.Snippet{}, fmt.Errorf("%s: unclosed frontmatter", filename)
	}

	header := rest[:idx]
	body := rest[idx+5:]

	var sn model.Snippet
	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		colonIdx := strings.Index(line, ": ")
		if colonIdx < 0 {
			continue
		}
		key := line[:colonIdx]
		val := line[colonIdx+2:]

		switch key {
		case "title":
			sn.Title = val
		case "language":
			sn.Language = val
		case "description":
			sn.Description = val
		case "pinned":
			sn.Pinned = val == "true"
		case "tags":
			val = strings.TrimPrefix(val, "[")
			val = strings.TrimSuffix(val, "]")
			if val != "" {
				parts := strings.Split(val, ", ")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						sn.Tags = append(sn.Tags, p)
					}
				}
			}
		}
	}

	sn.Content = body
	return sn, nil
}

// ComputeHash returns a truncated SHA-256 hash of the snippet's frontmatter content.
func ComputeHash(sn model.Snippet) string {
	content := ToFrontmatter(sn)
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", hash[:8])
}
