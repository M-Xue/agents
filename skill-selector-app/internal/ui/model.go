package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/maxx/skill-selector-app/internal/app"
	"github.com/maxx/skill-selector-app/internal/fuzzy"
	"github.com/maxx/skill-selector-app/internal/ui/component/checklist"
	"github.com/maxx/skill-selector-app/internal/ui/component/panel"
	"github.com/maxx/skill-selector-app/internal/ui/component/searchbar"
	"github.com/maxx/skill-selector-app/internal/ui/theme"
)

// appName is rendered as the title tab above the panels.
const appName = "skill-selector"

// listTopPadding is the number of blank rows between a panel's top border and
// the first row of content inside it.
const listTopPadding = 1

// searchFieldWidth caps the search field so it reads as a compact input
// rather than stretching across the panel.
const searchFieldWidth = 30

// searchRows is how many interior rows the search occupies above the list:
// the search row itself and a blank line beneath it.
const searchRows = 2

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
	search  searchbar.Model
	skills  checklist.Model
	help    help.Model
	keys    keyMap
}

// New returns a Model ready to be passed to tea.NewProgram.
func New(application *app.App) *Model {
	h := help.New()
	h.ShowAll = false
	m := &Model{
		app:    application,
		help:   h,
		keys:   defaultKeyMap(),
		search: searchbar.New(""),
		skills: checklist.New(),
	}
	for i := range m.panels {
		m.panels[i] = panel.New(panelTitles[i], 0, 0)
	}
	m.syncSkills()
	return m
}

// syncSkills rebuilds the checklist rows: app state is mapped to items, then
// piped through the fuzzy predicate built from the search value. The search
// bar and checklist never see each other; this is the only place they meet.
// The checklist keeps its cursor on the same ID when it survives the filter
// and falls back to the top row when it does not.
func (m *Model) syncSkills() {
	skills := m.app.Skills()
	items := make([]checklist.Item, 0, len(skills))
	for _, s := range skills {
		items = append(items, checklist.Item{ID: s.ID, Checked: m.app.IsSelected(s.ID), Label: s.Label})
	}
	match := fuzzy.Predicate(m.search.Value())
	items = fuzzy.Filter(items, func(it checklist.Item) bool { return match(it.Label) })
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
			m.setFocus((m.focused + 1) % panelCount)
		case key.Matches(msg, m.keys.PrevPanel):
			m.setFocus((m.focused + panelCount - 1) % panelCount)
		case m.focused == panelSkills:
			m.updateSkills(msg)
		}
	}
	return m, nil
}

// setFocus moves keyboard focus to the given panel and focuses or blurs the
// components that live in it, so focus state is settled before the next key
// arrives rather than at render time.
func (m *Model) setFocus(id panelID) {
	m.focused = id
	for i := range m.panels {
		if panelID(i) == id {
			m.panels[i].Focus()
		} else {
			m.panels[i].Blur()
		}
	}
	if id == panelSkills {
		m.search.Focus()
	} else {
		m.search.Blur()
	}
}

// updateSkills handles keys while the Skills panel has focus. Arrows and
// enter drive the list; everything else goes to the search bar, and any key
// it consumes re-runs the filter.
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
	default:
		if m.search.Update(msg) {
			m.syncSkills()
		}
	}
}

// searchLabel renders the caption beside the search field in the active
// accent colour.
func searchLabel(text string) string {
	return lipgloss.NewStyle().Foreground(theme.BorderActive).Bold(true).Render(text)
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
		content := ""
		if panelID(i) == panelSkills {
			content = m.skillsContent(p)
		}
		views[i] = p.View(content)
	}
	body := panel.Tab(appName) + "\n\n" + sideBySide(views, widths, panelHeight)

	m.help.Width = contentWidth
	footer := m.help.ShortHelpView(m.keys.ShortHelp())
	return frame(body, footer, m.width, m.height)
}

// skillsContent lays out the Skills panel interior: a blank row, the
// labelled search field, a blank row, then the checklist filling the rest.
func (m *Model) skillsContent(p *panel.Model) string {
	label := searchLabel("Search") + " "
	m.search.SetWidth(max(0, min(searchFieldWidth, p.InnerWidth()-lipgloss.Width(label))))
	searchRow := label + m.search.View()

	listHeight := max(0, p.InnerHeight()-listTopPadding-searchRows)
	m.skills.SetSize(p.InnerWidth(), listHeight)

	rows := make([]string, 0, listTopPadding+searchRows+listHeight)
	for i := 0; i < listTopPadding; i++ {
		rows = append(rows, "")
	}
	rows = append(rows, searchRow, "")
	rows = append(rows, strings.Split(m.skills.View(), "\n")...)
	return strings.Join(rows, "\n")
}
