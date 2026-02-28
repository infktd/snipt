package find

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/common"
	"github.com/infktd/snipt/src/internal/tui/components"
)

const (
	findDefaultWidth  = 96
	findDefaultHeight = 24
	findPadding       = 4 // rows reserved for panel chrome outside the result list
	findListMaxHeight = 12
)

// FindModel is the Bubbletea model for the snipt find palette.
type FindModel struct {
	searchInput textinput.Model
	resultList  components.ResultList
	allSnippets []model.Snippet
	filtered    []components.ResultItem
	selected  *model.Snippet // final selection (nil if cancelled)
	cancelled bool
	width     int
	height    int
}

// NewFindModel creates a new find palette model.
func NewFindModel(snippets []model.Snippet, initialQuery string) FindModel {
	ti := textinput.New()
	ti.Placeholder = "Search snippets..."
	ti.CharLimit = 120
	ti.SetWidth(findDefaultWidth - 20) // leave room for badge + count

	// Configure styles: text colors + bgSurface backgrounds.
	styles := ti.Styles()

	// StyleState fields (Focused and Blurred) -- no Cursor here.
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(common.ColorMauve).Background(common.ColorBgSurface)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(common.ColorText).Background(common.ColorBgSurface)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(common.ColorBgSurface)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSurface)
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(common.ColorTextDim).Background(common.ColorBgSurface)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(common.ColorTextMuted).Background(common.ColorBgSurface)

	// Cursor styling lives on Styles.Cursor (top-level CursorStyle), NOT
	// inside StyleState. This matches the textarea v2 API.
	// CursorStyle has: Blink bool, Color color.Color, Shape tea.CursorShape
	styles.Cursor.Blink = false              // static, no blinking
	styles.Cursor.Color = common.ColorMauve  // mauve cursor color
	styles.Cursor.Shape = tea.CursorBar      // thin bar instead of block

	ti.SetStyles(styles)
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
	}

	// Initialize the result list with available height for results.
	listHeight := findListMaxHeight
	if listHeight > m.height-findPadding-4 {
		listHeight = m.height - findPadding - 4
	}
	if listHeight < 3 {
		listHeight = 3
	}
	m.resultList = components.NewResultList(findDefaultWidth-6, listHeight)
	m.resultList.SetShowPreview(true)

	// Run initial filter.
	m.applyFilter()

	return m
}

func (m FindModel) Init() tea.Cmd {
	return nil
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
		if panelWidth < 40 {
			panelWidth = 40
		}
		if panelWidth > m.width {
			panelWidth = m.width
		}

		listHeight := findListMaxHeight
		if listHeight > m.height-findPadding-4 {
			listHeight = m.height - findPadding - 4
		}
		if listHeight < 3 {
			listHeight = 3
		}

		m.resultList.SetWidth(panelWidth - 6)
		m.resultList.SetHeight(listHeight)
		m.searchInput.SetWidth(panelWidth - 20)

		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if sel := m.resultList.Selected(); sel != nil {
				sn := sel.Snippet
				m.selected = &sn
				// Copy snippet content to clipboard via OSC52, then quit.
				// tea.Sequence runs commands one at a time, in order.
				// This ensures the clipboard write finishes before quitting,
				// unlike tea.Batch which runs concurrently and can race.
				return m, tea.Sequence(
					tea.SetClipboard(sn.Content),
					tea.Quit,
				)
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

func (m FindModel) View() tea.View {
	if m.selected != nil || m.cancelled {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	panelWidth := findDefaultWidth
	if panelWidth > m.width-4 {
		panelWidth = m.width - 4
	}
	if panelWidth < 40 {
		panelWidth = 40
	}
	if panelWidth > m.width {
		panelWidth = m.width
	}

	innerWidth := panelWidth - 6 // border (2) + padding 2*2 (4) = 6

	// -- Header line: SNIPT badge + search input + count --
	badge := common.RenderBadgePill("SNIPT")

	countStr := fmt.Sprintf("%d/%d", len(m.filtered), len(m.allSnippets))
	countStyle := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(common.ColorBgSurface)
	count := countStyle.Render(countStr)

	// Calculate available width for search input.
	searchWidth := innerWidth - lipgloss.Width(badge) - lipgloss.Width(count) - 4
	if searchWidth < 20 {
		searchWidth = 20
	}
	m.searchInput.SetWidth(searchWidth)

	searchView := m.searchInput.View()

	// Left portion: badge + search, right-aligned count, all on one line.
	left := badge + lipgloss.NewStyle().Background(common.ColorBgSurface).Render("  ") + searchView
	leftWidth := innerWidth - lipgloss.Width(count)
	if leftWidth < 1 {
		leftWidth = 1
	}
	leftStyled := lipgloss.NewStyle().
		Width(leftWidth).
		MaxWidth(leftWidth).
		Background(common.ColorBgSurface).
		Render(left)
	headerContent := leftStyled + count

	// Wrap full header in bgSurface at exact inner width.
	header := lipgloss.NewStyle().
		Width(innerWidth).
		Background(common.ColorBgSurface).
		Render(headerContent)

	// -- Separator (between header and results) --
	sepStyle := lipgloss.NewStyle().
		Foreground(common.ColorBorderDim).
		Background(common.ColorBgSurface)
	separator := sepStyle.Render(strings.Repeat("\u2500", innerWidth))

	// -- Result list --
	listView := m.resultList.View()

	// Wrap list in bgSurface so non-selected rows get the panel background.
	listView = lipgloss.NewStyle().
		Width(innerWidth).
		Background(common.ColorBgSurface).
		Render(listView)

	// -- Bottom separator (above hint) --
	bottomSep := sepStyle.Render(strings.Repeat("\u2500", innerWidth))

	// -- Hint line --
	hintContent := lipgloss.NewStyle().
		Foreground(common.ColorTextDim).
		Background(common.ColorBgSurface).
		Render("\u2191\u2193 navigate  enter copy to clipboard  esc close")

	hint := lipgloss.NewStyle().
		Width(innerWidth).
		Background(common.ColorBgSurface).
		Render(hintContent)

	// -- Compose the inner content --
	inner := lipgloss.JoinVertical(lipgloss.Left,
		header,
		separator,
		listView,
		bottomSep,
		hint,
	)

	// -- Wrap in a rounded border --
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorBorder).
		Background(common.ColorBgSurface).
		Padding(1, 2).
		Width(panelWidth)

	panelStr := borderStyle.Render(inner)

	// -- Compositor layers --
	// Background layer (fills terminal).
	bgLines := make([]string, m.height)
	for i := range bgLines {
		bgLines[i] = strings.Repeat(" ", m.width)
	}
	bgStr := strings.Join(bgLines, "\n")
	bgLayer := lipgloss.NewLayer(bgStr)

	// Panel layer (centered, floating).
	panelH := lipgloss.Height(panelStr)
	panelW := lipgloss.Width(panelStr)
	px := (m.width - panelW) / 2
	py := (m.height - panelH) / 3
	if px < 0 {
		px = 0
	}
	if py < 0 {
		py = 0
	}
	// Prevent right-edge bleed.
	if px+panelW > m.width {
		px = m.width - panelW
		if px < 0 {
			px = 0
		}
	}
	panelLayer := lipgloss.NewLayer(panelStr).X(px).Y(py).Z(1)

	output := lipgloss.NewCompositor(bgLayer, panelLayer).Render()

	v := tea.NewView(output)
	v.AltScreen = true
	return v
}

// RunFind launches the find palette TUI and returns the selected snippet, or nil if cancelled.
func RunFind(snippets []model.Snippet, initialQuery string) (*model.Snippet, error) {
	m := NewFindModel(snippets, initialQuery)
	p := tea.NewProgram(m)

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
