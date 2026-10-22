package app

import (
	"fmt"
	"path/filepath"

	"github.com/maxx/skill-selector-app/internal/files"
	"github.com/maxx/skill-selector-app/internal/skills"
)

// Library is one skill library: a directory of skills.
type Library struct {
	// ID is the library's directory path, unique across the workspace.
	ID string
	// Label is the directory name, for display.
	Label string
}

// Skill is one skill inside the current library.
type Skill struct {
	// ID is the skill's directory path, unique across all libraries.
	ID string
	// Label is the directory name, for display.
	Label string
}

// App is the central coordinator for business operations. It holds the
// library list, loaded once, and the skills of the currently selected
// library. Whether a skill is exposed is always read from disk so the
// display reflects links made or removed outside the program.
type App struct {
	repo      skills.Repo
	libraries []Library
	library   string // ID of the current library, "" when none
	skills    []Skill
}

// New loads the libraries of the workspace. It fails if the libraries
// directory cannot be listed; a workspace with no libraries installed is
// valid and yields an App with empty lists.
func New(repo skills.Repo) (*App, error) {
	dirs, err := files.ListDirs(repo.LibrariesDir())
	if err != nil {
		return nil, fmt.Errorf("listing skill libraries: %w", err)
	}
	a := &App{repo: repo}
	for _, d := range dirs {
		a.libraries = append(a.libraries, Library{ID: d, Label: filepath.Base(d)})
	}
	return a, nil
}

// Libraries returns the known skill libraries in display order.
func (a *App) Libraries() []Library {
	return append([]Library(nil), a.libraries...)
}

// SelectLibrary makes the library with the given ID current and loads its
// skills. Selecting "" or an unknown ID clears the skill list. A library
// without a skills directory yields an empty list and an error describing
// why, so the caller can report it while the UI stays consistent.
func (a *App) SelectLibrary(id string) error {
	a.library = id
	a.skills = nil
	if id == "" {
		return nil
	}
	dirs, err := files.ListDirs(a.repo.LibrarySkillsDir(id))
	if err != nil {
		return fmt.Errorf("listing skills of %s: %w", filepath.Base(id), err)
	}
	for _, d := range dirs {
		a.skills = append(a.skills, Skill{ID: d, Label: filepath.Base(d)})
	}
	return nil
}

// Skills returns the skills of the current library in display order.
func (a *App) Skills() []Skill {
	return append([]Skill(nil), a.skills...)
}

// IsSelected reports whether the skill with the given ID is exposed to the
// agent, that is, linked into the exposed skills directory. A filesystem
// error is treated as not selected.
func (a *App) IsSelected(id string) bool {
	linked, err := files.IsLinked(a.repo.ExposedDir(), id)
	return err == nil && linked
}

// ToggleSkill exposes the skill with the given ID if it is not exposed, and
// hides it if it is, by adding or removing its link in the exposed skills
// directory. An error is returned when the link cannot be changed, for
// example when another library's skill of the same name is already exposed.
func (a *App) ToggleSkill(id string) error {
	exposed := a.repo.ExposedDir()
	linked, err := files.IsLinked(exposed, id)
	if err != nil {
		return fmt.Errorf("checking %s: %w", filepath.Base(id), err)
	}
	if linked {
		if err := files.Unlink(exposed, id); err != nil {
			return fmt.Errorf("hiding %s: %w", filepath.Base(id), err)
		}
		return nil
	}
	if err := files.Link(exposed, id); err != nil {
		return fmt.Errorf("exposing %s: %w", filepath.Base(id), err)
	}
	return nil
}

// SkillFile returns the text of the main file of the skill with the given
// ID, for the details panel. When the file cannot be read the returned text
// says so, so the panel always has something to show.
func (a *App) SkillFile(id string) string {
	if id == "" {
		return ""
	}
	text, err := files.Read(a.repo.SkillFile(id))
	if err != nil {
		return fmt.Sprintf("Could not read %s:\n\n%v", skills.SkillFileName, err)
	}
	return text
}
