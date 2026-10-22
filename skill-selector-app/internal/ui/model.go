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
	"github.com/maxx/skill-selector-app/internal/ui/component/fileview"
	"github.com/maxx/skill-selector-app/internal/ui/component/list"
	"github.com/maxx/skill-selector-app/internal/ui/component/markdown"
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
	// Skill Libraries panel: a plain list with its own search.
	libSearch searchbar.Model
	libraries list.Model
	// Skills panel: a checklist with its own search.
	search searchbar.Model
	skills checklist.Model
	// Details panel: the selected skill's file, rendered from markdown at
	// detailsWidth. detailsSrc is kept so a resize can re-render.
	details      fileview.Model
	detailsSrc   string
	detailsWidth int
	help         help.Model
	keys         keyMap
}

// New returns a Model ready to be passed to tea.NewProgram.
func New(application *app.App) *Model {
	h := help.New()
	h.ShowAll = false
	m := &Model{
		app:       application,
		help:      h,
		keys:      defaultKeyMap(),
		libSearch: searchbar.New(""),
		libraries: list.New(),
		search:    searchbar.New(""),
		skills:    checklist.New(),
		details:   fileview.New(),
	}
	for i := range m.panels {
		m.panels[i] = panel.New(panelTitles[i], 0, 0)
	}
	m.setFocus(panelLibraries)
	m.syncLibraries()
	m.syncSkills()
	m.syncDetails()
	return m
}

// syncDetails loads the file for the skill under the cursor and renders it
// into the details viewer at the last known width.
func (m *Model) syncDetails() {
	id := ""
	if sel, ok := m.skills.Selected(); ok {
		id = sel.ID
	}
	m.detailsSrc = m.app.SkillFile(id)
	m.renderDetails()
	m.details.ScrollToTop()
}

// renderDetails pipes the markdown source through the markdown component at
// the current details width and hands the styled text to the file viewer.
func (m *Model) renderDetails() {
	m.details.SetContent(markdown.Render(m.detailsSrc, m.detailsWidth))
}

// syncLibraries rebuilds the library rows from app state, piped through the
// fuzzy predicate built from the libraries search value.
func (m *Model) syncLibraries() {
	libs := m.app.Libraries()
	items := make([]list.Item, 0, len(libs))
	for _, l := range libs {
		items = append(items, list.Item{ID: l.ID, Label: l.Label})
	}
	match := fuzzy.Predicate(m.libSearch.Value())
	items = fuzzy.Filter(items, func(it list.Item) bool { return match(it.Label) })
	m.libraries.SetItems(items)
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
		case m.focused == panelLibraries:
			m.updateLibraries(msg)
		case m.focused == panelSkills:
			m.updateSkills(msg)
		case m.focused == panelDetails:
			m.updateDetails(msg)
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
	if id == panelLibraries {
		m.libSearch.Focus()
	} else {
		m.libSearch.Blur()
	}
	if id == panelSkills {
		m.search.Focus()
	} else {
		m.search.Blur()
	}
}

// updateLibraries handles keys while the Skill Libraries panel has focus.
// Arrows move the cursor; everything else goes to the search bar.
func (m *Model) updateLibraries(msg tea.KeyMsg) {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.libraries.MoveUp()
	case key.Matches(msg, m.keys.Down):
		m.libraries.MoveDown()
	default:
		if m.libSearch.Update(msg) {
			m.syncLibraries()
		}
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

// updateDetails handles keys while the Details panel has focus: arrows
// scroll the file.
func (m *Model) updateDetails(msg tea.KeyMsg) {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.details.ScrollUp()
	case key.Matches(msg, m.keys.Down):
		m.details.ScrollDown()
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
		switch panelID(i) {
		case panelLibraries:
			content = m.librariesContent(p)
		case panelSkills:
			content = m.skillsContent(p)
		case panelDetails:
			content = m.detailsContent(p)
		}
		views[i] = p.View(content)
	}
	body := panel.Tab(appName) + "\n\n" + sideBySide(views, widths, panelHeight)

	m.help.Width = contentWidth
	footer := m.help.ShortHelpView(m.keys.ShortHelp())
	return frame(body, footer, m.width, m.height)
}

// librariesContent lays out the Skill Libraries panel interior: search bar
// on top, the plain list filling the rest.
func (m *Model) librariesContent(p *panel.Model) string {
	listHeight := searchPanelListHeight(p)
	m.libraries.SetSize(p.InnerWidth(), listHeight)
	return searchPanelContent(p, &m.libSearch, m.libraries.View())
}

// skillsContent lays out the Skills panel interior: search bar on top, the
// checklist filling the rest.
func (m *Model) skillsContent(p *panel.Model) string {
	listHeight := searchPanelListHeight(p)
	m.skills.SetSize(p.InnerWidth(), listHeight)
	return searchPanelContent(p, &m.search, m.skills.View())
}

// detailsContent lays out the Details panel interior: a blank row, then the
// file viewer filling the rest. Markdown is wrapped by the renderer, so a
// width change re-renders rather than letting the viewer re-wrap.
func (m *Model) detailsContent(p *panel.Model) string {
	if w := p.InnerWidth(); w != m.detailsWidth {
		m.detailsWidth = w
		m.renderDetails()
	}
	m.details.SetSize(p.InnerWidth(), max(0, p.InnerHeight()-listTopPadding))
	return strings.Repeat("\n", listTopPadding) + m.details.View()
}

// searchPanelListHeight is the rows left for a list beneath a panel's top
// padding and search rows.
func searchPanelListHeight(p *panel.Model) int {
	return max(0, p.InnerHeight()-listTopPadding-searchRows)
}

// searchPanelContent stacks a panel interior: a blank row, the labelled
// search field sized to the panel, a blank row, then body.
func searchPanelContent(p *panel.Model, search *searchbar.Model, body string) string {
	label := searchLabel("Search") + " "
	search.SetWidth(max(0, min(searchFieldWidth, p.InnerWidth()-lipgloss.Width(label))))
	rows := make([]string, 0, listTopPadding+searchRows)
	for i := 0; i < listTopPadding; i++ {
		rows = append(rows, "")
	}
	rows = append(rows, label+search.View(), "")
	rows = append(rows, strings.Split(body, "\n")...)
	return strings.Join(rows, "\n")
}
