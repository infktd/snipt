package picker

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/model"
	"github.com/infktd/snipt/src/internal/tui/common"
	"github.com/infktd/snipt/src/internal/tui/components"
)

const (
	pickerDefaultWidth  = 60
	pickerDefaultHeight = 16
)

// PickerModel is a stripped-down Bubbletea model for choosing among ambiguous matches.
// It reuses ResultList for rendering and navigation.
type PickerModel struct {
	resultList components.ResultList
	ref        string         // the original reference string
	selected   *model.Snippet // final selection (nil if cancelled)
	cancelled  bool
	width      int
	height     int
}

// NewPickerModel creates a new mini-picker model.
func NewPickerModel(results []model.SearchResult, ref string) PickerModel {
	items := make([]components.ResultItem, len(results))
	for i, r := range results {
		items[i] = components.ResultItem{
			Snippet: r.Snippet,
			FuzzyResult: components.FuzzyResult{
				Match:   true,
				Score:   0,
				Indices: r.TitleIndices,
			},
		}
	}

	listHeight := pickerDefaultHeight - 5 // header + border + hint
	if listHeight < 3 {
		listHeight = 3
	}
	if listHeight > len(items) {
		listHeight = len(items)
	}

	rl := components.NewResultList(pickerDefaultWidth-4, listHeight)
	rl.SetItems(items)

	return PickerModel{
		resultList: rl,
		ref:        ref,
		width:      pickerDefaultWidth,
		height:     pickerDefaultHeight,
	}
}

func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		panelWidth := pickerDefaultWidth
		if panelWidth > m.width-4 {
			panelWidth = m.width - 4
		}

		listHeight := m.height - 5
		if listHeight < 3 {
			listHeight = 3
		}
		if listHeight > m.resultList.Len() {
			listHeight = m.resultList.Len()
		}

		m.resultList.SetWidth(panelWidth - 4)
		m.resultList.SetHeight(listHeight)

		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
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

	return m, nil
}

func (m PickerModel) View() tea.View {
	if m.selected != nil || m.cancelled {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	panelWidth := pickerDefaultWidth
	if panelWidth > m.width-4 {
		panelWidth = m.width - 4
	}
	if panelWidth < 30 {
		panelWidth = 30
	}

	innerWidth := panelWidth - 4

	// -- Header --
	headerStyle := lipgloss.NewStyle().
		Foreground(common.ColorYellow).
		Bold(true)
	header := headerStyle.Render(fmt.Sprintf("Multiple matches for %q", m.ref))

	// Truncate header if too wide.
	if lipgloss.Width(header) > innerWidth {
		header = headerStyle.Render("Multiple matches")
	}

	countStyle := lipgloss.NewStyle().Foreground(common.ColorTextDim)
	count := countStyle.Render(fmt.Sprintf(" (%d)", m.resultList.Len()))
	header = header + count

	// -- Separator --
	sepStyle := lipgloss.NewStyle().Foreground(common.ColorBorderDim)
	separator := sepStyle.Render(strings.Repeat("\u2500", innerWidth))

	// -- Result list --
	listView := m.resultList.View()

	// -- Compose inner content --
	inner := lipgloss.JoinVertical(lipgloss.Left,
		header,
		separator,
		listView,
	)

	// -- Wrap in rounded border --
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.ColorBorder).
		Padding(0, 1).
		Width(panelWidth)

	panel := borderStyle.Render(inner)

	// -- Hint line --
	hintStyle := lipgloss.NewStyle().Foreground(common.ColorTextDim)
	hint := hintStyle.Render("  \u2191\u2193 navigate  enter select  esc cancel")

	full := lipgloss.JoinVertical(lipgloss.Left, panel, hint)

	// Center horizontally.
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(full)

	// Vertical padding.
	topPad := (m.height - lipgloss.Height(centered)) / 3
	if topPad < 0 {
		topPad = 0
	}

	v := tea.NewView(strings.Repeat("\n", topPad) + centered)
	v.AltScreen = true
	return v
}

// RunPicker launches the mini-picker TUI and returns the selected snippet, or nil if cancelled.
func RunPicker(results []model.SearchResult, ref string) (*model.Snippet, error) {
	m := NewPickerModel(results, ref)
	p := tea.NewProgram(m)

	final, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("run picker: %w", err)
	}

	pm := final.(PickerModel)
	if pm.cancelled {
		return nil, nil
	}

	return pm.selected, nil
}
