package skills

import (
	"fmt"
	"path/filepath"

	"github.com/maxx/skill-selector-app/internal/files"
)

const (
	// SkillFileName is the file every skill directory must contain.
	SkillFileName = "SKILL.md"

	// exposedDirName is the workspace directory the agent reads skills
	// from. Selected skills are symlinked into it.
	exposedDirName = "skills"
	// librariesDirName holds one cloned skill library per directory.
	librariesDirName = "skill-libraries"
	// librarySkillsDirName is the flat directory of skills inside a library.
	librarySkillsDirName = "skills"
	// appDirName is this program's source directory. Together with
	// exposedDirName it marks the workspace root; librariesDirName cannot be
	// a marker because it is gitignored and absent on a fresh clone.
	appDirName = "skill-selector-app"
)

// Repo is the agents workspace rooted at Root. All paths it returns are
// absolute when Root is.
type Repo struct {
	Root string
}

// FindRoot walks up from start until it reaches the workspace root, the
// directory containing both the exposed skills directory and this app's
// source directory.
func FindRoot(start string) (Repo, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return Repo{}, err
	}
	for {
		if isRoot(dir) {
			return Repo{Root: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Repo{}, fmt.Errorf("no agents workspace found at or above %s", start)
		}
		dir = parent
	}
}

// ExposedDir is the directory the agent reads skills from. Selecting a skill
// links it in here; deselecting removes the link.
func (r Repo) ExposedDir() string {
	return filepath.Join(r.Root, exposedDirName)
}

// LibrariesDir holds the skill libraries, one per subdirectory.
func (r Repo) LibrariesDir() string {
	return filepath.Join(r.Root, librariesDirName)
}

// LibrarySkillsDir is the directory of skills inside the library at
// libraryDir, one skill per subdirectory.
func (r Repo) LibrarySkillsDir(libraryDir string) string {
	return filepath.Join(libraryDir, librarySkillsDirName)
}

// SkillFile is the path of the main file of the skill at skillDir.
func (r Repo) SkillFile(skillDir string) string {
	return filepath.Join(skillDir, SkillFileName)
}

// isRoot reports whether dir contains the two marker directories.
func isRoot(dir string) bool {
	children, err := files.ListDirs(dir)
	if err != nil {
		return false
	}
	found := map[string]bool{}
	for _, c := range children {
		found[filepath.Base(c)] = true
	}
	return found[exposedDirName] && found[appDirName]
}
