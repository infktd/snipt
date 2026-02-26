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
		contentHeight := m.height - 2 // header + status bar

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
	content := m.renderContent()
	statusBar := m.renderStatusBar()

	screen := header + "\n" + content + "\n" + statusBar

	v := tea.NewView(screen)
	v.AltScreen = true
	return v
}

// ---------------------------------------------------------------------------
// Header
// ---------------------------------------------------------------------------

func (m ManageModel) renderHeader() string {
	badge := common.RenderBadgePill("SNIPT")

	// Search input or placeholder.
	var searchView string
	if m.mode == modeSearch {
		searchView = m.searchInput.View()
	} else if m.searchInput.Value() != "" {
		// Show query as blurred text.
		searchView = m.searchInput.View()
	} else {
		searchView = lipgloss.NewStyle().
			Foreground(common.ColorTextDim).
			Background(common.ColorBgSurface).
			Render("  Search snippets...")
	}

	countStr := fmt.Sprintf("%d/%d", len(m.filtered), len(m.allSnippets))
	count := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(common.ColorBgSurface).
		Render(countStr)

	left := badge + searchView
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(count)
	gap := m.width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}
	filler := lipgloss.NewStyle().
		Width(gap).
		Background(common.ColorBgSurface).
		Render("")

	row := left + filler + count

	// Ensure full width with background.
	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Background(common.ColorBgSurface)

	return headerStyle.Render(row)
}

// ---------------------------------------------------------------------------
// Content: sidebar + border + preview
// ---------------------------------------------------------------------------

func (m ManageModel) renderContent() string {
	contentHeight := m.height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	sidebarWidth := m.width * 30 / 100
	if sidebarWidth < 25 {
		sidebarWidth = 25
	}
	borderWidth := 1
	previewWidth := m.width - sidebarWidth - borderWidth
	if previewWidth < 10 {
		previewWidth = 10
	}

	sidebarStr := m.renderSidebar(sidebarWidth, contentHeight)
	previewStr := m.renderPreview(previewWidth, contentHeight)

	// Split both into lines and join line by line.
	sidebarLines := strings.Split(sidebarStr, "\n")
	previewLines := strings.Split(previewStr, "\n")

	borderStyle := lipgloss.NewStyle().
		Foreground(common.ColorBorder).
		Background(common.ColorBg)

	var lines []string
	for i := 0; i < contentHeight; i++ {
		sLine := ""
		if i < len(sidebarLines) {
			sLine = sidebarLines[i]
		}
		// Ensure sidebar line fills width.
		sLineWidth := lipgloss.Width(sLine)
		if sLineWidth < sidebarWidth {
			pad := lipgloss.NewStyle().
				Width(sidebarWidth - sLineWidth).
				Background(common.ColorBgSurface).
				Render("")
			sLine += pad
		}

		border := borderStyle.Render("\u2502")

		pLine := ""
		if i < len(previewLines) {
			pLine = previewLines[i]
		}
		// Ensure preview line fills width.
		pLineWidth := lipgloss.Width(pLine)
		if pLineWidth < previewWidth {
			pad := lipgloss.NewStyle().
				Width(previewWidth - pLineWidth).
				Background(common.ColorBg).
				Render("")
			pLine += pad
		}

		lines = append(lines, sLine+border+pLine)
	}

	return strings.Join(lines, "\n")
}

func (m ManageModel) renderSidebar(width, height int) string {
	// Wrap the result list in a surface-colored container.
	listStr := m.resultList.View()
	listLines := strings.Split(listStr, "\n")

	// Pad each line to full sidebar width with left padding.
	padStyle := lipgloss.NewStyle().Background(common.ColorBgSurface)
	var out []string
	for i := 0; i < height; i++ {
		if i < len(listLines) {
			line := " " + listLines[i] // 1-char left margin
			lineWidth := lipgloss.Width(line)
			if lineWidth < width {
				line += padStyle.Width(width - lineWidth).Render("")
			}
			out = append(out, line)
		} else {
			out = append(out, padStyle.Width(width).Render(""))
		}
	}
	return strings.Join(out, "\n")
}

func (m ManageModel) renderPreview(width, height int) string {
	sel := m.resultList.Selected()
	if sel == nil {
		return m.renderEmptyPreview(width, height)
	}
	return m.renderSnippetPreview(sel.Snippet, width, height)
}

func (m ManageModel) renderEmptyPreview(width, height int) string {
	emptyMsg := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(common.ColorBg).
		Render("No snippets yet. Press n to create one.")

	var lines []string
	msgLine := height / 2
	for i := 0; i < height; i++ {
		if i == msgLine {
			msgWidth := lipgloss.Width(emptyMsg)
			leftPad := (width - msgWidth) / 2
			if leftPad < 0 {
				leftPad = 0
			}
			padStr := lipgloss.NewStyle().
				Width(leftPad).
				Background(common.ColorBg).
				Render("")
			line := padStr + emptyMsg
			lineWidth := lipgloss.Width(line)
			if lineWidth < width {
				line += lipgloss.NewStyle().Width(width - lineWidth).Background(common.ColorBg).Render("")
			}
			lines = append(lines, line)
		} else {
			lines = append(lines, lipgloss.NewStyle().Width(width).Background(common.ColorBg).Render(""))
		}
	}
	return strings.Join(lines, "\n")
}

func (m ManageModel) renderSnippetPreview(sn model.Snippet, width, height int) string {
	bg := common.ColorBg
	padding := 2 // left padding inside preview
	contentWidth := width - padding
	if contentWidth < 5 {
		contentWidth = 5
	}

	var lines []string

	// --- Title line with language badge ---
	titleStyle := lipgloss.NewStyle().
		Foreground(common.ColorText).
		Background(bg).
		Bold(true)
	titleStr := titleStyle.Render(sn.Title)

	langBadge := ""
	if sn.Language != "" {
		langBadge = common.RenderLangBadge(sn.Language, bg)
	}

	titleWidth := lipgloss.Width(titleStr)
	badgeWidth := lipgloss.Width(langBadge)
	titleGap := contentWidth - titleWidth - badgeWidth
	if titleGap < 1 {
		titleGap = 1
	}
	titleFiller := lipgloss.NewStyle().Width(titleGap).Background(bg).Render("")
	titleLine := titleStr + titleFiller + langBadge
	lines = append(lines, titleLine)

	// --- Blank line after title ---
	lines = append(lines, "")

	// --- Code with line numbers ---
	codeLines := strings.Split(sn.Content, "\n")
	maxCodeLines := height - 6 // reserve space for title, blank, blank, tags, meta
	if maxCodeLines < 1 {
		maxCodeLines = 1
	}
	if len(codeLines) < maxCodeLines {
		maxCodeLines = len(codeLines)
	}

	lineNumWidth := len(fmt.Sprintf("%d", maxCodeLines))
	if lineNumWidth < 2 {
		lineNumWidth = 2
	}

	lineNumStyle := lipgloss.NewStyle().
		Foreground(common.ColorTextMuted).
		Background(bg)
	sepStyle := lipgloss.NewStyle().
		Foreground(common.ColorBorderDim).
		Background(bg)

	for i := 0; i < maxCodeLines; i++ {
		num := fmt.Sprintf("%*d", lineNumWidth, i+1)
		numStr := lineNumStyle.Render(num)
		sep := sepStyle.Render(" \u2502 ")

		codeLine := codeLines[i]
		// Truncate long lines.
		codeAvail := contentWidth - lineNumWidth - 4 // " | " = 3 chars + num width
		if len([]rune(codeLine)) > codeAvail {
			codeLine = string([]rune(codeLine)[:codeAvail-1]) + "\u2026"
		}
		highlighted := components.SyntaxHighlightLine(codeLine, sn.Language)

		lines = append(lines, numStr+sep+highlighted)
	}

	// --- Blank line before metadata ---
	lines = append(lines, "")

	// --- Tags ---
	if len(sn.Tags) > 0 {
		tagLabel := lipgloss.NewStyle().
			Foreground(common.ColorTextDim).
			Background(bg).
			Render("Tags: ")
		var tagParts []string
		for _, tag := range sn.Tags {
			tagParts = append(tagParts, common.RenderTagBadge(tag, common.ColorTextSub, bg))
		}
		tagsLine := tagLabel + strings.Join(tagParts, " ")
		lines = append(lines, tagsLine)
	}

	// --- Pinned / Uses ---
	pinnedStr := "no"
	if sn.Pinned {
		pinnedStr = lipgloss.NewStyle().
			Foreground(common.ColorPink).
			Background(bg).
			Render("yes")
	} else {
		pinnedStr = lipgloss.NewStyle().
			Foreground(common.ColorTextDim).
			Background(bg).
			Render("no")
	}
	metaLabel := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(bg)
	metaLine := metaLabel.Render("Pinned: ") + pinnedStr +
		metaLabel.Render("  Uses: ") +
		lipgloss.NewStyle().Foreground(common.ColorTextSub).Background(bg).Render(fmt.Sprintf("%d", sn.UseCount))
	lines = append(lines, metaLine)

	// --- Pad lines to height and add left padding ---
	padLeft := lipgloss.NewStyle().Width(padding).Background(bg).Render("")
	var finalLines []string
	for i := 0; i < height; i++ {
		if i < len(lines) {
			line := padLeft + lines[i]
			lineWidth := lipgloss.Width(line)
			if lineWidth < width {
				line += lipgloss.NewStyle().Width(width - lineWidth).Background(bg).Render("")
			}
			finalLines = append(finalLines, line)
		} else {
			finalLines = append(finalLines, lipgloss.NewStyle().Width(width).Background(bg).Render(""))
		}
	}
	return strings.Join(finalLines, "\n")
}

// ---------------------------------------------------------------------------
// Status bar
// ---------------------------------------------------------------------------

func (m ManageModel) renderStatusBar() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(common.ColorText).
		Background(common.ColorBgOverlay).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(common.ColorBgOverlay)

	var content string

	if m.mode == modeConfirmDelete {
		title := m.deleteTarget.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		confirmMsg := lipgloss.NewStyle().
			Foreground(common.ColorRed).
			Background(common.ColorBgOverlay).
			Bold(true).
			Render(fmt.Sprintf("delete \"%s\"?", title))
		hintMsg := descStyle.Render(" y to confirm, any other key to cancel")
		content = "  " + confirmMsg + hintMsg
	} else {
		var hints []struct{ key, desc string }
		if m.mode == modeSearch {
			hints = []struct{ key, desc string }{
				{"esc", "clear"},
				{"\u2191\u2193", "navigate"},
				{"enter", "confirm"},
			}
		} else {
			hints = []struct{ key, desc string }{
				{"\u2191\u2193/jk", "navigate"},
				{"enter", "copy"},
				{"/", "search"},
				{"n", "new"},
				{"e", "edit"},
				{"d", "delete"},
				{"p", "pin"},
				{"q", "quit"},
			}
		}

		var parts []string
		for _, h := range hints {
			parts = append(parts, keyStyle.Render(h.key)+" "+descStyle.Render(h.desc))
		}

		content = "  " + strings.Join(parts, descStyle.Render("  "))

		if m.copiedFeedback {
			feedback := lipgloss.NewStyle().
				Foreground(common.ColorGreen).
				Background(common.ColorBgOverlay).
				Bold(true).
				Render("Copied!")
			content = "  " + feedback + descStyle.Render("  ") + content
		}
	}

	barStyle := lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Background(common.ColorBgOverlay)

	return barStyle.Render(content)
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
