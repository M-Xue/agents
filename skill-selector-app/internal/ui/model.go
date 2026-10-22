package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxx/skill-selector-app/internal/ui/components/panel"
)

// appName is rendered as the title tab above the panels.
const appName = "skill-selector"

// panelID indexes the three side-by-side panels, left to right.
type panelID int

const (
	panelLibraries panelID = iota
	panelSkills
	panelDetails
	panelCount
)

// panelTitles are the tab labels, indexed by panelID.
var panelTitles = [panelCount]string{"Skill Libraries", "Skills", "Details"}

// panelRatios are the relative widths of the panels, indexed by panelID.
var panelRatios = [panelCount]int{1, 1, 2}

// Model is the root Bubble Tea model for the application.
type Model struct {
	width   int
	height  int
	focused panelID
	help    help.Model
	keys    keyMap
}

// New returns a Model ready to be passed to tea.NewProgram.
func New() *Model {
	h := help.New()
	h.ShowAll = false
	return &Model{help: h, keys: defaultKeyMap()}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextPanel):
			m.focused = (m.focused + 1) % panelCount
		case key.Matches(msg, m.keys.PrevPanel):
			m.focused = (m.focused + panelCount - 1) % panelCount
		}
	}
	return m, nil
}

// View composes the app tab, a blank line, then the three panels boxed
// lazygit-style in a 1:1:2 split: Skill Libraries, Skills, Details. The
// focused panel takes the active border. The footer sits at the bottom of
// the padded content area.
func (m *Model) View() string {
	contentWidth := contentWidthFor(m.width)
	bodyHeight := max(0, m.height-verticalPadding*2-footerRows)

	// App tab plus blank line take two rows above the panels.
	panelHeight := max(2, bodyHeight-2)
	widths := splitWidths(contentWidth, panelRatios[:]...)

	panels := make([]string, panelCount)
	for i := range panels {
		id := panelID(i)
		panels[i] = panel.Render(panelTitles[id], "", widths[i], panelHeight, id == m.focused)
	}
	body := panel.Tab(appName) + "\n\n" + sideBySide(panels, widths, panelHeight)

	m.help.Width = contentWidth
	footer := m.help.ShortHelpView(m.keys.ShortHelp())
	return frame(body, footer, m.width, m.height)
}
