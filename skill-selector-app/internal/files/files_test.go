package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SKILL.md")
	if err := os.WriteFile(path, []byte("# hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "# hi\n" {
		t.Fatalf("Read = %q", got)
	}
	if _, err := Read(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestListDirs(t *testing.T) {
	root := t.TempDir()
	mkdir(t, root, "b")
	mkdir(t, root, "a")
	writeFile(t, root, "file.txt")
	other := t.TempDir()
	mkdir(t, other, "linked")
	if err := os.Symlink(filepath.Join(other, "linked"), filepath.Join(root, "c")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(other, "gone"), filepath.Join(root, "dangling")); err != nil {
		t.Fatal(err)
	}

	got, err := ListDirs(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(root, "a"),
		filepath.Join(root, "b"),
		filepath.Join(root, "c"),
	}
	if len(got) != len(want) {
		t.Fatalf("ListDirs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ListDirs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	if _, err := ListDirs(filepath.Join(root, "nope")); err == nil {
		t.Fatal("expected error for missing dir")
	}
}

func TestLinkRoundTrip(t *testing.T) {
	dir := t.TempDir()
	target := mkdir(t, t.TempDir(), "code-review")

	if ok, _ := IsLinked(dir, target); ok {
		t.Fatal("IsLinked true before Link")
	}
	if err := Link(dir, target); err != nil {
		t.Fatal(err)
	}
	ok, err := IsLinked(dir, target)
	if err != nil || !ok {
		t.Fatalf("IsLinked after Link = %v, %v", ok, err)
	}
	// The link is named after the target and stores an absolute path.
	dest, err := os.Readlink(filepath.Join(dir, "code-review"))
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(dest) {
		t.Fatalf("link stores relative path %q", dest)
	}
	// Linking twice is an error rather than a silent overwrite.
	if err := Link(dir, target); err == nil {
		t.Fatal("second Link should fail")
	}
	if err := Unlink(dir, target); err != nil {
		t.Fatal(err)
	}
	if ok, _ := IsLinked(dir, target); ok {
		t.Fatal("IsLinked true after Unlink")
	}
	if _, err := os.Lstat(filepath.Join(dir, "code-review")); !os.IsNotExist(err) {
		t.Fatalf("link still present: %v", err)
	}
	// Unlinking something that is not there is a no-op.
	if err := Unlink(dir, target); err != nil {
		t.Fatalf("Unlink of absent link = %v", err)
	}
}

func TestLinkRelativeTarget(t *testing.T) {
	dir := t.TempDir()
	base := t.TempDir()
	mkdir(t, base, "rel")
	t.Chdir(base)

	if err := Link(dir, "rel"); err != nil {
		t.Fatal(err)
	}
	// Still detected when asked with the absolute spelling.
	if ok, err := IsLinked(dir, filepath.Join(base, "rel")); err != nil || !ok {
		t.Fatalf("IsLinked(abs) = %v, %v", ok, err)
	}
	if err := Unlink(dir, "rel"); err != nil {
		t.Fatal(err)
	}
}

func TestIsLinkedDistinguishes(t *testing.T) {
	dir := t.TempDir()
	target := mkdir(t, t.TempDir(), "skill")
	elsewhere := mkdir(t, t.TempDir(), "skill")

	// A real directory with the same name is not a link.
	mkdir(t, dir, "skill")
	if ok, err := IsLinked(dir, target); err != nil || ok {
		t.Fatalf("real dir reported as linked: %v, %v", ok, err)
	}
	// And Unlink refuses to delete it.
	if err := Unlink(dir, target); err == nil {
		t.Fatal("Unlink removed a real directory")
	}
	if _, err := os.Stat(filepath.Join(dir, "skill")); err != nil {
		t.Fatalf("real dir gone: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "skill")); err != nil {
		t.Fatal(err)
	}

	// A link with the right name to a different folder is not a link to target.
	if err := Link(dir, elsewhere); err != nil {
		t.Fatal(err)
	}
	if ok, err := IsLinked(dir, target); err != nil || ok {
		t.Fatalf("link to other folder reported as linked: %v, %v", ok, err)
	}
	if err := Unlink(dir, target); err == nil {
		t.Fatal("Unlink removed a link to a different folder")
	}
	if ok, _ := IsLinked(dir, elsewhere); !ok {
		t.Fatal("other link was removed")
	}
}

func mkdir(t *testing.T, parent, name string) string {
	t.Helper()
	path := filepath.Join(parent, name)
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFile(t *testing.T, parent, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(parent, name), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBlocker(t *testing.T) {
	dir := t.TempDir()
	lib := t.TempDir()
	target := mkdir(t, mkdir(t, lib, "a"), "code-review")
	other := mkdir(t, mkdir(t, lib, "b"), "code-review")

	// Nothing there: free.
	if got, err := Blocker(dir, target); err != nil || got != "" {
		t.Fatalf("Blocker on empty dir = %q, %v", got, err)
	}
	// Our own link: still free.
	if err := Link(dir, target); err != nil {
		t.Fatal(err)
	}
	if got, err := Blocker(dir, target); err != nil || got != "" {
		t.Fatalf("Blocker with own link = %q, %v", got, err)
	}
	// From the other skill's point of view the name is held by a link to
	// target, and Blocker names that destination.
	got, err := Blocker(dir, other)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := canonical(target); got != want && got != target {
		t.Fatalf("Blocker(other) = %q, want %q", got, target)
	}
	// A real directory of the same name blocks with its own path.
	if err := Unlink(dir, target); err != nil {
		t.Fatal(err)
	}
	real := mkdir(t, dir, "code-review")
	if got, err := Blocker(dir, target); err != nil || got != real {
		t.Fatalf("Blocker with real dir = %q, %v, want %q", got, err, real)
	}
}
