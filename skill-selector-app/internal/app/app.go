package app

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/maxx/skill-selector-app/internal/files"
	"github.com/maxx/skill-selector-app/internal/skills"
)

// The built-in libraries are views over the installed ones. Their IDs are
// not paths, so they can never collide with a real library's directory.
const (
	// AllLibraryID combines the skills of every installed library.
	AllLibraryID = "*all"
	// SelectedLibraryID lists only the skills currently exposed to the
	// agent, from whichever library they come.
	SelectedLibraryID = "*selected"
)

// Labels of the built-in libraries. The suffix marks them as provided by the
// program rather than cloned from a repository.
const (
	allLibraryLabel      = "All [built-in]"
	selectedLibraryLabel = "Currently selected [built-in]"
)

// IsBuiltIn reports whether id names one of the built-in libraries rather
// than an installed one.
func IsBuiltIn(id string) bool {
	return id == AllLibraryID || id == SelectedLibraryID
}

// Library is one skill library: a directory of skills, or a built-in view
// over all of them.
type Library struct {
	// ID is the library's directory path, unique across the workspace, or
	// one of the built-in IDs.
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
	// Library is the name of the library the skill belongs to, for display
	// when skills from several libraries are listed together.
	Library string
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
// valid and yields an App with only the built-in libraries. Those are
// always first in the list: All, then Currently selected.
func New(repo skills.Repo) (*App, error) {
	dirs, err := files.ListDirs(repo.LibrariesDir())
	if err != nil {
		return nil, fmt.Errorf("listing skill libraries: %w", err)
	}
	a := &App{repo: repo}
	a.libraries = append(a.libraries,
		Library{ID: AllLibraryID, Label: allLibraryLabel},
		Library{ID: SelectedLibraryID, Label: selectedLibraryLabel},
	)
	for _, d := range dirs {
		a.libraries = append(a.libraries, Library{ID: d, Label: filepath.Base(d)})
	}
	return a, nil
}

// Libraries returns the known skill libraries in display order: the
// built-in libraries, then the installed libraries by name.
func (a *App) Libraries() []Library {
	return append([]Library(nil), a.libraries...)
}

// SelectLibrary makes the library with the given ID current and loads its
// skills. Selecting "" or an unknown ID clears the skill list. A library
// without a skills directory yields an empty list and an error describing
// why, so the caller can report it while the UI stays consistent.
//
// Selecting a built-in library loads the skills of every installed library,
// sorted by skill name and then library name so skills that share a name
// sit together. A library whose skills cannot be listed is skipped and the
// skills of the others are still loaded; the failures are returned as one
// error. For SelectedLibraryID, Skills then narrows that list to the skills
// exposed at the time it is called.
func (a *App) SelectLibrary(id string) error {
	a.library = id
	a.skills = nil
	if id == "" {
		return nil
	}
	if IsBuiltIn(id) {
		return a.loadAll()
	}
	skills, err := a.loadLibrary(id)
	a.skills = skills
	return err
}

// loadLibrary lists the skills of the library with the given ID.
func (a *App) loadLibrary(id string) ([]Skill, error) {
	dirs, err := files.ListDirs(a.repo.LibrarySkillsDir(id))
	if err != nil {
		return nil, fmt.Errorf("listing skills of %s: %w", filepath.Base(id), err)
	}
	library := filepath.Base(id)
	skills := make([]Skill, 0, len(dirs))
	for _, d := range dirs {
		skills = append(skills, Skill{ID: d, Label: filepath.Base(d), Library: library})
	}
	return skills, nil
}

// loadAll fills the skill list with the skills of every installed library.
func (a *App) loadAll() error {
	var failures []string
	for _, lib := range a.libraries {
		if IsBuiltIn(lib.ID) {
			continue
		}
		skills, err := a.loadLibrary(lib.ID)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		a.skills = append(a.skills, skills...)
	}
	slices.SortStableFunc(a.skills, func(x, y Skill) int {
		if c := strings.Compare(x.Label, y.Label); c != 0 {
			return c
		}
		return strings.Compare(x.Library, y.Library)
	})
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

// Skills returns the skills of the current library in display order. While
// the Currently selected library is current, only the skills exposed right
// now are returned: the check is made against the disk on every call, so a
// skill that has just been hidden disappears from the next call and one
// linked outside the program appears.
func (a *App) Skills() []Skill {
	if a.library != SelectedLibraryID {
		return append([]Skill(nil), a.skills...)
	}
	var selected []Skill
	for _, s := range a.skills {
		if a.IsSelected(s.ID) {
			selected = append(selected, s)
		}
	}
	return selected
}

// IsSelected reports whether the skill with the given ID is exposed to the
// agent, that is, linked into the exposed skills directory. A filesystem
// error is treated as not selected.
func (a *App) IsSelected(id string) bool {
	linked, err := files.IsLinked(a.repo.ExposedDir(), id)
	return err == nil && linked
}

// ShadowedBy reports what already holds the skill's name in the exposed
// directory when it is neither free nor this skill's own link: the name of
// the library whose same-named skill is exposed, or, for an entry that is not
// a library skill, its path. It returns "" when the skill can be exposed or
// already is. Like IsSelected it reads from disk on every call, and a
// filesystem error is treated as not shadowed.
func (a *App) ShadowedBy(id string) string {
	by, _, err := a.shadow(id)
	if err != nil {
		return ""
	}
	return by
}

// shadow is ShadowedBy with the detail of whether the blocker is a skill of
// an installed library, so callers can phrase messages accordingly.
func (a *App) shadow(id string) (by string, library bool, err error) {
	blocker, err := files.Blocker(a.repo.ExposedDir(), id)
	if err != nil || blocker == "" {
		return "", false, err
	}
	if lib := a.libraryOf(blocker); lib != "" {
		return lib, true, nil
	}
	return blocker, false, nil
}

// libraryOf returns the name of the installed library whose skills directory
// contains path, or "" when path is not a skill of an installed library.
func (a *App) libraryOf(path string) string {
	for _, lib := range a.libraries {
		if IsBuiltIn(lib.ID) {
			continue
		}
		rel, err := filepath.Rel(a.repo.LibrarySkillsDir(lib.ID), path)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}
		return lib.Label
	}
	return ""
}

// SetSelected exposes the skill with the given ID when selected is true and
// hides it when false, by adding or removing its link in the exposed skills
// directory. It is a no-op when the skill is already in the requested state.
// An error is returned when the link cannot be changed. When the skill's name
// is already taken, the error says by what: another library's skill of the
// same name, or a foreign entry in the exposed directory.
func (a *App) SetSelected(id string, selected bool) error {
	exposed := a.repo.ExposedDir()
	name := filepath.Base(id)
	linked, err := files.IsLinked(exposed, id)
	if err != nil {
		return fmt.Errorf("checking %s: %w", name, err)
	}
	switch {
	case selected && !linked:
		by, library, err := a.shadow(id)
		switch {
		case err != nil:
			return fmt.Errorf("checking %s: %w", name, err)
		case by != "" && library:
			return fmt.Errorf("exposing %s: already exposed from %s", name, by)
		case by != "":
			return fmt.Errorf("exposing %s: name already taken by %s", name, by)
		}
		if err := files.Link(exposed, id); err != nil {
			return fmt.Errorf("exposing %s: %w", name, err)
		}
	case !selected && linked:
		if err := files.Unlink(exposed, id); err != nil {
			return fmt.Errorf("hiding %s: %w", name, err)
		}
	}
	return nil
}

// ToggleSkill exposes the skill with the given ID if it is not exposed, and
// hides it if it is.
func (a *App) ToggleSkill(id string) error {
	return a.SetSelected(id, !a.IsSelected(id))
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
