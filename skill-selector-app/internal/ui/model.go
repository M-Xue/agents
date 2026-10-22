package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// panelCount is the number of panels shown on the single screen.
const panelCount = 3

// panelTitles are placeholder headings until the panels have real content.
var panelTitles = [panelCount]string{"Panel 1", "Panel 2", "Panel 3"}

// Model is the root Bubble Tea model for the application.
type Model struct {
	focused int // index of the panel that has focus
	width   int
	height  int
	help    help.Model
	keys    keyMap
}

// New returns a Model ready to be passed to tea.NewProgram.
func New() Model {
	return Model{
		help: help.New(),
		keys: keys,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextPanel):
			m.focused = (m.focused + 1) % panelCount
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	helpView := m.help.View(m.keys)

	// Reserve one line for help and account for each panel's border.
	panelWidth := m.width / panelCount
	panelHeight := m.height - lipgloss.Height(helpView)

	innerWidth := panelWidth - panelStyle.GetHorizontalFrameSize()
	innerHeight := panelHeight - panelStyle.GetVerticalFrameSize()

	panels := make([]string, 0, panelCount)
	for i := 0; i < panelCount; i++ {
		style := panelStyle
		if i == m.focused {
			style = focusedPanelStyle
		}
		content := titleStyle.Render(panelTitles[i])
		panels = append(panels, style.
			Width(innerWidth).
			Height(innerHeight).
			Render(content))
	}

	var b strings.Builder
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, panels...))
	b.WriteString("\n")
	b.WriteString(helpView)
	return b.String()
}
