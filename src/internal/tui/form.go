package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	fieldTitle = iota
	fieldLanguage
	fieldTags
	fieldDescription
	fieldCount
)

// FormResult holds the metadata from the new-snippet form.
type FormResult struct {
	Title       string
	Language    string
	Tags        string
	Description string
	Cancelled   bool
}

// formModel is the Bubbletea model for the new-snippet metadata form.
type formModel struct {
	inputs  []textinput.Model
	focused int
	result  FormResult
	done    bool
}

// Styles for the form, using Catppuccin Mocha palette.
var (
	formTitleStyle = lipgloss.NewStyle().
			Foreground(ColorMauve).
			Bold(true).
			MarginBottom(1)

	formLabelStyle = lipgloss.NewStyle().
			Foreground(ColorTextSub).
			Width(14)

	formLabelFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorMauve).
				Bold(true).
				Width(14)

	formBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	formBorderFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderFocus).
				Padding(1, 2)

	formHintStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			MarginTop(1)
)

func newStyledInput(placeholder string, charLimit, width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = charLimit
	ti.SetWidth(width)
	s := ti.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(ColorMauve)
	s.Focused.Text = lipgloss.NewStyle().Foreground(ColorText)
	s.Focused.Placeholder = lipgloss.NewStyle().Foreground(ColorTextMuted)
	s.Blurred.Prompt = lipgloss.NewStyle().Foreground(ColorTextDim)
	s.Blurred.Text = lipgloss.NewStyle().Foreground(ColorTextDim)
	s.Blurred.Placeholder = lipgloss.NewStyle().Foreground(ColorTextMuted)
	ti.SetStyles(s)
	return ti
}

func newFormModel(defaultLang string) formModel {
	inputs := make([]textinput.Model, fieldCount)

	inputs[fieldTitle] = newStyledInput("My snippet title", 120, 40)
	inputs[fieldLanguage] = newStyledInput("go, python, bash...", 30, 40)
	if defaultLang != "" {
		inputs[fieldLanguage].SetValue(defaultLang)
	}
	inputs[fieldTags] = newStyledInput("api, utils, auth (comma-separated)", 200, 40)
	inputs[fieldDescription] = newStyledInput("Brief description of this snippet", 300, 40)

	// Focus the first field
	inputs[fieldTitle].Focus()

	return formModel{
		inputs:  inputs,
		focused: fieldTitle,
	}
}

func (m formModel) Init() tea.Cmd {
	return nil
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.result.Cancelled = true
			m.done = true
			return m, tea.Quit

		case "tab", "down":
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()

		case "shift+tab", "up":
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			return m, m.updateFocus()

		case "enter":
			if m.focused == fieldDescription {
				// Submit the form
				m.result = FormResult{
					Title:       strings.TrimSpace(m.inputs[fieldTitle].Value()),
					Language:    strings.TrimSpace(m.inputs[fieldLanguage].Value()),
					Tags:        strings.TrimSpace(m.inputs[fieldTags].Value()),
					Description: strings.TrimSpace(m.inputs[fieldDescription].Value()),
				}
				m.done = true
				return m, tea.Quit
			}
			// Move to next field on enter
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		}
	}

	// Update the focused input
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *formModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
	for i := range m.inputs {
		if i == m.focused {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m *formModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m formModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}

	labels := []string{"Title", "Language", "Tags", "Description"}

	var fields strings.Builder
	for i, input := range m.inputs {
		label := formLabelStyle.Render(labels[i])
		if i == m.focused {
			label = formLabelFocusedStyle.Render(labels[i])
		}
		fields.WriteString(fmt.Sprintf("%s %s\n", label, input.View()))
		if i < fieldCount-1 {
			fields.WriteString("\n")
		}
	}

	borderStyle := formBorderStyle
	if m.focused >= 0 {
		borderStyle = formBorderFocusedStyle
	}

	title := formTitleStyle.Render("New Snippet")
	form := borderStyle.Render(fields.String())
	hint := formHintStyle.Render("tab/shift+tab: navigate  enter: next/submit  esc: cancel")

	return tea.NewView(fmt.Sprintf("\n%s\n%s\n%s\n", title, form, hint))
}

// RunForm launches a Bubbletea form to collect snippet metadata.
// defaultLang is pre-filled in the language field.
func RunForm(defaultLang string) (FormResult, error) {
	m := newFormModel(defaultLang)
	p := tea.NewProgram(m)

	final, err := p.Run()
	if err != nil {
		return FormResult{}, fmt.Errorf("run form: %w", err)
	}

	fm := final.(formModel)
	return fm.result, nil
}
