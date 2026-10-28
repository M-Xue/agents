package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maxx/skill-selector-app/internal/skills"
)

// newWorkspace builds a workspace with two libraries. "alpha" has skills
// one and two; "beta" has skill one (a name clash with alpha) and no
// SKILL.md in it.
func newWorkspace(t *testing.T) skills.Repo {
	t.Helper()
	root := t.TempDir()
	mk := func(parts ...string) string {
		p := filepath.Join(append([]string{root}, parts...)...)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	mk("skills")
	one := mk("skill-libraries", "alpha", "skills", "one")
	mk("skill-libraries", "alpha", "skills", "two")
	mk("skill-libraries", "beta", "skills", "one")
	mk("skill-libraries", "empty")
	if err := os.WriteFile(filepath.Join(one, "SKILL.md"), []byte("# one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return skills.Repo{Root: root}
}

// lib returns the ID of the library with the given label.
func lib(t *testing.T, a *App, label string) string {
	t.Helper()
	for _, l := range a.Libraries() {
		if l.Label == label {
			return l.ID
		}
	}
	t.Fatalf("no library labelled %q", label)
	return ""
}

func labels[T interface{ Library | Skill }](xs []T) string {
	parts := make([]string, 0, len(xs))
	for _, x := range xs {
		switch v := any(x).(type) {
		case Library:
			parts = append(parts, v.Label)
		case Skill:
			parts = append(parts, v.Label)
		}
	}
	return strings.Join(parts, ",")
}

func TestLibrariesAndSkills(t *testing.T) {
	a, err := New(newWorkspace(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := labels(a.Libraries()); got != "All,Currently selected,alpha,beta,empty" {
		t.Fatalf("Libraries = %s", got)
	}
	if a.Libraries()[0].ID != AllLibraryID || a.Libraries()[1].ID != SelectedLibraryID {
		t.Fatalf("built-in IDs = %q, %q", a.Libraries()[0].ID, a.Libraries()[1].ID)
	}
	if len(a.Skills()) != 0 {
		t.Fatal("skills before any library is selected")
	}
	if err := a.SelectLibrary(lib(t, a, "alpha")); err != nil {
		t.Fatal(err)
	}
	if got := labels(a.Skills()); got != "one,two" {
		t.Fatalf("alpha skills = %s", got)
	}
	for _, sk := range a.Skills() {
		if sk.Library != "alpha" {
			t.Fatalf("skill %s has Library %q, want alpha", sk.Label, sk.Library)
		}
	}
	if err := a.SelectLibrary(lib(t, a, "empty")); err == nil {
		t.Fatal("expected error for library without skills dir")
	}
	if len(a.Skills()) != 0 {
		t.Fatal("skills not cleared after failed select")
	}
	if err := a.SelectLibrary(""); err != nil || len(a.Skills()) != 0 {
		t.Fatalf("SelectLibrary(\"\") = %v, %d skills", err, len(a.Skills()))
	}
}

func TestToggleAndClash(t *testing.T) {
	repo := newWorkspace(t)
	a, err := New(repo)
	if err != nil {
		t.Fatal(err)
	}
	_ = a.SelectLibrary(lib(t, a, "alpha"))
	one := a.Skills()[0]

	if a.IsSelected(one.ID) {
		t.Fatal("selected before toggle")
	}
	if err := a.ToggleSkill(one.ID); err != nil {
		t.Fatal(err)
	}
	if !a.IsSelected(one.ID) {
		t.Fatal("not selected after toggle")
	}
	link := filepath.Join(repo.ExposedDir(), "one")
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s: %v", link, err)
	}

	// beta/one clashes with the exposed alpha/one: it is not selected, and
	// toggling it must fail rather than replace alpha's link.
	_ = a.SelectLibrary(lib(t, a, "beta"))
	betaOne := a.Skills()[0]
	if a.IsSelected(betaOne.ID) {
		t.Fatal("beta/one reported selected via alpha's link")
	}
	if by := a.ShadowedBy(betaOne.ID); by != "alpha" {
		t.Fatalf("ShadowedBy(beta/one) = %q, want alpha", by)
	}
	if by := a.ShadowedBy(one.ID); by != "" {
		t.Fatalf("ShadowedBy(alpha/one) = %q, want none for the exposed skill", by)
	}
	err = a.ToggleSkill(betaOne.ID)
	if err == nil {
		t.Fatal("toggling clashing skill succeeded")
	}
	if want := "exposing one: already exposed from alpha"; err.Error() != want {
		t.Fatalf("clash error = %q, want %q", err, want)
	}
	if !a.IsSelected(one.ID) {
		t.Fatal("alpha/one link was disturbed")
	}

	if err := a.ToggleSkill(one.ID); err != nil {
		t.Fatal(err)
	}
	if a.IsSelected(one.ID) {
		t.Fatal("still selected after second toggle")
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("link still present: %v", err)
	}
	// With alpha's link gone, beta/one is no longer shadowed.
	if by := a.ShadowedBy(betaOne.ID); by != "" {
		t.Fatalf("ShadowedBy(beta/one) after unlink = %q, want none", by)
	}
}

func TestForeignEntryShadows(t *testing.T) {
	repo := newWorkspace(t)
	a, err := New(repo)
	if err != nil {
		t.Fatal(err)
	}
	_ = a.SelectLibrary(lib(t, a, "alpha"))
	one := a.Skills()[0]
	taken := filepath.Join(repo.ExposedDir(), "one")
	if err := os.Mkdir(taken, 0o755); err != nil {
		t.Fatal(err)
	}
	if by := a.ShadowedBy(one.ID); by != taken {
		t.Fatalf("ShadowedBy with a real dir = %q, want %q", by, taken)
	}
	err = a.ToggleSkill(one.ID)
	if err == nil || !strings.Contains(err.Error(), "name already taken by "+taken) {
		t.Fatalf("error for foreign entry = %v", err)
	}
}

func TestSkillFile(t *testing.T) {
	a, err := New(newWorkspace(t))
	if err != nil {
		t.Fatal(err)
	}
	_ = a.SelectLibrary(lib(t, a, "alpha"))
	if got := a.SkillFile(a.Skills()[0].ID); got != "# one\n" {
		t.Fatalf("SkillFile(one) = %q", got)
	}
	if got := a.SkillFile(a.Skills()[1].ID); !strings.Contains(got, "Could not read") {
		t.Fatalf("SkillFile(two) = %q", got)
	}
	if got := a.SkillFile(""); got != "" {
		t.Fatalf("SkillFile(\"\") = %q", got)
	}
}

func TestSetSelectedIsIdempotent(t *testing.T) {
	a, err := New(newWorkspace(t))
	if err != nil {
		t.Fatal(err)
	}
	_ = a.SelectLibrary(lib(t, a, "alpha"))
	id := a.Skills()[0].ID
	for i := 0; i < 2; i++ {
		if err := a.SetSelected(id, true); err != nil {
			t.Fatalf("SetSelected(true) #%d: %v", i, err)
		}
		if !a.IsSelected(id) {
			t.Fatal("not selected")
		}
	}
	for i := 0; i < 2; i++ {
		if err := a.SetSelected(id, false); err != nil {
			t.Fatalf("SetSelected(false) #%d: %v", i, err)
		}
		if a.IsSelected(id) {
			t.Fatal("still selected")
		}
	}
}

func TestSelectAllCombinesLibraries(t *testing.T) {
	repo := newWorkspace(t)
	a, err := New(repo)
	if err != nil {
		t.Fatal(err)
	}
	// "empty" has no skills directory, so All reports it while still loading
	// the skills of alpha and beta.
	err = a.SelectLibrary(AllLibraryID)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("SelectLibrary(All) error = %v, want one naming the empty library", err)
	}
	got := make([]string, 0, len(a.Skills()))
	for _, sk := range a.Skills() {
		got = append(got, sk.Label+"/"+sk.Library)
	}
	// Sorted by skill name, then library, so the two "one" skills sit together.
	if want := "one/alpha,one/beta,two/alpha"; strings.Join(got, ",") != want {
		t.Fatalf("All skills = %s, want %s", strings.Join(got, ","), want)
	}
	// IDs are still the real skill directories, so toggling works unchanged.
	alphaOne := a.Skills()[0]
	if !strings.HasPrefix(alphaOne.ID, filepath.Join(repo.LibrariesDir(), "alpha")) {
		t.Fatalf("skill ID %q is not alpha's directory", alphaOne.ID)
	}
	if err := a.ToggleSkill(alphaOne.ID); err != nil {
		t.Fatal(err)
	}
	if !a.IsSelected(alphaOne.ID) || a.IsSelected(a.Skills()[1].ID) {
		t.Fatal("toggling alpha/one via All did not select exactly alpha/one")
	}

	// A workspace with no libraries at all still offers All, with no skills
	// and no error.
	empty := skills.Repo{Root: t.TempDir()}
	if err := os.MkdirAll(empty.LibrariesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := New(empty)
	if err != nil {
		t.Fatal(err)
	}
	if got := labels(b.Libraries()); got != "All,Currently selected" {
		t.Fatalf("Libraries = %s", got)
	}
	if err := b.SelectLibrary(AllLibraryID); err != nil || len(b.Skills()) != 0 {
		t.Fatalf("SelectLibrary(All) on empty workspace = %v, %d skills", err, len(b.Skills()))
	}
}

func TestSelectedLibraryTracksLinksLive(t *testing.T) {
	a, err := New(newWorkspace(t))
	if err != nil {
		t.Fatal(err)
	}
	// Nothing exposed: the view is empty but not an error about alpha/beta
	// (only "empty" is reported, as for All).
	err = a.SelectLibrary(SelectedLibraryID)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("SelectLibrary(Selected) error = %v", err)
	}
	if n := len(a.Skills()); n != 0 {
		t.Fatalf("%d skills selected before any link", n)
	}

	// Expose alpha/two and beta/one from another view; they appear in the
	// selected view without re-selecting it, each tagged with its library.
	_ = a.SelectLibrary(AllLibraryID)
	var alphaTwo, betaOne string
	for _, s := range a.Skills() {
		switch s.Label + "/" + s.Library {
		case "two/alpha":
			alphaTwo = s.ID
		case "one/beta":
			betaOne = s.ID
		}
	}
	for _, id := range []string{alphaTwo, betaOne} {
		if err := a.SetSelected(id, true); err != nil {
			t.Fatal(err)
		}
	}
	_ = a.SelectLibrary(SelectedLibraryID)
	got := make([]string, 0, 2)
	for _, s := range a.Skills() {
		got = append(got, s.Label+"/"+s.Library)
	}
	if want := "one/beta,two/alpha"; strings.Join(got, ",") != want {
		t.Fatalf("selected skills = %s, want %s", strings.Join(got, ","), want)
	}

	// Hiding one removes it from the very next Skills call.
	if err := a.SetSelected(betaOne, false); err != nil {
		t.Fatal(err)
	}
	if got := labels(a.Skills()); got != "two" {
		t.Fatalf("selected skills after hiding beta/one = %s", got)
	}
	if err := a.SetSelected(alphaTwo, false); err != nil {
		t.Fatal(err)
	}
	if n := len(a.Skills()); n != 0 {
		t.Fatalf("%d skills still listed after hiding all", n)
	}
}
