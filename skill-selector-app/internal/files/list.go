package files

import (
	"os"
	"path/filepath"
	"sort"
)

// ListDirs returns the full path of every directory directly beneath path,
// sorted by name. Symlinks that resolve to a directory are included; files
// and dangling links are not.
func ListDirs(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(entries))
	for _, e := range entries {
		full := filepath.Join(path, e.Name())
		if e.IsDir() {
			dirs = append(dirs, full)
			continue
		}
		if e.Type()&os.ModeSymlink == 0 {
			continue
		}
		// Follow the link; Stat fails for a dangling link, which we skip.
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			dirs = append(dirs, full)
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}
