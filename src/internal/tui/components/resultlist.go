package components

import (
	"fmt"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/common"
)

// ResultItem pairs a snippet with its fuzzy match result for display.
type ResultItem struct {
	Snippet     model.Snippet
	FuzzyResult FuzzyResult
}

// ResultList is a reusable navigable list of snippet results.
// Used by both the find palette and the mini-picker.
type ResultList struct {
	items       []ResultItem
	cursor      int
	width       int
	height      int  // max visible rows
	showPreview bool // two-column preview layout for selected row
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

// View renders the list with highlighted characters and preview for the selected row.
func (r ResultList) View() string {
	if len(r.items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(common.ColorTextDim).
			Background(common.ColorBgSurface).
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

	var rows []string
	for i := start; i < end; i++ {
		item := r.items[i]
		selected := i == r.cursor

		if selected && r.showPreview {
			// Add a blank spacer line before the selected row (unless it's first).
			if len(rows) > 0 {
				rows = append(rows, "")
			}
			rows = append(rows, r.renderSelectedRow(item))
			// Add a blank spacer line after the selected row.
			rows = append(rows, "")
		} else {
			rows = append(rows, r.renderSingleRow(item, selected))
		}
	}

	// Trim trailing empty line if the selected row is last.
	if len(rows) > 0 && rows[len(rows)-1] == "" {
		rows = rows[:len(rows)-1]
	}

	return strings.Join(rows, "\n")
}

// renderSelectedRow renders a two-column selected row with a pink left
// accent bar that covers the full height. Instead of using lipgloss Border
// (which doesn't match heights well), we manually prepend a pink bar
// character to each output line.
func (r ResultList) renderSelectedRow(item ResultItem) string {
	sn := item.Snippet

	// Reserve 3 chars left (accent bar + breathing space + column gap)
	// and 4 chars right margin for breathing room.
	contentWidth := r.width - 3 - 4

	// ~40/60 split: narrower left column pushes the preview box closer.
	leftWidth := contentWidth * 2 / 5
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := contentWidth - leftWidth
	if rightWidth < 10 {
		rightWidth = 10
	}

	// -- LEFT COLUMN: every line has explicit bgSelected --
	leftLines := r.buildSelectedLeftLines(item, leftWidth)

	// -- RIGHT COLUMN: preview box --
	// Prepend one blank line to push the box down so the top border
	// aligns with the tags/meta line, not the title.
	rightTopPad := lipgloss.NewStyle().Width(rightWidth).Background(common.ColorBgSelected).Render("")
	rightStr := r.renderPreviewBox(sn, rightWidth)
	rightLines := append([]string{rightTopPad}, strings.Split(rightStr, "\n")...)

	// -- Equalize heights --
	maxH := len(leftLines)
	if len(rightLines) > maxH {
		maxH = len(rightLines)
	}

	// Pad left with bgSelected empty lines.
	for len(leftLines) < maxH {
		leftLines = append(leftLines,
			lipgloss.NewStyle().Width(leftWidth).Background(common.ColorBgSelected).Render(""))
	}

	// Pad right if needed.
	for len(rightLines) < maxH {
		rightLines = append(rightLines,
			lipgloss.NewStyle().Width(rightWidth).Background(common.ColorBgSelected).Render(""))
	}

	// -- Compose each line: accent bar + left + gap + right --
	// Use a left-half-block for a visually thicker accent bar (1 cell wide
	// but renders as a solid pink slab).
	barStyle := lipgloss.NewStyle().Foreground(common.ColorPink).Background(common.ColorBgSelected)
	barAfter := lipgloss.NewStyle().Width(1).Background(common.ColorBgSelected) // breathing room
	gapStyle := lipgloss.NewStyle().Width(1).Background(common.ColorBgSelected)

	var finalLines []string
	for i := 0; i < maxH; i++ {
		bar := barStyle.Render("\u258c") + barAfter.Render(" ")
		gap := gapStyle.Render(" ")

		// Enforce width on each right-column line so bgSelected fills to
		// the panel edge (prevents the bottom-right border corner from
		// sitting on the dark base background).
		rLine := rightLines[i]
		rLineWidth := lipgloss.Width(rLine)
		if rLineWidth < rightWidth {
			pad := lipgloss.NewStyle().
				Width(rightWidth - rLineWidth).
				Background(common.ColorBgSelected).
				Render("")
			rLine = rLine + pad
		}

		line := bar + leftLines[i] + gap + rLine

		// Fill the right margin with bgSelected so the background
		// extends to the full panel width.
		lineWidth := lipgloss.Width(line)
		if lineWidth < r.width {
			rightFill := lipgloss.NewStyle().
				Width(r.width - lineWidth).
				Background(common.ColorBgSelected).
				Render("")
			line = line + rightFill
		}

		finalLines = append(finalLines, line)
	}

	return strings.Join(finalLines, "\n")
}

// buildSelectedLeftLines returns individual lines for the left column,
// each rendered with explicit bgSelected background and exact width.
func (r ResultList) buildSelectedLeftLines(item ResultItem, width int) []string {
	sn := item.Snippet

	// -- Line 1: [pin] title --
	var titleParts []string
	if sn.Pinned {
		titleParts = append(titleParts,
			lipgloss.NewStyle().Foreground(common.ColorPink).Background(common.ColorBgSelected).Render("\u2726"))
	} else {
		titleParts = append(titleParts,
			lipgloss.NewStyle().Background(common.ColorBgSelected).Render(" "))
	}

	title := common.RenderFuzzyTitleWithBg(sn.Title, item.FuzzyResult.Indices, true, common.ColorBgSelected)
	titleParts = append(titleParts, title)

	titleContent := strings.Join(titleParts,
		lipgloss.NewStyle().Background(common.ColorBgSelected).Render(" "))

	// Truncate if needed.
	if lipgloss.Width(titleContent) > width {
		titleContent = lipgloss.NewStyle().MaxWidth(width - 1).Render(titleContent)
		titleContent += lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSelected).Render("\u2026")
	}

	titleLine := lipgloss.NewStyle().
		Width(width).
		Background(common.ColorBgSelected).
		Bold(true).
		Render(titleContent)

	// -- Line 2: [indent] lang badge + tags --
	var metaParts []string

	if sn.Language != "" {
		metaParts = append(metaParts, common.RenderLangBadge(sn.Language, common.ColorBgSelected))
	}

	if len(sn.Tags) > 0 {
		used := 2 + lipgloss.Width(strings.Join(metaParts, " ")) // indent + lang badge
		for _, tag := range sn.Tags {
			rendered := common.RenderTagBadge(tag, common.ColorTextSub, common.ColorBgSelected)
			needed := lipgloss.Width(rendered) + 1
			if used+needed > width {
				break
			}
			metaParts = append(metaParts, rendered)
			used += needed
		}
	}

	metaContent := lipgloss.NewStyle().Background(common.ColorBgSelected).Render("  ") +
		strings.Join(metaParts, lipgloss.NewStyle().Background(common.ColorBgSelected).Render(" "))

	if lipgloss.Width(metaContent) > width {
		metaContent = lipgloss.NewStyle().MaxWidth(width - 1).Render(metaContent)
		metaContent += lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSelected).Render("\u2026")
	}

	metaLine := lipgloss.NewStyle().
		Width(width).
		Background(common.ColorBgSelected).
		Render(metaContent)

	return []string{titleLine, metaLine}
}

// renderPreviewBox renders a syntax-highlighted code preview inside a bordered box.
// Interior: bgSurface. Border outer edge: bgSelected.
func (r ResultList) renderPreviewBox(sn model.Snippet, width int) string {
	if sn.Content == "" {
		return lipgloss.NewStyle().Width(width).Background(common.ColorBgSelected).Render("")
	}

	lines := strings.Split(sn.Content, "\n")
	maxLines := 4
	if maxLines > len(lines) {
		maxLines = len(lines)
	}

	// Content width inside the box: width - border(2) - padding(2).
	codeWidth := width - 4
	if codeWidth < 5 {
		codeWidth = 5
	}

	var preview []string
	for _, line := range lines[:maxLines] {
		truncated := false
		if runeLen := len([]rune(line)); runeLen > codeWidth {
			line = string([]rune(line)[:codeWidth-1])
			truncated = true
		}
		highlighted := SyntaxHighlightLine(line, sn.Language)
		if truncated {
			highlighted += lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(common.ColorBgSurface).Render("\u2026")
		}
		preview = append(preview, highlighted)
	}

	previewText := strings.Join(preview, "\n")

	if len(lines) > maxLines {
		moreStyle := lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(common.ColorBgSurface)
		more := moreStyle.Render(fmt.Sprintf("+%d more lines", len(lines)-maxLines))
		previewText += "\n" + more
	}

	previewStyle := lipgloss.NewStyle().
		Width(width).
		Background(common.ColorBgSurface).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorBorderDim).
		BorderBackground(common.ColorBgSelected).
		Foreground(common.ColorTextSub).
		Padding(0, 1)

	return previewStyle.Render(previewText)
}

// renderSingleRow renders a single-line row (non-selected rows,
// or selected rows when showPreview is false).
func (r ResultList) renderSingleRow(item ResultItem, selected bool) string {
	sn := item.Snippet

	width := r.width
	indent := ""
	if r.showPreview {
		width = r.width - 1
		indent = " "
	}

	// Pick background for this row.
	rowBg := common.ColorBgSurface
	if selected {
		rowBg = common.ColorBgSelected
	}

	var parts []string

	if sn.Pinned {
		pinStyle := lipgloss.NewStyle().Foreground(common.ColorPink).Background(rowBg)
		parts = append(parts, pinStyle.Render("\u2726"))
	} else {
		parts = append(parts, lipgloss.NewStyle().Background(rowBg).Render(" "))
	}

	title := common.RenderFuzzyTitleWithBg(sn.Title, item.FuzzyResult.Indices, selected, rowBg)
	parts = append(parts, title)

	if sn.Language != "" {
		parts = append(parts, common.RenderLangBadge(sn.Language, rowBg))
	}

	if len(sn.Tags) > 0 {
		tagColor := common.ColorTextDim
		if selected {
			tagColor = common.ColorTextSub
		}

		prefix := strings.Join(parts, " ")
		used := lipgloss.Width(prefix)

		for _, tag := range sn.Tags {
			rendered := common.RenderTagBadge(tag, tagColor, rowBg)
			needed := lipgloss.Width(rendered) + 1
			if used+needed > width {
				break
			}
			parts = append(parts, rendered)
			used += needed
		}
	}

	content := strings.Join(parts, lipgloss.NewStyle().Background(rowBg).Render(" "))

	if lipgloss.Width(content) > width {
		content = lipgloss.NewStyle().MaxWidth(width-1).Render(content) + "\u2026"
	}

	rowStyle := lipgloss.NewStyle().Width(width).MaxWidth(width).Background(rowBg)
	if selected {
		rowStyle = rowStyle.Bold(true)
	}

	return indent + rowStyle.Render(content)
}

// ---------------------------------------------------------------------------
// Syntax highlighting for code preview
// ---------------------------------------------------------------------------

// numberRe matches standalone digit sequences for syntax highlighting.
var numberRe = regexp.MustCompile(`\b\d+\b`)

// Language keyword patterns for syntax highlighting.
var langKeywords = map[string]*regexp.Regexp{
	"go":         regexp.MustCompile(`\b(func|return|if|else|for|range|switch|case|type|struct|var|const|package|import|defer|go|select|chan|map|interface|nil|true|false|err)\b`),
	"nix":        regexp.MustCompile(`\b(let|in|with|rec|inherit|import|if|then|else|true|false|null)\b`),
	"sql":        regexp.MustCompile(`(?i)\b(CREATE|TABLE|VIRTUAL|USING|SELECT|FROM|JOIN|ON|WHERE|MATCH|ORDER|BY|INSERT|INTO|VALUES|TEXT|INTEGER|PRIMARY|KEY|DEFAULT|NOT|NULL|INDEX)\b`),
	"bash":       regexp.MustCompile(`\b(if|then|else|fi|for|do|done|echo|set|exit|function|return)\b`),
	"lua":        regexp.MustCompile(`\b(local|function|if|then|else|elseif|end|return|true|false|nil)\b`),
	"python":     regexp.MustCompile(`\b(def|class|return|if|elif|else|for|while|import|from|as|try|except|finally|with|yield|lambda|True|False|None|self)\b`),
	"typescript": regexp.MustCompile(`\b(function|return|if|else|for|while|const|let|var|import|export|from|async|await|class|new|this|throw|try|catch|finally|true|false|null|undefined|type|interface)\b`),
	"javascript": regexp.MustCompile(`\b(function|return|if|else|for|while|const|let|var|import|export|from|async|await|class|new|this|throw|try|catch|finally|true|false|null|undefined)\b`),
	"rust":       regexp.MustCompile(`\b(fn|let|mut|if|else|for|while|loop|match|return|pub|struct|enum|impl|trait|use|mod|self|Self|true|false|None|Some|Ok|Err)\b`),
}

// SyntaxHighlightLine applies basic syntax coloring to a single line of code.
// Keywords: mauve, strings: green, numbers: peach, comments: textMuted.
func SyntaxHighlightLine(line string, language string) string {
	if len(line) == 0 {
		return ""
	}

	// Detect line comments first.
	commentIdx := FindCommentStart(line, language)

	if commentIdx == 0 {
		// Entire line is a comment.
		return lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(common.ColorBgSurface).Render(line)
	}

	var codePart, commentPart string
	if commentIdx > 0 {
		codePart = line[:commentIdx]
		commentPart = line[commentIdx:]
	} else {
		codePart = line
	}

	// Tokenize and highlight the code part.
	highlighted := HighlightTokens(codePart, language)

	// Append comment in muted style.
	if commentPart != "" {
		highlighted += lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(common.ColorBgSurface).Render(commentPart)
	}

	return highlighted
}

// FindCommentStart returns the index of a line comment start, or -1 if none.
// Respects string literals (won't flag // inside "...").
func FindCommentStart(line string, language string) int {
	inString := false
	var stringChar rune
	for i, ch := range line {
		if inString {
			if ch == stringChar {
				inString = false
			}
			if ch == '\\' {
				continue
			}
			continue
		}
		if ch == '"' || ch == '\'' || ch == '`' {
			inString = true
			stringChar = ch
			continue
		}
		if ch == '/' && i+1 < len(line) && line[i+1] == '/' {
			return i
		}
		if ch == '#' && (language == "bash" || language == "nix" || language == "python") {
			return i
		}
		if ch == '-' && i+1 < len(line) && line[i+1] == '-' && language == "lua" {
			return i
		}
	}
	return -1
}

// HighlightTokens walks through a code string and applies syntax colors
// to keywords, strings, and numbers.
func HighlightTokens(code string, language string) string {
	keywordRe := langKeywords[language]

	kwStyle := lipgloss.NewStyle().Foreground(common.ColorMauve).Bold(true).Background(common.ColorBgSurface)
	strStyle := lipgloss.NewStyle().Foreground(common.ColorGreen).Background(common.ColorBgSurface)
	numStyle := lipgloss.NewStyle().Foreground(common.ColorPeach).Background(common.ColorBgSurface)
	defaultStyle := lipgloss.NewStyle().Foreground(common.ColorTextSub).Background(common.ColorBgSurface)

	var out strings.Builder
	runes := []rune(code)
	i := 0

	for i < len(runes) {
		ch := runes[i]

		// String literals.
		if ch == '"' || ch == '\'' || ch == '`' {
			quote := ch
			j := i + 1
			for j < len(runes) {
				if runes[j] == '\\' && j+1 < len(runes) {
					j += 2
					continue
				}
				if runes[j] == quote {
					j++
					break
				}
				j++
			}
			out.WriteString(strStyle.Render(string(runes[i:j])))
			i = j
			continue
		}

		// Word tokens (identifiers, keywords, numbers).
		if isWordChar(ch) {
			j := i
			for j < len(runes) && isWordChar(runes[j]) {
				j++
			}
			word := string(runes[i:j])

			if numberRe.MatchString(word) && !hasLetters(word) {
				out.WriteString(numStyle.Render(word))
			} else if keywordRe != nil && keywordRe.MatchString(word) {
				out.WriteString(kwStyle.Render(word))
			} else {
				out.WriteString(defaultStyle.Render(word))
			}
			i = j
			continue
		}

		// Everything else (operators, punctuation, whitespace).
		out.WriteString(defaultStyle.Render(string(ch)))
		i++
	}

	return out.String()
}

// isWordChar returns true for characters that form identifiers/numbers.
func isWordChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_'
}

// hasLetters returns true if the string contains at least one letter.
func hasLetters(s string) bool {
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			return true
		}
	}
	return false
}
