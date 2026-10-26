package ui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxx/skill-selector-app/internal/app"
	"github.com/maxx/skill-selector-app/internal/skills"
)

// newModel builds a workspace with libraries alpha (skills apple, banana)
// and beta (skill cherry), each with a distinctive SKILL.md, and returns a
// sized model focused on the libraries panel.
func newModel(t *testing.T) (*Model, skills.Repo) {
	t.Helper()
	root := t.TempDir()
	write := func(lib, skill string) {
		dir := filepath.Join(root, "skill-libraries", lib, "skills", skill)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "# " + strings.ToUpper(skill) + " SKILL BODY\n"
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	write("alpha", "apple")
	write("alpha", "banana")
	write("beta", "cherry")

	repo := skills.Repo{Root: root}
	a, err := app.New(repo)
	if err != nil {
		t.Fatal(err)
	}
	m := New(a)
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 40})
	return m, repo
}

func press(m *Model, k tea.KeyType) { m.Update(tea.KeyMsg{Type: k}) }

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// view returns the rendered screen with ANSI styling stripped, since Glamour
// splits styled headings into several escape-delimited runs.
func view(m *Model) string { return ansi.ReplaceAllString(m.View(), "") }

func TestStartupSelectsFirstLibraryAndSkill(t *testing.T) {
	m, _ := newModel(t)
	v := view(m)
	for _, want := range []string{"alpha", "beta", "apple", "banana", "APPLE SKILL BODY"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q", want)
		}
	}
	if strings.Contains(v, "cherry") {
		t.Error("beta's skill shown while alpha is selected")
	}
}

func TestLibraryCursorSwapsSkillsAndDetails(t *testing.T) {
	m, _ := newModel(t)
	press(m, tea.KeyDown) // libraries: alpha -> beta
	v := view(m)
	if !strings.Contains(v, "cherry") || !strings.Contains(v, "CHERRY SKILL BODY") {
		t.Fatal("beta's first skill not shown after moving library cursor")
	}
	if strings.Contains(v, "apple") {
		t.Fatal("alpha's skills still shown")
	}
	press(m, tea.KeyUp) // back to alpha: first skill again
	if v := view(m); !strings.Contains(v, "APPLE SKILL BODY") {
		t.Fatal("alpha's first skill not restored")
	}
}

func TestSkillCursorChangesDetails(t *testing.T) {
	m, _ := newModel(t)
	press(m, tea.KeyRight) // focus skills
	press(m, tea.KeyDown)  // apple -> banana
	if v := view(m); !strings.Contains(v, "BANANA SKILL BODY") || strings.Contains(v, "APPLE SKILL BODY") {
		t.Fatal("details did not follow the skill cursor")
	}
}

func TestEnterTogglesSymlink(t *testing.T) {
	m, repo := newModel(t)
	press(m, tea.KeyRight) // focus skills, cursor on apple
	link := filepath.Join(repo.ExposedDir(), "apple")

	press(m, tea.KeyEnter)
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("enter did not create symlink: %v", err)
	}
	if m.status != "" {
		t.Fatalf("unexpected status %q", m.status)
	}
	press(m, tea.KeyEnter)
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("second enter did not remove symlink: %v", err)
	}
}

func TestToggleFailureShowsInFooter(t *testing.T) {
	m, repo := newModel(t)
	// Occupy the link name with a real directory so Link must fail.
	if err := os.Mkdir(filepath.Join(repo.ExposedDir(), "apple"), 0o755); err != nil {
		t.Fatal(err)
	}
	press(m, tea.KeyRight)
	press(m, tea.KeyEnter)
	if m.status == "" || !strings.Contains(view(m), "exposing apple") {
		t.Fatalf("failure not reported, status=%q", m.status)
	}
	press(m, tea.KeyDown)
	if m.status != "" {
		t.Fatal("status not cleared on next key")
	}
}

func TestSearchFilterMovesSelection(t *testing.T) {
	m, _ := newModel(t)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bet")}) // libraries search
	if v := view(m); !strings.Contains(v, "CHERRY SKILL BODY") {
		t.Fatal("filtering libraries did not select the surviving library")
	}
	press(m, tea.KeyRight)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("zzz")}) // no skills match
	if v := view(m); !strings.Contains(v, "No skill selected") {
		t.Fatal("empty state not shown when no skill matches")
	}
}

func click(m *Model, x, y int) {
	m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
}

func TestClickFocusesPanel(t *testing.T) {
	m, _ := newModel(t) // 160x40 window
	widths, height := m.panelGeometry()
	top := verticalPadding + panelTopRows
	row := top + height/2

	cases := []struct {
		name string
		x    int
		want panelID
	}{
		{"skills", horizontalPadding + widths[0] + panelGap + 1, panelSkills},
		{"details", horizontalPadding + widths[0] + widths[1] + 2*panelGap + 1, panelDetails},
		{"libraries", horizontalPadding + 1, panelLibraries},
	}
	for _, tc := range cases {
		click(m, tc.x, row)
		if m.focused != tc.want {
			t.Fatalf("click %s at x=%d: focused %d, want %d", tc.name, tc.x, m.focused, tc.want)
		}
	}

	// Clicks outside the panels leave focus alone: the title row, the gap
	// between panels, the left margin and the footer.
	m.setFocus(panelSkills)
	for _, p := range [][2]int{
		{horizontalPadding + 1, verticalPadding},
		{horizontalPadding + widths[0], row},
		{0, row},
		{horizontalPadding + 1, m.height - verticalPadding - 1},
	} {
		click(m, p[0], p[1])
		if m.focused != panelSkills {
			t.Fatalf("click at %v changed focus to %d", p, m.focused)
		}
	}
	// Right button and release events are ignored.
	m.Update(tea.MouseMsg{X: horizontalPadding + 1, Y: row, Action: tea.MouseActionPress, Button: tea.MouseButtonRight})
	m.Update(tea.MouseMsg{X: horizontalPadding + 1, Y: row, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	if m.focused != panelSkills {
		t.Fatal("non-left-press mouse event changed focus")
	}
}

func wheel(m *Model, x, y int, button tea.MouseButton) {
	m.Update(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: button})
}

func TestWheelScrollsPanelUnderPointerWithoutFocus(t *testing.T) {
	m, _ := newModel(t) // 160x40 window, libraries focused
	widths, height := m.panelGeometry()
	top := verticalPadding + panelTopRows
	row := top + height/2
	detailsX := horizontalPadding + widths[0] + widths[1] + 2*panelGap + 1
	skillsX := horizontalPadding + widths[0] + panelGap + 1

	// Make the details long enough to scroll.
	m.detailsSrc = strings.Repeat("line\n\n", 200) // separate paragraphs survive markdown rendering
	m.renderDetails()
	view(m) // size the components

	wheel(m, detailsX, row, tea.MouseButtonWheelDown)
	wheel(m, detailsX, row, tea.MouseButtonWheelDown)
	if m.details.Offset() != 2 {
		t.Fatalf("details offset = %d after two wheel downs, want 2", m.details.Offset())
	}
	wheel(m, detailsX, row, tea.MouseButtonWheelUp)
	if m.details.Offset() != 1 {
		t.Fatalf("details offset = %d after wheel up, want 1", m.details.Offset())
	}
	if m.focused != panelLibraries {
		t.Fatal("wheel over details changed focus")
	}

	// Shrink the window until the Skills list shows a single row, then the
	// wheel scrolls the list window while the cursor stays on apple.
	m.Update(tea.WindowSizeMsg{Width: 160, Height: 13})
	view(m)
	_, height = m.panelGeometry()
	row = verticalPadding + panelTopRows + height/2
	wheel(m, skillsX, row, tea.MouseButtonWheelDown)
	if m.skills.Offset() != 1 {
		t.Fatalf("skills offset = %d after wheel down, want 1", m.skills.Offset())
	}
	if sel, _ := m.skills.Selected(); sel.ID == "" || !strings.HasSuffix(sel.ID, "apple") {
		t.Fatalf("wheel moved the skills cursor to %q", sel.ID)
	}
	if v := view(m); !strings.Contains(v, "banana") || strings.Contains(v, "apple") {
		t.Fatalf("skills list did not scroll to banana: %q", v)
	}
	if m.focused != panelLibraries {
		t.Fatal("wheel over skills changed focus")
	}

	// Wheel events outside every panel are ignored.
	before := m.skills.Offset()
	wheel(m, 0, row, tea.MouseButtonWheelUp)
	wheel(m, skillsX, verticalPadding, tea.MouseButtonWheelUp)
	if m.skills.Offset() != before {
		t.Fatal("wheel outside the panels scrolled a list")
	}
}

func TestClickSelectsRowInFocusedPanel(t *testing.T) {
	m, repo := newModel(t) // 160x40 window, libraries focused
	widths, _ := m.panelGeometry()
	firstRow := verticalPadding + panelTopRows + 1 + listTopPadding + searchRows
	libX := horizontalPadding + 1
	skillsX := horizontalPadding + widths[0] + panelGap + 1
	linked := func(name string) bool {
		fi, err := os.Lstat(filepath.Join(repo.ExposedDir(), name))
		return err == nil && fi.Mode()&os.ModeSymlink != 0
	}

	// Libraries is focused: clicking its second row selects beta.
	click(m, libX, firstRow+1)
	if sel, _ := m.libraries.Selected(); sel.Label != "beta" {
		t.Fatalf("click did not select beta: %q", sel.Label)
	}
	if v := view(m); !strings.Contains(v, "CHERRY SKILL BODY") {
		t.Fatal("selecting beta by click did not load its skills")
	}
	click(m, libX, firstRow) // back to alpha
	if v := view(m); !strings.Contains(v, "APPLE SKILL BODY") {
		t.Fatal("click back to alpha did not reload its skills")
	}

	// Clicking the search row or past the last row changes nothing.
	click(m, libX, firstRow-searchRows)
	click(m, libX, firstRow+5)
	if sel, _ := m.libraries.Selected(); sel.Label != "alpha" {
		t.Fatalf("non-row click moved the selection to %q", sel.Label)
	}

	// Skills is not focused: the first click only focuses it.
	click(m, skillsX, firstRow+1)
	if m.focused != panelSkills {
		t.Fatal("click on unfocused skills did not focus it")
	}
	if sel, _ := m.skills.Selected(); !strings.HasSuffix(sel.ID, "apple") || linked("banana") {
		t.Fatalf("focusing click acted on a row: cursor=%q banana linked=%v", sel.ID, linked("banana"))
	}

	// Now focused: clicking banana only selects it, since it was not the
	// selected row. The details follow, but nothing is toggled yet.
	click(m, skillsX, firstRow+1)
	if sel, _ := m.skills.Selected(); !strings.HasSuffix(sel.ID, "banana") {
		t.Fatalf("click did not select banana: %q", sel.ID)
	}
	if linked("banana") {
		t.Fatal("click on an unselected skill toggled it")
	}
	if v := view(m); !strings.Contains(v, "BANANA SKILL BODY") {
		t.Fatal("details did not follow the clicked skill")
	}
	// Clicking the selected row toggles it on, and again toggles it off,
	// keeping the selection throughout.
	click(m, skillsX, firstRow+1)
	if !linked("banana") {
		t.Fatal("click on the selected skill did not link it")
	}
	click(m, skillsX, firstRow+1)
	if linked("banana") {
		t.Fatal("third click did not unlink banana")
	}
	if sel, _ := m.skills.Selected(); !strings.HasSuffix(sel.ID, "banana") {
		t.Fatalf("toggling clicks moved the cursor to %q", sel.ID)
	}
	// A click below the last skill neither moves nor toggles.
	click(m, skillsX, firstRow+2)
	if sel, _ := m.skills.Selected(); !strings.HasSuffix(sel.ID, "banana") || linked("apple") {
		t.Fatal("click past the last row acted on a skill")
	}
}

// footer returns the last non-blank line of the rendered screen.
func footer(m *Model) string {
	lines := strings.Split(view(m), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

func TestHelpFollowsFocus(t *testing.T) {
	m, _ := newModel(t)
	common := []string{"↑/shift+tab up", "↓/tab down", "→ next panel", "← prev panel", "esc quit"}
	for _, p := range []panelID{panelLibraries, panelSkills, panelDetails} {
		m.setFocus(p)
		f := footer(m)
		for _, w := range common {
			if !strings.Contains(f, w) {
				t.Errorf("panel %d footer %q missing %q", p, f, w)
			}
		}
		hasToggle := strings.Contains(f, "toggle")
		if hasToggle != (p == panelSkills) {
			t.Errorf("panel %d footer %q: toggle shown = %v", p, f, hasToggle)
		}
	}
	m.setFocus(panelSkills)
	if f := footer(m); !strings.HasSuffix(strings.TrimRight(f, " "), "enter toggle • ctrl+a toggle all") {
		t.Errorf("toggle hints are not last in %q", f)
	}
}

func TestLettersTypeIntoSearchAndEsc(t *testing.T) {
	m, _ := newModel(t)
	typed := func(r rune) { m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}) }

	// j and k are not navigation keys: they must reach the focused search
	// bar rather than move the cursor. "jk" matches no library, so the
	// filtered list empties and no skill body is shown at all.
	typed('j')
	typed('k')
	if got := m.libSearch.Value(); got != "jk" {
		t.Fatalf("library search = %q, want %q", got, "jk")
	}
	if v := view(m); strings.Contains(v, "CHERRY SKILL BODY") {
		t.Fatal("j moved the library cursor to beta")
	}
	if v := view(m); strings.Contains(v, "alpha") || strings.Contains(v, "beta") {
		t.Fatalf("typing jk did not filter the library list: %q", v)
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc did not produce a quit command")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c no longer quits")
	}
}

func TestToggleAll(t *testing.T) {
	m, repo := newModel(t)
	press(m, tea.KeyRight) // skills: apple, banana
	linked := func(name string) bool {
		fi, err := os.Lstat(filepath.Join(repo.ExposedDir(), name))
		return err == nil && fi.Mode()&os.ModeSymlink != 0
	}
	toggleAll := func() { press(m, tea.KeyCtrlA) }

	// Nothing selected: select all.
	toggleAll()
	if !linked("apple") || !linked("banana") {
		t.Fatal("toggle all did not select every skill")
	}
	// All selected: deselect all.
	toggleAll()
	if linked("apple") || linked("banana") {
		t.Fatal("toggle all did not deselect every skill")
	}
	// One selected, one not: select all (fills in the gap, keeps the other).
	press(m, tea.KeyEnter) // apple on
	toggleAll()
	if !linked("apple") || !linked("banana") {
		t.Fatal("toggle all with a mixed state did not select every skill")
	}
	// A search narrows what toggle all touches.
	toggleAll() // all off
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ban")})
	toggleAll()
	if linked("apple") || !linked("banana") {
		t.Fatal("toggle all acted outside the filtered rows")
	}
	if m.status != "" {
		t.Fatalf("unexpected status %q", m.status)
	}
}

func TestToggleAllReportsFailures(t *testing.T) {
	m, repo := newModel(t)
	// Occupy apple's link name so it cannot be linked; banana still can.
	if err := os.Mkdir(filepath.Join(repo.ExposedDir(), "apple"), 0o755); err != nil {
		t.Fatal(err)
	}
	press(m, tea.KeyRight)
	press(m, tea.KeyCtrlA)
	if fi, err := os.Lstat(filepath.Join(repo.ExposedDir(), "banana")); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("a failure on one skill stopped the others")
	}
	if !strings.Contains(m.status, "exposing apple") {
		t.Fatalf("failure not reported: %q", m.status)
	}
}

func TestTabMovesAndArrowsCyclePanels(t *testing.T) {
	m, _ := newModel(t) // libraries focused, cursor on alpha
	expectFocus := func(step string, want panelID) {
		t.Helper()
		if m.focused != want {
			t.Fatalf("%s: focused %d, want %d", step, m.focused, want)
		}
	}

	// Tab and shift+tab move the cursor like down and up, without changing
	// focus; the library change flows through to the other panels.
	press(m, tea.KeyTab)
	expectFocus("tab", panelLibraries)
	if sel, _ := m.libraries.Selected(); sel.Label != "beta" {
		t.Fatalf("tab did not move down: on %q", sel.Label)
	}
	if v := view(m); !strings.Contains(v, "CHERRY SKILL BODY") {
		t.Fatal("tab did not load beta's skills")
	}
	press(m, tea.KeyShiftTab)
	if sel, _ := m.libraries.Selected(); sel.Label != "alpha" {
		t.Fatalf("shift+tab did not move up: on %q", sel.Label)
	}

	// Arrows cycle all three panels, wrapping at both ends.
	press(m, tea.KeyRight)
	expectFocus("right from libraries", panelSkills)
	press(m, tea.KeyTab)
	if sel, _ := m.skills.Selected(); !strings.HasSuffix(sel.ID, "banana") {
		t.Fatalf("tab in skills did not move down: on %q", sel.ID)
	}
	press(m, tea.KeyRight)
	expectFocus("right from skills", panelDetails)
	press(m, tea.KeyRight)
	expectFocus("right wraps", panelLibraries)
	press(m, tea.KeyLeft)
	expectFocus("left wraps", panelDetails)

	// In Details, tab scrolls like down and shift+tab like up.
	m.detailsSrc = strings.Repeat("line\n\n", 200) // separate paragraphs, so it wraps to many lines
	m.renderDetails()
	m.View()
	press(m, tea.KeyTab)
	if m.details.Offset() != 1 {
		t.Fatalf("tab did not scroll details: offset %d", m.details.Offset())
	}
	press(m, tea.KeyShiftTab)
	if m.details.Offset() != 0 {
		t.Fatalf("shift+tab did not scroll details back: offset %d", m.details.Offset())
	}

	if m.libSearch.Value() != "" || m.search.Value() != "" {
		t.Fatalf("search values changed: %q %q", m.libSearch.Value(), m.search.Value())
	}
}
