package cli

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/spf13/cobra"
)

// exportEnvelope is the JSON export envelope format.
type exportEnvelope struct {
	Version    int             `json:"version"`
	ExportedAt string          `json:"exported_at"`
	Count      int             `json:"count"`
	Snippets   []model.Snippet `json:"snippets"`
}

func newExportCmd() *cobra.Command {
	var (
		format     string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export snippets",
		RunE: func(cmd *cobra.Command, args []string) error {
			snippets, err := store.List(db.ListOpts{})
			if err != nil {
				return fmt.Errorf("list snippets: %w", err)
			}

			var out io.Writer = cmd.OutOrStdout()
			if outputFile != "" {
				f, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("create output file: %w", err)
				}
				defer f.Close()
				out = f
			}

			switch format {
			case "json":
				return exportJSON(out, snippets)
			case "markdown":
				return exportMarkdown(out, snippets)
			case "tar":
				if outputFile == "" {
					return fmt.Errorf("tar format requires -o output file")
				}
				return exportTar(outputFile, snippets)
			default:
				return fmt.Errorf("unsupported format: %s (use json, markdown, or tar)", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "export format: json, markdown, tar")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path")

	return cmd
}

func exportJSON(out io.Writer, snippets []model.Snippet) error {
	envelope := exportEnvelope{
		Version:    1,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Count:      len(snippets),
		Snippets:   snippets,
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(envelope)
}

func exportMarkdown(out io.Writer, snippets []model.Snippet) error {
	for i, s := range snippets {
		if i > 0 {
			fmt.Fprintln(out, "---")
			fmt.Fprintln(out)
		}

		title := s.Title
		if title == "" {
			title = s.ID
		}
		fmt.Fprintf(out, "# %s\n\n", title)

		fmt.Fprintf(out, "- **ID**: %s\n", s.ID)
		if s.Language != "" {
			fmt.Fprintf(out, "- **Language**: %s\n", s.Language)
		}
		if len(s.Tags) > 0 {
			fmt.Fprintf(out, "- **Tags**: %s\n", strings.Join(s.Tags, ", "))
		}
		if s.Description != "" {
			fmt.Fprintf(out, "- **Description**: %s\n", s.Description)
		}
		if s.Source != "" {
			fmt.Fprintf(out, "- **Source**: %s\n", s.Source)
		}
		if s.Pinned {
			fmt.Fprintln(out, "- **Pinned**: yes")
		}
		fmt.Fprintf(out, "- **Created**: %s\n", s.CreatedAt.Format(time.RFC3339))
		fmt.Fprintln(out)

		lang := s.Language
		fmt.Fprintf(out, "```%s\n%s\n```\n\n", lang, s.Content)
	}
	return nil
}

func exportTar(path string, snippets []model.Snippet) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create tar file: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, s := range snippets {
		slug := slugify(s.Title)
		if slug == "" {
			slug = "untitled"
		}
		filename := fmt.Sprintf("%s-%s.md", s.ID, slug)

		var buf strings.Builder

		// YAML frontmatter.
		buf.WriteString("---\n")
		fmt.Fprintf(&buf, "id: %s\n", s.ID)
		fmt.Fprintf(&buf, "title: %q\n", s.Title)
		fmt.Fprintf(&buf, "language: %s\n", s.Language)
		if len(s.Tags) > 0 {
			fmt.Fprintf(&buf, "tags: [%s]\n", strings.Join(s.Tags, ", "))
		}
		if s.Description != "" {
			fmt.Fprintf(&buf, "description: %q\n", s.Description)
		}
		if s.Source != "" {
			fmt.Fprintf(&buf, "source: %q\n", s.Source)
		}
		fmt.Fprintf(&buf, "pinned: %t\n", s.Pinned)
		fmt.Fprintf(&buf, "use_count: %d\n", s.UseCount)
		fmt.Fprintf(&buf, "created_at: %s\n", s.CreatedAt.Format(time.RFC3339))
		fmt.Fprintf(&buf, "updated_at: %s\n", s.UpdatedAt.Format(time.RFC3339))
		buf.WriteString("---\n\n")

		// Content as fenced code block.
		fmt.Fprintf(&buf, "```%s\n%s\n```\n", s.Language, s.Content)

		content := buf.String()

		hdr := &tar.Header{
			Name:    filename,
			Size:    int64(len(content)),
			Mode:    0o644,
			ModTime: s.UpdatedAt,
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("write tar header: %w", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return fmt.Errorf("write tar content: %w", err)
		}
	}

	return nil
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
