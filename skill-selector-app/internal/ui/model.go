package ui

import (
	"fmt"
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
	// libraryID is the library whose skills are loaded, so moving the
	// library cursor only reloads when it lands on a different library.
	libraryID string
	// Details panel: the selected skill's file, rendered from markdown at
	// detailsWidth. detailsSrc is kept so a resize can re-render and
	// detailsID so moving within the list only re-reads on a new skill.
	details      fileview.Model
	detailsID    string
	detailsSrc   string
	detailsWidth int
	// status is an error from the last action, shown in the footer in place
	// of the help until the next key.
	status string
	help   help.Model
	keys   keyMap
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
	return m
}

// syncDetails shows the file of the skill under the Skills cursor. It only
// re-reads and rewinds when the cursor has moved to a different skill, so
// re-filtering that leaves the cursor in place keeps the reader's position.
func (m *Model) syncDetails() {
	id := ""
	if sel, ok := m.skills.Selected(); ok {
		id = sel.ID
	}
	if id == m.detailsID && m.detailsSrc != "" {
		return
	}
	m.detailsID = id
	m.detailsSrc = m.app.SkillFile(id)
	m.renderDetails()
	m.details.ScrollToTop()
}

// syncLibrary makes the library under the Skill Libraries cursor current in
// the app and rebuilds the Skills panel from it. Between two installed
// libraries the checklist snaps to its first row because no skill ID from the
// old library survives, which is how the first skill of a newly chosen
// library ends up selected. Moving between an installed library and the
// built-in All library keeps the cursor on the same skill, since its ID is
// present in both. It is a no-op when the cursor is still on the same
// library.
func (m *Model) syncLibrary() {
	id := ""
	if sel, ok := m.libraries.Selected(); ok {
		id = sel.ID
	}
	if id == m.libraryID && m.skills.Len() > 0 {
		return
	}
	m.libraryID = id
	if err := m.app.SelectLibrary(id); err != nil {
		m.status = err.Error()
	}
	m.syncSkills()
}

// renderDetails pipes the markdown source through the markdown component at
// the current details width and hands the styled text to the file viewer.
func (m *Model) renderDetails() {
	m.details.SetContent(markdown.Render(m.detailsSrc, m.detailsWidth))
}

// syncLibraries rebuilds the library rows from app state, piped through the
// fuzzy predicate built from the libraries search value. The built-in All
// library is filtered like any other row.
func (m *Model) syncLibraries() {
	libs := m.app.Libraries()
	items := make([]list.Item, 0, len(libs))
	for _, l := range libs {
		items = append(items, list.Item{ID: l.ID, Label: l.Label})
	}
	match := fuzzy.Predicate(m.libSearch.Value())
	items = fuzzy.Filter(items, func(it list.Item) bool { return match(it.Label) })
	m.libraries.SetItems(items)
	m.syncLibrary()
}

// syncSkills rebuilds the checklist rows: app state is mapped to items, then
// piped through the fuzzy predicate built from the search value. The search
// bar and checklist never see each other; this is the only place they meet.
// The checklist keeps its cursor on the same ID when it survives the filter
// and falls back to the top row when it does not. While the built-in All
// library is current, each row carries its library's name so skills that
// share a name across libraries can be told apart; the search matches the
// whole row, so a library name narrows the list too.
func (m *Model) syncSkills() {
	skills := m.app.Skills()
	items := make([]checklist.Item, 0, len(skills))
	for _, s := range skills {
		shadowed := m.app.ShadowedBy(s.ID) != ""
		items = append(items, checklist.Item{
			ID:       s.ID,
			Checked:  m.app.IsSelected(s.ID),
			Disabled: shadowed,
			Label:    m.skillLabel(s, shadowed),
		})
	}
	match := fuzzy.Predicate(m.search.Value())
	items = fuzzy.Filter(items, func(it checklist.Item) bool { return match(it.Label) })
	m.skills.SetItems(items)
	m.syncDetails()
}

// shadowedSuffix is appended to the row of a skill whose name is already
// held in the exposed directory by something else.
const shadowedSuffix = "[skill with same name already exists]"

// skillLabel is the checklist row text for a skill: its name, followed by
// its library in parentheses while the All library is current, followed by
// shadowedSuffix when the skill is shadowed.
func (m *Model) skillLabel(s app.Skill, shadowed bool) string {
	label := s.Label
	if m.libraryID == app.AllLibraryID {
		label += " (" + s.Library + ")"
	}
	if shadowed {
		label += " " + shadowedSuffix
	}
	return label
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.MouseMsg:
		m.updateMouse(msg)
		return m, nil
	case tea.KeyMsg:
		m.status = ""
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

// updateMouse routes mouse events by the panel under the pointer, not by
// focus. A left click focuses that panel, or, when the panel already has
// focus, acts on the row under the pointer (see clickRow). The wheel scrolls the list or file
// in that panel one row per notch without changing focus or moving the
// cursor, so a user can read a long skill file or skim a long list while
// still typing in a search elsewhere. Any other mouse event is ignored.
func (m *Model) updateMouse(msg tea.MouseMsg) {
	id, ok := m.panelAt(msg.X, msg.Y)
	if !ok {
		return
	}
	switch {
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		m.status = ""
		if id != m.focused {
			m.setFocus(id)
			return
		}
		m.clickRow(id, msg.Y)
	case msg.Button == tea.MouseButtonWheelUp:
		m.scrollPanel(id, -1)
	case msg.Button == tea.MouseButtonWheelDown:
		m.scrollPanel(id, 1)
	}
}

// clickRow handles a left click inside an already-focused panel. In Skill
// Libraries it selects the library row under the pointer and loads its
// skills. In Skills a click on an unselected row selects it and shows its
// file; a click on the row that is already selected toggles its link, the
// same as pressing enter. Clicks on the search bar, the padding or below the
// last row do nothing. Details has no rows to click.
func (m *Model) clickRow(id panelID, y int) {
	row, ok := m.listRowAt(y)
	if !ok {
		return
	}
	switch id {
	case panelLibraries:
		if m.libraries.Select(m.libraries.Offset() + row) {
			m.syncLibrary()
		}
	case panelSkills:
		idx := m.skills.Offset() + row
		if idx != m.skills.Cursor() {
			if m.skills.Select(idx) {
				m.syncDetails()
			}
			return
		}
		if sel, ok := m.skills.Selected(); ok {
			if err := m.app.ToggleSkill(sel.ID); err != nil {
				m.status = err.Error()
			}
			m.syncSkills()
		}
	}
}

// listRowAt maps a window row to a visible row index of a search panel's
// list: 0 is the first row beneath the search bar. It reports false when y
// falls on the border, padding, search rows or past the list's height.
// Both search panels share the same vertical layout, so one mapping serves
// Skill Libraries and Skills alike.
func (m *Model) listRowAt(y int) (int, bool) {
	_, panelHeight := m.panelGeometry()
	// Top border, then the interior rows above the list.
	first := verticalPadding + panelTopRows + 1 + listTopPadding + searchRows
	listHeight := max(0, panelHeight-2-listTopPadding-searchRows)
	row := y - first
	if row < 0 || row >= listHeight {
		return 0, false
	}
	return row, true
}

// scrollPanel scrolls the scrollable component in the given panel by one
// row: up when dir is negative, down otherwise.
func (m *Model) scrollPanel(id panelID, dir int) {
	switch id {
	case panelLibraries:
		if dir < 0 {
			m.libraries.ScrollUp()
		} else {
			m.libraries.ScrollDown()
		}
	case panelSkills:
		if dir < 0 {
			m.skills.ScrollUp()
		} else {
			m.skills.ScrollDown()
		}
	case panelDetails:
		if dir < 0 {
			m.details.ScrollUp()
		} else {
			m.details.ScrollDown()
		}
	}
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
// Arrows move the cursor and load that library's skills; everything else
// goes to the search bar, which re-filters and may land on a new library.
func (m *Model) updateLibraries(msg tea.KeyMsg) {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.libraries.MoveUp()
		m.syncLibrary()
	case key.Matches(msg, m.keys.Down):
		m.libraries.MoveDown()
		m.syncLibrary()
	default:
		if m.libSearch.Update(msg) {
			m.syncLibraries()
		}
	}
}

// updateSkills handles keys while the Skills panel has focus. Arrows move
// the cursor and show that skill's file; enter toggles the link for the
// skill under the cursor, reporting a failure in the footer; everything else
// goes to the search bar, and any key it consumes re-runs the filter.
func (m *Model) updateSkills(msg tea.KeyMsg) {
	switch {
	case key.Matches(msg, m.keys.Up):
		m.skills.MoveUp()
		m.syncDetails()
	case key.Matches(msg, m.keys.Down):
		m.skills.MoveDown()
		m.syncDetails()
	case key.Matches(msg, m.keys.Toggle):
		if sel, ok := m.skills.Selected(); ok {
			if err := m.app.ToggleSkill(sel.ID); err != nil {
				m.status = err.Error()
			}
			m.syncSkills()
		}
	case key.Matches(msg, m.keys.ToggleAll):
		m.toggleAll()
		m.syncSkills()
	default:
		if m.search.Update(msg) {
			m.syncSkills()
		}
	}
}

// toggleAll acts on the rows currently shown in the checklist, so a search
// narrows what it touches. Shadowed rows, whose name another exposed skill
// already holds, are left out of the decision and the sweep: they count as
// toggled. So if any shown skill is unselected and not shadowed, every such
// skill is selected; if every shown skill is selected or shadowed, all are
// deselected. While selecting, a skill whose name was taken earlier in the
// same sweep is skipped silently rather than reported, since it is shadowed
// by the time its turn comes. Any other failure is skipped so one problem
// does not stop the rest, and the failures are summarised in the footer.
func (m *Model) toggleAll() {
	items := m.skills.Items()
	if len(items) == 0 {
		return
	}
	target := false
	for _, it := range items {
		if !it.Checked && !it.Disabled {
			target = true
			break
		}
	}
	failed, first := 0, ""
	for _, it := range items {
		if it.Checked == target || it.Disabled {
			continue
		}
		if target && m.app.ShadowedBy(it.ID) != "" {
			continue
		}
		if err := m.app.SetSelected(it.ID, target); err != nil {
			failed++
			if first == "" {
				first = err.Error()
			}
		}
	}
	switch failed {
	case 0:
	case 1:
		m.status = first
	default:
		m.status = fmt.Sprintf("%d skills could not be changed; first: %s", failed, first)
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

// panelTopRows is the number of rows above the panels inside the content
// area: the app tab and the blank line beneath it.
const panelTopRows = 2

// panelGeometry returns the column widths of the panels and their shared
// height for the current window size. View and panelAt both use it so what
// is drawn and what is clickable always agree.
func (m *Model) panelGeometry() (widths []int, height int) {
	contentWidth := contentWidthFor(m.width)
	bodyHeight := max(0, m.height-verticalPadding*2-footerRows)
	return splitWidths(contentWidth, panelRatios[:]...), max(2, bodyHeight-panelTopRows)
}

// panelAt returns the panel drawn at window cell (x, y), or false when the
// cell is padding, the title row, a gap between panels or the footer.
func (m *Model) panelAt(x, y int) (panelID, bool) {
	widths, height := m.panelGeometry()
	top := verticalPadding + panelTopRows
	if y < top || y >= top+height {
		return 0, false
	}
	left := horizontalPadding
	for i, w := range widths {
		if x >= left && x < left+w {
			return panelID(i), true
		}
		left += w + panelGap
	}
	return 0, false
}

// View composes the app tab, a blank line, then the three panels boxed
// lazygit-style in a 1:1:2 split: Skill Libraries, Skills, Details. The
// focused panel takes the active border. The footer sits at the bottom of
// the padded content area.
func (m *Model) View() string {
	contentWidth := contentWidthFor(m.width)
	widths, panelHeight := m.panelGeometry()

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
	body := panel.Tab(appName) + strings.Repeat("\n", panelTopRows) + sideBySide(views, widths, panelHeight)

	m.help.Width = contentWidth
	footer := m.help.ShortHelpView(m.keys.helpFor(m.focused))
	if m.status != "" {
		footer = lipgloss.NewStyle().Foreground(theme.Error).Render(m.status)
	}
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
	if m.detailsID == "" {
		empty := lipgloss.NewStyle().Foreground(theme.TextInactive).Render("No skill selected")
		return strings.Repeat("\n", listTopPadding) + empty
	}
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
