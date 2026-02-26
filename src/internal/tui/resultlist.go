package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/model"
)

// ResultItem pairs a snippet with its fuzzy match result for display.
type ResultItem struct {
	Snippet     model.Snippet
	FuzzyResult FuzzyResult
}

// ResultList is a reusable navigable list of snippet results.
// Used by both the find palette and the mini-picker.
type ResultList struct {
	items  []ResultItem
	cursor int
	width  int
	height int // max visible rows
}

// NewResultList creates a new result list with the given dimensions.
func NewResultList(width, height int) ResultList {
	return ResultList{
		width:  width,
		height: height,
	}
}

// SetItems replaces the current items and resets the cursor.
func (r *ResultList) SetItems(items []ResultItem) {
	r.items = items
	r.cursor = 0
}

// Selected returns the currently highlighted item, or nil if the list is empty.
func (r *ResultList) Selected() *ResultItem {
	if len(r.items) == 0 {
		return nil
	}
	return &r.items[r.cursor]
}

// Len returns the number of items in the list.
func (r *ResultList) Len() int {
	return len(r.items)
}

// Update handles keyboard navigation (up/down with wrapping).
func (r ResultList) Update(msg tea.Msg) (ResultList, tea.Cmd) {
	if len(r.items) == 0 {
		return r, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "ctrl+p":
			r.cursor--
			if r.cursor < 0 {
				r.cursor = len(r.items) - 1
			}
		case "down", "ctrl+n":
			r.cursor++
			if r.cursor >= len(r.items) {
				r.cursor = 0
			}
		}
	}

	return r, nil
}

// View renders the list with highlighted characters and inline preview for the selected row.
func (r ResultList) View() string {
	if len(r.items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Width(r.width).
			Align(lipgloss.Center).
			Padding(1, 0)
		return emptyStyle.Render("No snippets found")
	}

	// Determine the visible window of items.
	visibleCount := r.height
	if visibleCount > len(r.items) {
		visibleCount = len(r.items)
	}

	start := 0
	if r.cursor >= visibleCount {
		start = r.cursor - visibleCount + 1
	}
	end := start + visibleCount
	if end > len(r.items) {
		end = len(r.items)
		start = end - visibleCount
		if start < 0 {
			start = 0
		}
	}

	// Width for the main row content vs. the preview pane.
	previewWidth := r.width * 2 / 5
	if previewWidth > 40 {
		previewWidth = 40
	}
	rowWidth := r.width - previewWidth - 3 // 3 for separator and padding

	var rows []string
	for i := start; i < end; i++ {
		item := r.items[i]
		selected := i == r.cursor
		row := r.renderRow(item, selected, rowWidth)

		if selected {
			preview := r.renderPreview(item.Snippet, previewWidth)
			// Join row and preview side-by-side.
			sep := lipgloss.NewStyle().
				Foreground(ColorBorderDim).
				Render(" | ")
			combined := lipgloss.JoinHorizontal(lipgloss.Top, row, sep, preview)
			rows = append(rows, combined)
		} else {
			// Pad non-selected rows to the full width.
			rows = append(rows, row)
		}
	}

	return strings.Join(rows, "\n")
}

// renderRow renders a single list row.
func (r ResultList) renderRow(item ResultItem, selected bool, width int) string {
	sn := item.Snippet

	// Build the row content: [pin] [title] [lang] [tags...]
	var parts []string

	// Pinned indicator.
	if sn.Pinned {
		pinStyle := lipgloss.NewStyle().Foreground(ColorPink)
		parts = append(parts, pinStyle.Render("\u25cf")) // filled circle
	} else {
		parts = append(parts, " ")
	}

	// Title with fuzzy highlight.
	title := renderFuzzyTitle(sn.Title, item.FuzzyResult.Indices, selected)
	parts = append(parts, title)

	// Language badge.
	if sn.Language != "" {
		langStyle := lipgloss.NewStyle().Foreground(LanguageColor(sn.Language))
		parts = append(parts, langStyle.Render(sn.Language))
	}

	// Tags — on non-selected rows, truncate to keep everything on one line.
	if len(sn.Tags) > 0 {
		tagStyle := lipgloss.NewStyle().Foreground(ColorTextDim)
		if selected {
			for _, tag := range sn.Tags {
				parts = append(parts, tagStyle.Render("#"+tag))
			}
		} else {
			// Calculate remaining space for tags.
			prefix := strings.Join(parts, " ")
			used := lipgloss.Width(prefix)
			for _, tag := range sn.Tags {
				rendered := tagStyle.Render("#" + tag)
				needed := lipgloss.Width(rendered) + 1 // +1 for space separator
				if used+needed > width {
					break
				}
				parts = append(parts, rendered)
				used += needed
			}
		}
	}

	content := strings.Join(parts, " ")

	// Apply row-level styling.
	rowStyle := lipgloss.NewStyle().Width(width).MaxWidth(width)
	if selected {
		rowStyle = rowStyle.
			Background(ColorBgSelected).
			Bold(true)
	}

	return rowStyle.Render(content)
}

// renderFuzzyTitle renders the title with matched characters highlighted.
func renderFuzzyTitle(title string, indices []int, selected bool) string {
	if len(indices) == 0 {
		style := lipgloss.NewStyle().Foreground(ColorText)
		return style.Render(title)
	}

	matchSet := make(map[int]bool, len(indices))
	for _, idx := range indices {
		matchSet[idx] = true
	}

	matchStyle := lipgloss.NewStyle().Foreground(ColorPink).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(ColorText)

	var out strings.Builder
	for i, ch := range title {
		if matchSet[i] {
			out.WriteString(matchStyle.Render(string(ch)))
		} else {
			out.WriteString(normalStyle.Render(string(ch)))
		}
	}
	return out.String()
}

// renderPreview renders a truncated code preview for the selected row.
func (r ResultList) renderPreview(sn model.Snippet, width int) string {
	if sn.Content == "" {
		return ""
	}

	previewStyle := lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Width(width)

	lines := strings.Split(sn.Content, "\n")
	maxLines := 3
	if maxLines > len(lines) {
		maxLines = len(lines)
	}

	var preview []string
	for _, line := range lines[:maxLines] {
		// Truncate long lines.
		if len(line) > width {
			line = line[:width-1] + "\u2026" // ellipsis
		}
		preview = append(preview, line)
	}

	if len(lines) > maxLines {
		preview = append(preview, fmt.Sprintf("  ... +%d lines", len(lines)-maxLines))
	}

	return previewStyle.Render(strings.Join(preview, "\n"))
}
