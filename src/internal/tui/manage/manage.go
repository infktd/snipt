package manage

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/common"
	"github.com/infktd/snipt/src/internal/tui/components"
)

// mode tracks the current input mode.
type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeConfirmDelete
)

// ManageModel is the Bubbletea model for the full-screen manage TUI.
type ManageModel struct {
	store  *db.Store
	editor string

	allSnippets []model.Snippet
	filtered    []components.ResultItem
	resultList  components.ResultList
	searchInput textinput.Model

	width          int
	height         int
	mode           mode
	quitting       bool
	copiedFeedback bool

	editingID   string // ID of snippet being edited (empty for new)
	editTmpPath string // path to temp file being edited

	deleteTarget model.Snippet // snippet targeted for deletion
}

// NewManageModel creates a new manage screen model.
func NewManageModel(store *db.Store, editor string) (ManageModel, error) {
	snippets, err := store.List(db.ListOpts{})
	if err != nil {
		return ManageModel{}, fmt.Errorf("load snippets: %w", err)
	}

	items := make([]components.ResultItem, len(snippets))
	for i, sn := range snippets {
		items[i] = components.ResultItem{
			Snippet:     sn,
			FuzzyResult: components.FuzzyResult{Match: true},
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Snippet.Pinned != items[j].Snippet.Pinned {
			return items[i].Snippet.Pinned
		}
		return items[i].Snippet.Title < items[j].Snippet.Title
	})

	rl := components.NewResultList(30, 20)
	rl.SetShowPreview(false)
	rl.SetItems(items)

	// Initialize search textinput (starts blurred).
	ti := textinput.New()
	ti.Placeholder = "Search snippets..."
	ti.CharLimit = 120

	styles := ti.Styles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(common.ColorMauve).Background(common.ColorBgSurface)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(common.ColorText).Background(common.ColorBgSurface)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSurface)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSurface)
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSurface)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSurface)
	styles.Cursor.Color = common.ColorMauve
	styles.Cursor.Shape = tea.CursorBar
	ti.SetStyles(styles)
	ti.Prompt = " "
	ti.Blur()

	return ManageModel{
		store:       store,
		editor:      editor,
		allSnippets: snippets,
		filtered:    items,
		resultList:  rl,
		searchInput: ti,
		mode:        modeNormal,
	}, nil
}

func (m ManageModel) Init() tea.Cmd {
	return nil
}

// copiedMsg signals that the "Copied!" feedback timer has elapsed.
type copiedMsg struct{}

// editorFinishedMsg signals that the external editor has exited.
type editorFinishedMsg struct {
	err error
}

func (m ManageModel) clearCopiedAfter() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return copiedMsg{}
	})
}

func (m ManageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		sidebarWidth := m.width * 30 / 100
		if sidebarWidth < 25 {
			sidebarWidth = 25
		}
		if sidebarWidth > 40 {
			sidebarWidth = 40
		}
		contentHeight := m.height - 4 // header + 2 separators + status bar

		m.resultList.SetWidth(sidebarWidth - 2) // padding
		m.resultList.SetHeight(contentHeight)
		return m, nil

	case copiedMsg:
		m.copiedFeedback = false
		return m, nil

	case tea.KeyPressMsg:
		if m.mode == modeSearch {
			switch msg.String() {
			case "esc":
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.applyFilter()
				m.mode = modeNormal
				return m, nil
			case "enter":
				// Confirm search, switch back to normal mode keeping the filter.
				m.searchInput.Blur()
				m.mode = modeNormal
				return m, nil
			case "up", "ctrl+p":
				m.resultList, _ = m.resultList.Update(msg)
				return m, nil
			case "down", "ctrl+n":
				m.resultList, _ = m.resultList.Update(msg)
				return m, nil
			}

			// All other keys go to the textinput.
			prevValue := m.searchInput.Value()
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			if m.searchInput.Value() != prevValue {
				m.applyFilter()
			}
			return m, cmd
		}

		if m.mode == modeConfirmDelete {
			switch msg.String() {
			case "y":
				_ = m.store.Delete(m.deleteTarget.ID)
				m.deleteTarget = model.Snippet{}
				m.mode = modeNormal
				m.reloadSnippets()
				return m, nil
			default:
				// Any other key cancels.
				m.deleteTarget = model.Snippet{}
				m.mode = modeNormal
				return m, nil
			}
		}

		if m.mode == modeNormal {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "/":
				m.mode = modeSearch
				m.searchInput.Focus()
				return m, nil
			case "enter":
				if sel := m.resultList.Selected(); sel != nil {
					m.copiedFeedback = true
					_ = m.store.IncrementUseCount(sel.Snippet.ID)
					return m, tea.Batch(
						tea.SetClipboard(sel.Snippet.Content),
						m.clearCopiedAfter(),
					)
				}
				return m, nil
			case "j":
				m.resultList, _ = m.resultList.Update(tea.KeyPressMsg{Code: tea.KeyDown})
				return m, nil
			case "k":
				m.resultList, _ = m.resultList.Update(tea.KeyPressMsg{Code: tea.KeyUp})
				return m, nil
			case "down", "ctrl+n", "up", "ctrl+p":
				m.resultList, _ = m.resultList.Update(msg)
				return m, nil
			case "e":
				if sel := m.resultList.Selected(); sel != nil {
					sn := sel.Snippet
					m.editingID = sn.ID
					tmpPath, err := writeFrontmatterFile(sn)
					if err != nil {
						return m, nil
					}
					m.editTmpPath = tmpPath
					cmd := editorCommand(m.editor, tmpPath)
					if cmd == nil {
						os.Remove(tmpPath)
						return m, nil
					}
					return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
						return editorFinishedMsg{err: err}
					})
				}
				return m, nil
			case "n":
				m.editingID = "" // empty means create new
				tmpPath, err := writeFrontmatterFile(model.Snippet{
					Language: "text",
				})
				if err != nil {
					return m, nil
				}
				m.editTmpPath = tmpPath
				cmd := editorCommand(m.editor, tmpPath)
				if cmd == nil {
					os.Remove(tmpPath)
					return m, nil
				}
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return editorFinishedMsg{err: err}
				})
			case "d":
				if sel := m.resultList.Selected(); sel != nil {
					m.mode = modeConfirmDelete
					m.deleteTarget = sel.Snippet
				}
				return m, nil
			case "p":
				if sel := m.resultList.Selected(); sel != nil {
					newPinned := !sel.Snippet.Pinned
					_ = m.store.SetPinned(sel.Snippet.ID, newPinned)
					m.reloadSnippets()
				}
				return m, nil
			}
		}

	case editorFinishedMsg:
		defer os.Remove(m.editTmpPath)

		if msg.err != nil {
			m.editTmpPath = ""
			return m, nil
		}

		data, err := os.ReadFile(m.editTmpPath)
		if err != nil {
			m.editTmpPath = ""
			return m, nil
		}
		m.editTmpPath = ""

		content := string(data)
		if strings.TrimSpace(content) == "" {
			return m, nil
		}

		title, language, tags, pinned, body := parseFrontmatter(content)

		if m.editingID == "" {
			// Creating new snippet.
			sn := &model.Snippet{
				ID:       model.NewID(),
				Title:    title,
				Language: language,
				Tags:     tags,
				Pinned:   pinned,
				Content:  body,
			}
			m.store.Create(sn)
		} else {
			// Updating existing snippet.
			sn, err := m.store.Get(m.editingID)
			if err != nil {
				return m, nil
			}
			sn.Title = title
			sn.Language = language
			sn.Content = body
			sn.Pinned = pinned
			m.store.Update(sn)
			// Update tags: remove old, add new.
			m.store.RemoveTags(sn.ID, sn.Tags)
			m.store.AddTags(sn.ID, tags)
		}

		m.editingID = ""
		m.reloadSnippets()
		return m, nil
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Editor helpers
// ---------------------------------------------------------------------------

// writeFrontmatterFile writes a snippet to a temporary file with YAML frontmatter.
func writeFrontmatterFile(sn model.Snippet) (string, error) {
	f, err := os.CreateTemp("", "snipt-*.md")
	if err != nil {
		return "", err
	}

	fmt.Fprintf(f, "---\n")
	fmt.Fprintf(f, "title: %s\n", sn.Title)
	fmt.Fprintf(f, "language: %s\n", sn.Language)

	if len(sn.Tags) > 0 {
		fmt.Fprintf(f, "tags: [%s]\n", strings.Join(sn.Tags, ", "))
	} else {
		fmt.Fprintf(f, "tags: []\n")
	}

	fmt.Fprintf(f, "pinned: %v\n", sn.Pinned)
	fmt.Fprintf(f, "---\n\n")
	fmt.Fprintf(f, "%s", sn.Content)

	f.Close()
	return f.Name(), nil
}

// parseFrontmatter extracts title, language, tags, pinned flag, and body
// from a file that uses YAML frontmatter delimited by --- markers.
func parseFrontmatter(content string) (title, language string, tags []string, pinned bool, body string) {
	lines := strings.Split(content, "\n")

	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", nil, false, content
	}

	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}

	if closingIdx == -1 {
		return "", "", nil, false, content
	}

	for _, line := range lines[1:closingIdx] {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
		} else if strings.HasPrefix(line, "language:") {
			language = strings.TrimSpace(strings.TrimPrefix(line, "language:"))
		} else if strings.HasPrefix(line, "tags:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
			tagStr = strings.TrimPrefix(tagStr, "[")
			tagStr = strings.TrimSuffix(tagStr, "]")
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
		} else if strings.HasPrefix(line, "pinned:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "pinned:"))
			pinned = val == "true"
		}
	}

	bodyLines := lines[closingIdx+1:]
	if len(bodyLines) > 0 && strings.TrimSpace(bodyLines[0]) == "" {
		bodyLines = bodyLines[1:]
	}
	body = strings.Join(bodyLines, "\n")

	return
}

// reloadSnippets refreshes the snippet list from the store and re-applies the current filter.
func (m *ManageModel) reloadSnippets() {
	snippets, err := m.store.List(db.ListOpts{})
	if err != nil {
		return
	}
	m.allSnippets = snippets
	m.applyFilter()
}

// editorCommand builds an exec.Cmd for the configured editor. Returns nil if
// the editor string is empty (prevents index-out-of-range panic).
func editorCommand(editor, filePath string) *exec.Cmd {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return nil
	}
	return exec.Command(parts[0], append(parts[1:], filePath)...)
}

// applyFilter fuzzy-matches the current query against all snippets and updates the result list.
func (m *ManageModel) applyFilter() {
	query := m.searchInput.Value()

	if query == "" {
		// Show all snippets, no scoring needed (except pin bonus).
		items := make([]components.ResultItem, 0, len(m.allSnippets))
		for _, sn := range m.allSnippets {
			items = append(items, components.ResultItem{
				Snippet:     sn,
				FuzzyResult: components.FuzzyResult{Match: true, Score: 0},
			})
		}
		// Sort: pinned first, then by title.
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Snippet.Pinned != items[j].Snippet.Pinned {
				return items[i].Snippet.Pinned
			}
			return items[i].Snippet.Title < items[j].Snippet.Title
		})
		m.filtered = items
		m.resultList.SetItems(items)
		return
	}

	// Tag-only search: #foo matches only against tags, skips title/content.
	if strings.HasPrefix(query, "#") {
		tagQuery := strings.ToLower(strings.TrimPrefix(query, "#"))
		if tagQuery == "" {
			// Just "#" typed, show everything in same order as empty query.
			items := make([]components.ResultItem, 0, len(m.allSnippets))
			for _, sn := range m.allSnippets {
				items = append(items, components.ResultItem{
					Snippet:     sn,
					FuzzyResult: components.FuzzyResult{Match: true, Score: 0},
				})
			}
			sort.SliceStable(items, func(i, j int) bool {
				if items[i].Snippet.Pinned != items[j].Snippet.Pinned {
					return items[i].Snippet.Pinned
				}
				return items[i].Snippet.Title < items[j].Snippet.Title
			})
			m.filtered = items
			m.resultList.SetItems(items)
			return
		}

		var items []components.ResultItem
		for _, sn := range m.allSnippets {
			for _, tag := range sn.Tags {
				if strings.Contains(strings.ToLower(tag), tagQuery) {
					score := 0
					if sn.Pinned {
						score += 3
					}
					// Exact match gets a big boost.
					if strings.ToLower(tag) == tagQuery {
						score += 10
					}
					items = append(items, components.ResultItem{
						Snippet:     sn,
						FuzzyResult: components.FuzzyResult{Match: true, Score: score},
					})
					break
				}
			}
		}
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].FuzzyResult.Score > items[j].FuzzyResult.Score
		})
		m.filtered = items
		m.resultList.SetItems(items)
		return
	}

	// Fuzzy search on title + tag/content fallback.
	queryLower := strings.ToLower(query)

	type scored struct {
		item  components.ResultItem
		total int
	}

	var results []scored
	for _, sn := range m.allSnippets {
		fr := components.FuzzyMatch(sn.Title, query)
		if !fr.Match {
			hasTagMatch := false
			for _, tag := range sn.Tags {
				if strings.Contains(strings.ToLower(tag), queryLower) {
					hasTagMatch = true
					break
				}
			}
			hasContentMatch := strings.Contains(strings.ToLower(sn.Content), queryLower)

			if !hasTagMatch && !hasContentMatch {
				continue
			}
			fr = components.FuzzyResult{Match: true, Score: 0}
		}

		total := fr.Score

		if sn.Pinned {
			total += 3
		}

		for _, tag := range sn.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				total += 5
				break
			}
		}

		if strings.Contains(strings.ToLower(sn.Content), queryLower) {
			total += 2
		}

		results = append(results, scored{
			item: components.ResultItem{
				Snippet:     sn,
				FuzzyResult: fr,
			},
			total: total,
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].total > results[j].total
	})

	items := make([]components.ResultItem, len(results))
	for i, r := range results {
		items[i] = r.item
	}

	m.filtered = items
	m.resultList.SetItems(items)
}


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

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

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

func (m ManageModel) renderHorizontalRule() string {
	rule := strings.Repeat("\u2500", m.width)
	return lipgloss.NewStyle().
		Foreground(common.ColorBorderDim).
		Background(common.ColorBg).
		Width(m.width).
		MaxWidth(m.width).
		Render(rule)
}

// ---------------------------------------------------------------------------
// Content: sidebar + border + preview
// ---------------------------------------------------------------------------

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

		// Safety: pad or trim to exact terminal width.
		if rowW := lipgloss.Width(row); rowW < m.width {
			row += lipgloss.NewStyle().Width(m.width - rowW).Background(common.ColorBg).Render("")
		} else if rowW > m.width {
			row = lipgloss.NewStyle().MaxWidth(m.width).Render(row)
		}

		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

func (m ManageModel) renderSidebarLines(width, height int) []string {
	items := m.resultList.Items()
	cursor := m.resultList.Cursor()

	emptyLine := lipgloss.NewStyle().Width(width).Background(common.ColorBg).Render("")

	if len(items) == 0 {
		lines := make([]string, height)
		for i := range lines {
			lines[i] = emptyLine
		}
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

// ---------------------------------------------------------------------------
// Preview pane
// ---------------------------------------------------------------------------

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
		if codeAvail < 1 {
			codeAvail = 1
		}
		if len([]rune(codeLine)) > codeAvail {
			codeLine = string([]rune(codeLine)[:codeAvail-1]) + "\u2026"
		}
		highlighted := components.SyntaxHighlightLine(codeLine, sn.Language, bg)

		rawLines = append(rawLines, numStr+codeSep+highlighted)
	}

	// -- Footer separator --
	rawLines = append(rawLines, sep)

	// -- Footer: metadata left, tags right --
	lineCount := len(codeLines)
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

// ---------------------------------------------------------------------------
// Status bar
// ---------------------------------------------------------------------------

func (m ManageModel) renderStatusBar() string {
	barBg := common.ColorMauve
	textColor := common.ColorBg

	boldStyle := lipgloss.NewStyle().Foreground(textColor).Background(barBg).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(textColor).Background(barBg)

	if m.mode == modeConfirmDelete {
		title := m.deleteTarget.Title
		if titleRunes := []rune(title); len(titleRunes) > 30 {
			title = string(titleRunes[:27]) + "..."
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

	// Left side: SNIPT label + mode/feedback.
	label := boldStyle.Render(" SNIPT")
	var modeStr string
	if m.copiedFeedback {
		modeStr = normalStyle.Render("  ") + boldStyle.Render("Copied!")
	} else if m.mode == modeSearch {
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

// RunManage launches the manage TUI.
func RunManage(store *db.Store, editor string) error {
	m, err := NewManageModel(store, editor)
	if err != nil {
		return err
	}
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
