package manage

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/tui/common"
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
	editor string // resolved editor command

	width    int
	height   int
	mode     mode
	quitting bool
}

// NewManageModel creates a new manage screen model.
func NewManageModel(store *db.Store, editor string) ManageModel {
	return ManageModel{
		store:  store,
		editor: editor,
		mode:   modeNormal,
	}
}

func (m ManageModel) Init() tea.Cmd {
	return nil
}

func (m ManageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ManageModel) View() tea.View {
	if m.quitting {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	// Placeholder: just show "SNIPT MANAGE" centered.
	placeholder := lipgloss.NewStyle().
		Foreground(common.ColorText).
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(fmt.Sprintf("SNIPT MANAGE\n\n%dx%d\n\nPress q to quit", m.width, m.height))

	v := tea.NewView(placeholder)
	v.AltScreen = true
	return v
}

// RunManage launches the manage TUI.
func RunManage(store *db.Store, editor string) error {
	m := NewManageModel(store, editor)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
