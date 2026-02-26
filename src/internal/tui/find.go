package tui

import (
	"fmt"
	"sort"
	"strings"

	"image/color"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/model"
)

const (
	findDefaultWidth  = 80
	findDefaultHeight = 24
	findPadding       = 4 // vertical padding above/below the panel
)

// FindModel is the Bubbletea model for the snipt find palette.
type FindModel struct {
	searchInput textinput.Model
	resultList  ResultList
	allSnippets []model.Snippet
	filtered    []ResultItem
	selected    *model.Snippet // final selection (nil if cancelled)
	cancelled   bool
	copied      bool // whether "copied" feedback was shown
	width       int
	height      int
	idOnly      bool
	clipOutput  bool
}

// NewFindModel creates a new find palette model.
func NewFindModel(snippets []model.Snippet, initialQuery string, idOnly, clipOutput bool) FindModel {
	ti := textinput.New()
	ti.Placeholder = "Search snippets..."
	ti.CharLimit = 120
	ti.Width = findDefaultWidth - 20 // leave room for badge + count
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorMauve)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorText)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ColorTextMuted)
	ti.Prompt = " "
	ti.Focus()
	if initialQuery != "" {
		ti.SetValue(initialQuery)
	}

	m := FindModel{
		searchInput: ti,
		allSnippets: snippets,
		width:       findDefaultWidth,
		height:      findDefaultHeight,
		idOnly:      idOnly,
		clipOutput:  clipOutput,
	}

	// Initialize the result list with available height for results.
	listHeight := m.height - findPadding - 4 // header + border + hint rows
	if listHeight < 3 {
		listHeight = 3
	}
	m.resultList = NewResultList(findDefaultWidth-4, listHeight)

	// Run initial filter.
	m.applyFilter()

	return m
}

func (m FindModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FindModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		panelWidth := findDefaultWidth
		if panelWidth > m.width-4 {
			panelWidth = m.width - 4
		}

		listHeight := m.height - findPadding - 4
		if listHeight < 3 {
			listHeight = 3
		}

		m.resultList.width = panelWidth - 4
		m.resultList.height = listHeight
		m.searchInput.Width = panelWidth - 20

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if sel := m.resultList.Selected(); sel != nil {
				sn := sel.Snippet
				m.selected = &sn
			}
			return m, tea.Quit

		case "up", "ctrl+p", "down", "ctrl+n":
			var cmd tea.Cmd
			m.resultList, cmd = m.resultList.Update(msg)
			return m, cmd
		}
	}

	// Update text input and re-filter on change.
	prevValue := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	if m.searchInput.Value() != prevValue {
		m.applyFilter()
	}

	return m, cmd
}

// applyFilter fuzzy-matches the current query against all snippets and updates the result list.
func (m *FindModel) applyFilter() {
	query := m.searchInput.Value()

	if query == "" {
		// Show all snippets, no scoring needed (except pin bonus).
		items := make([]ResultItem, 0, len(m.allSnippets))
		for _, sn := range m.allSnippets {
			items = append(items, ResultItem{
				Snippet:     sn,
				FuzzyResult: FuzzyResult{Match: true, Score: 0},
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

	queryLower := strings.ToLower(query)

	type scored struct {
		item  ResultItem
		total int
	}

	var results []scored
	for _, sn := range m.allSnippets {
		fr := FuzzyMatch(sn.Title, query)
		if !fr.Match {
			// Even if the title does not fuzzy-match, include if tags or content match.
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
			// Create a minimal FuzzyResult for non-title matches.
			fr = FuzzyResult{Match: true, Score: 0}
		}

		total := fr.Score

		// Bonus: +3 if pinned.
		if sn.Pinned {
			total += 3
		}

		// Bonus: +5 if any tag matches the query.
		for _, tag := range sn.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				total += 5
				break
			}
		}

		// Bonus: +2 if content contains the query substring.
		if strings.Contains(strings.ToLower(sn.Content), queryLower) {
			total += 2
		}

		results = append(results, scored{
			item: ResultItem{
				Snippet:     sn,
				FuzzyResult: fr,
			},
			total: total,
		})
	}

	// Sort by total score descending.
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].total > results[j].total
	})

	items := make([]ResultItem, len(results))
	for i, r := range results {
		items[i] = r.item
	}

	m.filtered = items
	m.resultList.SetItems(items)
}

func (m FindModel) View() string {
	if m.selected != nil || m.cancelled {
		return ""
	}

	panelWidth := findDefaultWidth
	if panelWidth > m.width-4 {
		panelWidth = m.width - 4
	}
	if panelWidth < 40 {
		panelWidth = 40
	}

	innerWidth := panelWidth - 4 // account for border padding

	// -- Header line: SNIPT badge + search input + count --
	badge := renderBadgeGradient("SNIPT")

	countStr := fmt.Sprintf("%d/%d", len(m.filtered), len(m.allSnippets))
	countStyle := lipgloss.NewStyle().Foreground(ColorTextDim)
	count := countStyle.Render(countStr)

	// Calculate available width for search input.
	searchWidth := innerWidth - lipgloss.Width(badge) - lipgloss.Width(count) - 4
	if searchWidth < 20 {
		searchWidth = 20
	}
	m.searchInput.Width = searchWidth

	searchView := m.searchInput.View()

	// Pad the search to fill the middle space.
	headerParts := fmt.Sprintf("%s  %s", badge, searchView)
	headerGap := innerWidth - lipgloss.Width(headerParts) - lipgloss.Width(count)
	if headerGap < 1 {
		headerGap = 1
	}
	header := headerParts + strings.Repeat(" ", headerGap) + count

	// -- Separator --
	sepStyle := lipgloss.NewStyle().Foreground(ColorBorderDim)
	separator := sepStyle.Render(strings.Repeat("\u2500", innerWidth))

	// -- Result list --
	listView := m.resultList.View()

	// -- Compose the inner content --
	inner := lipgloss.JoinVertical(lipgloss.Left,
		header,
		separator,
		listView,
	)

	// -- Wrap in a rounded border --
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1).
		Width(panelWidth)

	panel := borderStyle.Render(inner)

	// -- Hint line --
	hintStyle := lipgloss.NewStyle().Foreground(ColorTextDim)
	hint := hintStyle.Render("  \u2191\u2193 navigate  enter copy  esc close")

	// -- Center the panel --
	full := lipgloss.JoinVertical(lipgloss.Left, panel, hint)

	// Center horizontally.
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(full)

	// Add some top padding to vertically center-ish.
	topPad := (m.height - lipgloss.Height(centered)) / 3
	if topPad < 0 {
		topPad = 0
	}

	return strings.Repeat("\n", topPad) + centered
}

// renderBadgeGradient renders text with a pink→mauve color gradient, one color per character.
func renderBadgeGradient(text string) string {
	// Pink (#f5c2e7) → Mauve (#cba6f7) gradient stops.
	colors := []lipgloss.Color{
		lipgloss.Color("#f5c2e7"),
		lipgloss.Color("#e4b4ef"),
		lipgloss.Color("#d8aaf3"),
		lipgloss.Color("#cba6f7"),
		lipgloss.Color("#c4a2f9"),
	}

	var out strings.Builder
	for i, ch := range text {
		ci := i
		if ci >= len(colors) {
			ci = len(colors) - 1
		}
		s := lipgloss.NewStyle().Foreground(colors[ci]).Bold(true)
		out.WriteString(s.Render(string(ch)))
	}
	return out.String()
}

// RunFind launches the find palette TUI and returns the selected snippet, or nil if cancelled.
func RunFind(snippets []model.Snippet, initialQuery string, idOnly, clipOutput bool) (*model.Snippet, error) {
	m := NewFindModel(snippets, initialQuery, idOnly, clipOutput)
	p := tea.NewProgram(m, tea.WithAltScreen())

	final, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("run find: %w", err)
	}

	fm := final.(FindModel)
	if fm.cancelled {
		return nil, nil
	}

	return fm.selected, nil
}
