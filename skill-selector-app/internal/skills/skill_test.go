package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"skills", "skill-selector-app", "skill-libraries/lib/skills/a"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// macOS temp dirs live under a symlinked /var; compare resolved paths.
	want, _ := filepath.EvalSymlinks(root)

	for _, start := range []string{root, filepath.Join(root, "skill-libraries/lib/skills/a")} {
		repo, err := FindRoot(start)
		if err != nil {
			t.Fatalf("FindRoot(%s): %v", start, err)
		}
		got, _ := filepath.EvalSymlinks(repo.Root)
		if got != want {
			t.Fatalf("FindRoot(%s) = %s, want %s", start, got, want)
		}
	}
	if _, err := FindRoot(t.TempDir()); err == nil {
		t.Fatal("expected error outside a workspace")
	}
}

func TestPaths(t *testing.T) {
	r := Repo{Root: "/w"}
	cases := map[string]string{
		r.ExposedDir():   "/w/skills",
		r.LibrariesDir(): "/w/skill-libraries",
		r.LibrarySkillsDir("/w/skill-libraries/lib"):   "/w/skill-libraries/lib/skills",
		r.SkillFile("/w/skill-libraries/lib/skills/a"): "/w/skill-libraries/lib/skills/a/SKILL.md",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %s want %s", got, want)
		}
	}
}
