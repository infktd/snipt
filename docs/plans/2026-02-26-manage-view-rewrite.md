# Manage View() Rewrite Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite the manage TUI's rendering to match the design mockup -- sidebar on `ColorBg`, preview with metadata footer, `ColorMauve` status bar, horizontal separators.

**Architecture:** Pure rendering rewrite. All behavior (keybindings, navigation, CRUD, filtering) stays unchanged. Three files change: syntax highlighter gets a bg parameter, ResultList gets cursor/items accessors, manage.go gets a complete View() rewrite.

**Tech Stack:** Go, Bubbletea v2, Lip Gloss v2, Bubbles v2

**Design doc:** `docs/plans/2026-02-26-manage-view-rewrite-design.md`
**Full visual spec:** `docs/plans/MANAGE-PROMPT.md`

---

### Task 1: Add bg parameter to syntax highlighter

**Files:**
- Modify: `src/internal/tui/components/resultlist.go:564-594` (SyntaxHighlightLine)
- Modify: `src/internal/tui/components/resultlist.go:631-691` (HighlightTokens)

**Step 1: Update `SyntaxHighlightLine` signature and body**

Change the function signature to accept `bg color.Color` and pass it through:

```go
func SyntaxHighlightLine(line string, language string, bg color.Color) string {
	if len(line) == 0 {
		return ""
	}

	commentIdx := FindCommentStart(line, language)

	if commentIdx == 0 {
		return lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(bg).Render(line)
	}

	var codePart, commentPart string
	if commentIdx > 0 {
		codePart = line[:commentIdx]
		commentPart = line[commentIdx:]
	} else {
		codePart = line
	}

	highlighted := HighlightTokens(codePart, language, bg)

	if commentPart != "" {
		highlighted += lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(bg).Render(commentPart)
	}

	return highlighted
}
```

**Step 2: Update `HighlightTokens` signature and body**

Change the function signature to accept `bg color.Color` and use it in all styles:

```go
func HighlightTokens(code string, language string, bg color.Color) string {
	keywordRe := langKeywords[language]

	kwStyle := lipgloss.NewStyle().Foreground(common.ColorMauve).Bold(true).Background(bg)
	strStyle := lipgloss.NewStyle().Foreground(common.ColorGreen).Background(bg)
	numStyle := lipgloss.NewStyle().Foreground(common.ColorPeach).Background(bg)
	defaultStyle := lipgloss.NewStyle().Foreground(common.ColorTextSub).Background(bg)
	// ... rest of function body unchanged ...
```

**Step 3: Add `image/color` import**

The file already imports `image/color` indirectly via lipgloss types. Check if `color.Color` is available. The function parameter type should be `color.Color` from `image/color`. The file doesn't currently import `image/color` directly -- add it to the import block.

**Step 4: Update internal callers in `resultlist.go`**

Two call sites in `resultlist.go` need the extra argument:

Line 332 (`renderPreviewBox`):
```go
highlighted := SyntaxHighlightLine(line, sn.Language, common.ColorBgSurface)
```

Line 334 (the truncation ellipsis in `renderPreviewBox` -- already has explicit bg, no change needed).

**Step 5: Update caller in `manage.go`**

Line 875:
```go
highlighted := components.SyntaxHighlightLine(codeLine, sn.Language, common.ColorBg)
```

This will be replaced entirely in Task 3 when we rewrite `renderSnippetPreview`, but we need it to compile now.

**Step 6: Verify it compiles**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt/src && go build ./...`
Expected: clean build, no errors.

**Step 7: Commit**

```
git add src/internal/tui/components/resultlist.go src/internal/tui/manage/manage.go
git commit -m "refactor: add bg color parameter to syntax highlighter"
```

---

### Task 2: Add Cursor and Items accessors to ResultList

**Files:**
- Modify: `src/internal/tui/components/accessors.go`

**Step 1: Add `Cursor()` and `Items()` methods**

Append to `accessors.go`:

```go
// Cursor returns the current cursor position.
func (r *ResultList) Cursor() int {
	return r.cursor
}

// Items returns the current items slice.
func (r *ResultList) Items() []ResultItem {
	return r.items
}
```

Note: `Len()` and `Selected()` already exist in `resultlist.go`.

**Step 2: Verify it compiles**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt/src && go build ./...`
Expected: clean build.

**Step 3: Commit**

```
git add src/internal/tui/components/accessors.go
git commit -m "feat: add Cursor and Items accessors to ResultList"
```

---

### Task 3: Rewrite manage View() and all render methods

This is the big task. Replace everything from the `View()` method through all `render*` helpers in `manage.go` (lines 571-1018). Keep everything above line 571 (types, NewManageModel, Init, Update, editor helpers, applyFilter) and everything below line 1018 (RunManage) untouched.

**Files:**
- Modify: `src/internal/tui/manage/manage.go:571-1018`

**Step 1: Update the import block**

The rewrite needs `time` for `CreatedAt` formatting (already imported) and `image/color` is NOT needed here since we pass `common.ColorBg` directly. Verify imports include `fmt`, `strings`, `time`, and all existing imports. No new imports needed.

**Step 2: Replace View()**

Delete lines 571-594 and replace with:

```go
func (m ManageModel) View() tea.View {
	if m.quitting {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	if m.width == 0 || m.height == 0 {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	header := m.renderHeader()
	sepLine := m.renderHorizontalRule()
	content := m.renderContent()
	statusBar := m.renderStatusBar()

	screen := header + "\n" + sepLine + "\n" + content + "\n" + sepLine + "\n" + statusBar

	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}
```

Key change: two `sepLine` instances (above and below content), and `renderStatusSeparator` replaced with reusable `renderHorizontalRule`.

**Step 3: Replace renderHeader()**

Delete the old `renderHeader` (lines 600-644) and replace with:

```go
func (m ManageModel) renderHeader() string {
	badge := common.RenderBadgePill("SNIPT")
	gap := lipgloss.NewStyle().Width(2).Background(common.ColorBgSurface).Render("")

	searchView := m.searchInput.View()

	countStr := fmt.Sprintf("%d/%d", len(m.filtered), len(m.allSnippets))
	count := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(common.ColorBgSurface).
		Render(countStr)

	left := badge + gap + searchView
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(count)
	fillerWidth := m.width - leftWidth - rightWidth
	if fillerWidth < 1 {
		fillerWidth = 1
	}
	filler := lipgloss.NewStyle().Width(fillerWidth).Background(common.ColorBgSurface).Render("")

	row := left + filler + count

	return lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Background(common.ColorBgSurface).
		Render(row)
}
```

Changes from old: always uses `searchInput.View()` (the textinput handles focused/blurred state itself via the styles set in NewManageModel), explicit 2-space gap between badge and search.

**Step 4: Add renderHorizontalRule()**

New function replacing `renderStatusSeparator`:

```go
func (m ManageModel) renderHorizontalRule() string {
	rule := strings.Repeat("\u2500", m.width)
	return lipgloss.NewStyle().
		Foreground(common.ColorBorderDim).
		Background(common.ColorBg).
		Width(m.width).
		MaxWidth(m.width).
		Render(rule)
}
```

Change from old: background is `ColorBg` (was `ColorBgSurface`).

**Step 5: Replace renderContent()**

Delete old `renderContent` + `renderSidebar` (lines 650-757) and replace with:

```go
func (m ManageModel) renderContent() string {
	contentHeight := m.height - 4 // header + 2 separators + status bar
	if contentHeight < 1 {
		contentHeight = 1
	}

	sidebarWidth := m.width * 30 / 100
	if sidebarWidth < 25 {
		sidebarWidth = 25
	}
	if sidebarWidth > 40 {
		sidebarWidth = 40
	}
	previewWidth := m.width - sidebarWidth - 1 // 1 for vertical separator
	if previewWidth < 10 {
		previewWidth = 10
	}

	sidebarLines := m.renderSidebarLines(sidebarWidth, contentHeight)
	previewLines := m.renderPreviewLines(previewWidth, contentHeight)

	borderChar := lipgloss.NewStyle().
		Foreground(common.ColorBorderDim).
		Background(common.ColorBg).
		Render("\u2502")

	var rows []string
	for i := 0; i < contentHeight; i++ {
		sLine := sidebarLines[i]
		pLine := previewLines[i]

		row := sLine + borderChar + pLine

		// Safety: pad to full width if needed.
		if rowW := lipgloss.Width(row); rowW < m.width {
			row += lipgloss.NewStyle().Width(m.width - rowW).Background(common.ColorBg).Render("")
		}

		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}
```

Changes: added `max 40` cap on sidebarWidth, `contentHeight = m.height - 4` (correct math for 2 separators), returns pre-split lines instead of joining/re-splitting.

**Step 6: Add renderSidebarLines()**

New function. Replaces old `renderSidebar` which delegated to `resultList.View()`. Now renders directly using `resultList.Cursor()` and `resultList.Items()`:

```go
func (m ManageModel) renderSidebarLines(width, height int) []string {
	items := m.resultList.Items()
	cursor := m.resultList.Cursor()

	emptyLine := lipgloss.NewStyle().Width(width).Background(common.ColorBg).Render("")

	if len(items) == 0 {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = emptyLine
		}
		// Center "No snippets found" message.
		if height > 0 {
			msg := lipgloss.NewStyle().
				Foreground(common.ColorTextDim).
				Background(common.ColorBg).
				Render("No snippets found")
			msgW := lipgloss.Width(msg)
			pad := (width - msgW) / 2
			if pad < 0 {
				pad = 0
			}
			centered := lipgloss.NewStyle().Width(pad).Background(common.ColorBg).Render("") + msg
			lines[height/2] = lipgloss.NewStyle().Width(width).Background(common.ColorBg).Render(centered)
		}
		return lines
	}

	// Scroll window: 2 lines per item.
	visibleCount := height / 2
	if visibleCount > len(items) {
		visibleCount = len(items)
	}

	start := 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}
	end := start + visibleCount
	if end > len(items) {
		end = len(items)
		start = end - visibleCount
		if start < 0 {
			start = 0
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		item := items[i]
		selected := i == cursor
		line1, line2 := m.renderSidebarRow(item, selected, width)
		lines = append(lines, line1, line2)
	}

	// Pad remaining rows to fill contentHeight.
	for len(lines) < height {
		lines = append(lines, emptyLine)
	}

	// Truncate if we somehow have too many lines.
	if len(lines) > height {
		lines = lines[:height]
	}

	return lines
}
```

**Step 7: Add renderSidebarRow()**

New function. Returns exactly 2 lines for one snippet row:

```go
func (m ManageModel) renderSidebarRow(item components.ResultItem, selected bool, width int) (string, string) {
	sn := item.Snippet

	rowBg := common.ColorBg
	if selected {
		rowBg = common.ColorBgSelected
	}

	// Accent bar (2 chars): ▌+space for selected, 2 spaces for unselected.
	accentWidth := 2
	contentWidth := width - accentWidth
	if contentWidth < 4 {
		contentWidth = 4
	}

	var accent string
	if selected {
		accent = lipgloss.NewStyle().Foreground(common.ColorPink).Background(common.ColorBgSelected).Render("\u258c") +
			lipgloss.NewStyle().Width(1).Background(common.ColorBgSelected).Render(" ")
	} else {
		accent = lipgloss.NewStyle().Width(2).Background(common.ColorBg).Render("  ")
	}

	// -- Line 1: [pin ●] Title --
	var pinStr string
	if sn.Pinned {
		pinStr = lipgloss.NewStyle().Foreground(common.ColorPink).Background(rowBg).Render("\u25cf")
	} else {
		pinStr = lipgloss.NewStyle().Background(rowBg).Render(" ")
	}

	titleAvail := contentWidth - 2 // pin(1) + space(1)
	if titleAvail < 4 {
		titleAvail = 4
	}

	titleText := sn.Title
	titleTruncated := false
	if len([]rune(titleText)) > titleAvail {
		titleText = string([]rune(titleText)[:titleAvail-1])
		titleTruncated = true
	}

	indices := item.FuzzyResult.Indices
	if titleTruncated {
		var trimmed []int
		for _, idx := range indices {
			if idx < len([]rune(titleText)) {
				trimmed = append(trimmed, idx)
			}
		}
		indices = trimmed
	}

	title := common.RenderFuzzyTitleWithBg(titleText, indices, selected, rowBg)
	if titleTruncated {
		title += lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(rowBg).Render("\u2026")
	}

	spacer := lipgloss.NewStyle().Background(rowBg).Render(" ")
	line1Content := pinStr + spacer + title

	line1Style := lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Background(rowBg)
	if selected {
		line1Style = line1Style.Bold(true)
	}
	line1 := accent + line1Style.Render(line1Content)

	// -- Line 2: [2-space indent] lang badge + #tags --
	metaIndent := lipgloss.NewStyle().Background(rowBg).Render("  ")
	var metaParts []string

	if sn.Language != "" {
		metaParts = append(metaParts, common.RenderLangBadge(sn.Language, rowBg))
	}

	if len(sn.Tags) > 0 {
		tagColor := common.ColorTextDim
		if selected {
			tagColor = common.ColorTextSub
		}

		used := 2 + lipgloss.Width(strings.Join(metaParts, " ")) // indent + lang badge
		for _, tag := range sn.Tags {
			rendered := common.RenderTagBadge(tag, tagColor, rowBg)
			needed := lipgloss.Width(rendered) + 1
			if used+needed > contentWidth {
				break
			}
			metaParts = append(metaParts, rendered)
			used += needed
		}
	}

	line2Content := metaIndent +
		strings.Join(metaParts, lipgloss.NewStyle().Background(rowBg).Render(" "))

	line2Style := lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Background(rowBg)
	line2 := accent + line2Style.Render(line2Content)

	return line1, line2
}
```

**Step 8: Replace renderPreview with renderPreviewLines()**

Delete old `renderPreview`, `renderEmptyPreview`, `renderSnippetPreview` (lines 759-935) and replace with:

```go
func (m ManageModel) renderPreviewLines(width, height int) []string {
	bg := common.ColorBg
	emptyLine := lipgloss.NewStyle().Width(width).Background(bg).Render("")

	sel := m.resultList.Selected()
	if sel == nil {
		// Empty state: center message.
		lines := make([]string, height)
		for i := range lines {
			lines[i] = emptyLine
		}
		if height > 0 {
			msg := lipgloss.NewStyle().
				Foreground(common.ColorTextDim).
				Background(bg).
				Italic(true).
				Render("select a snippet to preview")
			msgW := lipgloss.Width(msg)
			pad := (width - msgW) / 2
			if pad < 0 {
				pad = 0
			}
			centered := lipgloss.NewStyle().Width(pad).Background(bg).Render("") + msg
			lines[height/2] = lipgloss.NewStyle().Width(width).Background(bg).Render(centered)
		}
		return lines
	}

	sn := sel.Snippet
	padding := 2 // left margin inside preview
	contentWidth := width - padding
	if contentWidth < 5 {
		contentWidth = 5
	}

	padLeft := lipgloss.NewStyle().Width(padding).Background(bg).Render("")

	var rawLines []string

	// -- Preview header: Title ... lang badge --
	titleStyle := lipgloss.NewStyle().Foreground(common.ColorText).Background(bg).Bold(true)
	titleStr := titleStyle.Render(sn.Title)

	langBadge := ""
	if sn.Language != "" {
		langBadge = common.RenderLangBadge(sn.Language, bg)
	}

	titleW := lipgloss.Width(titleStr)
	badgeW := lipgloss.Width(langBadge)
	titleGap := contentWidth - titleW - badgeW
	if titleGap < 1 {
		titleGap = 1
	}
	titleFiller := lipgloss.NewStyle().Width(titleGap).Background(bg).Render("")
	rawLines = append(rawLines, titleStr+titleFiller+langBadge)

	// -- Header separator --
	sepStr := strings.Repeat("\u2500", contentWidth)
	sep := lipgloss.NewStyle().Foreground(common.ColorBorderDim).Background(bg).Render(sepStr)
	rawLines = append(rawLines, sep)

	// -- Code with line numbers --
	codeLines := strings.Split(sn.Content, "\n")
	// Reserve: header(1) + headerSep(1) + footerSep(1) + footer(1) = 4 lines overhead
	maxCodeLines := height - 4
	if maxCodeLines < 1 {
		maxCodeLines = 1
	}
	if len(codeLines) < maxCodeLines {
		maxCodeLines = len(codeLines)
	}

	lineNumWidth := 3 // right-aligned in 3-char field
	lineNumStyle := lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(bg)
	codeSepStyle := lipgloss.NewStyle().Foreground(common.ColorBorderDim).Background(bg)

	for i := 0; i < maxCodeLines; i++ {
		num := fmt.Sprintf("%*d", lineNumWidth, i+1)
		numStr := lineNumStyle.Render(num)
		codeSep := codeSepStyle.Render(" \u2502 ")

		codeLine := codeLines[i]
		codeAvail := contentWidth - lineNumWidth - 3 // " │ " = 3 chars
		if len([]rune(codeLine)) > codeAvail {
			codeLine = string([]rune(codeLine)[:codeAvail-1]) + "\u2026"
		}
		highlighted := components.SyntaxHighlightLine(codeLine, sn.Language, bg)

		rawLines = append(rawLines, numStr+codeSep+highlighted)
	}

	// -- Footer separator --
	rawLines = append(rawLines, sep)

	// -- Footer: metadata left, tags right --
	lineCount := len(strings.Split(sn.Content, "\n"))
	metaStyle := lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(bg)

	leftMeta := fmt.Sprintf("%d lines", lineCount)
	if sn.UseCount > 0 {
		leftMeta += fmt.Sprintf("  used %d\u00d7", sn.UseCount)
	}
	if !sn.CreatedAt.IsZero() {
		leftMeta += fmt.Sprintf("  created %s", sn.CreatedAt.Format("2006-01-02"))
	}

	var rightTags string
	if len(sn.Tags) > 0 {
		tagParts := make([]string, len(sn.Tags))
		for i, tag := range sn.Tags {
			tagParts[i] = "#" + tag
		}
		rightTags = strings.Join(tagParts, " ")
	}

	leftStr := metaStyle.Render(leftMeta)
	rightStr := metaStyle.Render(rightTags)
	leftW := lipgloss.Width(leftStr)
	rightW := lipgloss.Width(rightStr)
	footerGap := contentWidth - leftW - rightW
	if footerGap < 1 {
		footerGap = 1
	}
	footerFiller := lipgloss.NewStyle().Width(footerGap).Background(bg).Render("")
	rawLines = append(rawLines, leftStr+footerFiller+rightStr)

	// -- Pad each raw line with left margin and fill to width --
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(rawLines) {
			line := padLeft + rawLines[i]
			lineW := lipgloss.Width(line)
			if lineW < width {
				line += lipgloss.NewStyle().Width(width - lineW).Background(bg).Render("")
			}
			lines[i] = line
		} else {
			lines[i] = emptyLine
		}
	}

	return lines
}
```

**Step 9: Replace renderStatusBar()**

Delete old `renderStatusSeparator` + `renderStatusBar` (lines 941-1018) and replace with:

```go
func (m ManageModel) renderStatusBar() string {
	barBg := common.ColorMauve
	textColor := common.ColorBg

	boldStyle := lipgloss.NewStyle().Foreground(textColor).Background(barBg).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor).Background(barBg)

	if m.mode == modeConfirmDelete {
		title := m.deleteTarget.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		content := boldStyle.Render(fmt.Sprintf(" delete \"%s\"?", title)) +
			normalStyle.Render("  ") +
			boldStyle.Render("y") + normalStyle.Render(" confirm") +
			normalStyle.Render("  ") +
			boldStyle.Render("any key") + normalStyle.Render(" cancel")

		return lipgloss.NewStyle().
			Width(m.width).
			MaxWidth(m.width).
			Background(barBg).
			Render(content)
	}

	// Left side: SNIPT label + mode.
	label := boldStyle.Render(" SNIPT")
	var modeStr string
	if m.mode == modeSearch {
		query := m.searchInput.Value()
		if query != "" {
			modeStr = normalStyle.Render(fmt.Sprintf("  search: \"%s\"", query))
		} else {
			modeStr = normalStyle.Render("  search")
		}
	} else {
		modeStr = normalStyle.Render("  browse")
	}

	left := label + modeStr

	// Right side: count + key hints.
	countStr := normalStyle.Render(fmt.Sprintf("%d/%d snippets", len(m.filtered), len(m.allSnippets)))

	var hints []struct{ key, desc string }
	if m.mode == modeSearch {
		hints = []struct{ key, desc string }{
			{"\u2191\u2193", "navigate"},
			{"enter", "confirm"},
			{"esc", "cancel"},
		}
	} else {
		hints = []struct{ key, desc string }{
			{"\u2191\u2193", "navigate"},
			{"/", "search"},
			{"enter", "copy"},
			{"n", "new"},
			{"e", "edit"},
			{"d", "delete"},
			{"p", "pin"},
			{"q", "quit"},
		}
	}

	var hintParts []string
	for _, h := range hints {
		hintParts = append(hintParts, boldStyle.Render(h.key)+normalStyle.Render(" "+h.desc))
	}
	hintsStr := strings.Join(hintParts, normalStyle.Render("  "))

	right := countStr + normalStyle.Render("  ") + hintsStr + normalStyle.Render(" ")

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	spacerW := m.width - leftW - rightW
	if spacerW < 1 {
		spacerW = 1
	}
	spacer := lipgloss.NewStyle().Width(spacerW).Background(barBg).Render("")

	row := left + spacer + right

	return lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Background(barBg).
		Render(row)
}
```

Changes from old:
- Background: `ColorMauve` (was `ColorBgSurface`)
- Text: `ColorBg` dark text on purple (was colored text on dark bg)
- Layout: `SNIPT browse ... count hints` (was just hints)
- Delete confirmation: same `ColorMauve` bg, not red
- Removed `copiedFeedback` display from status bar (copy feedback was inline in old bar; if desired it can be added back later, but the mockup doesn't show it)

**Step 10: Update contentHeight in Update/WindowSizeMsg**

In `Update()`, line 133, change:
```go
contentHeight := m.height - 3 // header + status separator + status bar
```
to:
```go
contentHeight := m.height - 4 // header + 2 separators + status bar
```

And update the ResultList height calculation on line 136:
```go
m.resultList.SetHeight(contentHeight)
```

(Remove the `-2` since we no longer add breathing lines -- the sidebar renderer handles its own height.)

**Step 11: Verify it compiles**

Run: `cd /Users/jayne/Desktop/codingProjects/snipt/src && go build ./...`
Expected: clean build, no errors.

**Step 12: Visually test**

Run the TUI and verify:
1. Header: SNIPT badge + search input + count on surface bg
2. Horizontal separator below header
3. Sidebar: items on dark bg, selected item has pink accent + lighter bg
4. Vertical separator between sidebar and preview
5. Preview: title + lang badge, separator, numbered code, separator, metadata footer
6. Horizontal separator above status bar
7. Status bar: purple bg with dark text, SNIPT label + mode + count + hints

**Step 13: Commit**

```
git add src/internal/tui/manage/manage.go
git commit -m "feat: rewrite manage View() to match design mockup

Sidebar on ColorBg with ColorBgSelected highlight, preview with
metadata footer, ColorMauve status bar, horizontal separators
between zones."
```
