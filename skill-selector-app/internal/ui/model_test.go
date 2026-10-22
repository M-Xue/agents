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
		dir := filepath.Join(root, "third-party-skills", lib, "skills", skill)
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
	press(m, tea.KeyTab)  // focus skills
	press(m, tea.KeyDown) // apple -> banana
	if v := view(m); !strings.Contains(v, "BANANA SKILL BODY") || strings.Contains(v, "APPLE SKILL BODY") {
		t.Fatal("details did not follow the skill cursor")
	}
}

func TestEnterTogglesSymlink(t *testing.T) {
	m, repo := newModel(t)
	press(m, tea.KeyTab) // focus skills, cursor on apple
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
	press(m, tea.KeyTab)
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
	press(m, tea.KeyTab)
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
