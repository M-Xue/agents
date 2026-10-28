package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// IsLinked reports whether dir contains a symlink to target. It is false when
// nothing with that name exists in dir, when the entry is a real file or
// directory rather than a link, and when the link points somewhere else.
func IsLinked(dir, target string) (bool, error) {
	link := linkPath(dir, target)
	info, err := os.Lstat(link)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}
	return pointsAt(link, target)
}

// Link creates a symlink to target inside dir. The link stores target's
// absolute path so it keeps working regardless of the working directory.
// It fails if dir already has an entry with that name.
func Link(dir, target string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	return os.Symlink(abs, linkPath(dir, target))
}

// Unlink removes the symlink to target from dir. It only ever removes a
// symlink that points at target: a real file or directory, or a link to
// something else, is left alone and reported as an error. Removing a link
// that does not exist is not an error.
func Unlink(dir, target string) error {
	link := linkPath(dir, target)
	info, err := os.Lstat(link)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("unlink %s: not a symlink", link)
	}
	ok, err := pointsAt(link, target)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unlink %s: does not point at %s", link, target)
	}
	return os.Remove(link)
}

// Blocker reports what stops target from being linked into dir. It returns
// "" when the link name is free or already holds a link to target. When a
// symlink to somewhere else holds the name, it returns that link's
// destination as an absolute path; when a real file or directory holds the
// name, it returns that entry's own path.
func Blocker(dir, target string) (string, error) {
	link := linkPath(dir, target)
	info, err := os.Lstat(link)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return link, nil
	}
	ok, err := pointsAt(link, target)
	if err != nil {
		return "", err
	}
	if ok {
		return "", nil
	}
	dest, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(link), dest)
	}
	return filepath.Clean(dest), nil
}

// linkPath is where the link for target lives inside dir.
func linkPath(dir, target string) string {
	return filepath.Join(dir, filepath.Base(target))
}

// pointsAt reports whether the symlink at link resolves to target.
func pointsAt(link, target string) (bool, error) {
	dest, err := os.Readlink(link)
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(link), dest)
	}
	want, err := canonical(target)
	if err != nil {
		return false, err
	}
	got, err := canonical(dest)
	if err != nil {
		return false, err
	}
	return got == want, nil
}

// canonical returns an absolute, cleaned path with any symlinks along the
// way resolved, so two spellings of the same folder compare equal. If the
// path does not exist it falls back to the absolute cleaned form.
func canonical(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved, nil
	}
	return abs, nil
}
