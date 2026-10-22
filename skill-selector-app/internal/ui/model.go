package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxx/skill-selector-app/internal/app"
	"github.com/maxx/skill-selector-app/internal/ui/component/checklist"
	"github.com/maxx/skill-selector-app/internal/ui/component/panel"
)

// appName is rendered as the title tab above the panels.
const appName = "skill-selector"

// listTopPadding is the number of blank rows between a panel's top border and
// the first row of the list inside it.
const listTopPadding = 1

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
	app     *app.App
	width   int
	height  int
	focused panelID
	panels  [panelCount]panel.Model
	skills  checklist.Model
	help    help.Model
	keys    keyMap
}

// New returns a Model ready to be passed to tea.NewProgram.
func New(application *app.App) *Model {
	h := help.New()
	h.ShowAll = false
	m := &Model{app: application, help: h, keys: defaultKeyMap(), skills: checklist.New()}
	for i := range m.panels {
		m.panels[i] = panel.New(panelTitles[i], 0, 0)
	}
	m.syncSkills()
	return m
}

// syncSkills rebuilds the checklist rows from app state. The checklist keeps
// its own cursor; checked state always comes from app.
func (m *Model) syncSkills() {
	skills := m.app.Skills()
	items := make([]checklist.Item, 0, len(skills))
	for _, s := range skills {
		items = append(items, checklist.Item{ID: s.ID, Checked: m.app.IsSelected(s.ID), Label: s.Label})
	}
	m.skills.SetItems(items)
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
		case m.focused == panelSkills:
			m.updateSkills(msg)
		}
	}
	return m, nil
}

// updateSkills handles keys while the Skills panel has focus.
func (m *Model) updateSkills(msg tea.KeyMsg) {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.skills.MoveUp()
	case key.Matches(msg, m.keys.Down):
		m.skills.MoveDown()
	case key.Matches(msg, m.keys.Toggle):
		if sel, ok := m.skills.Selected(); ok {
			m.app.ToggleSkill(sel.ID)
			m.syncSkills()
		}
	}
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

	views := make([]string, panelCount)
	for i := range m.panels {
		p := &m.panels[i]
		p.SetSize(widths[i], panelHeight)
		if panelID(i) == m.focused {
			p.Focus()
		} else {
			p.Blur()
		}
		content := ""
		if panelID(i) == panelSkills {
			// One blank row between the top border and the first item.
			m.skills.SetSize(p.InnerWidth(), max(0, p.InnerHeight()-listTopPadding))
			content = strings.Repeat("\n", listTopPadding) + m.skills.View()
		}
		views[i] = p.View(content)
	}
	body := panel.Tab(appName) + "\n\n" + sideBySide(views, widths, panelHeight)

	m.help.Width = contentWidth
	footer := m.help.ShortHelpView(m.keys.ShortHelp())
	return frame(body, footer, m.width, m.height)
}
