package app

import "fmt"

// dummySkillCount is how many placeholder skills dummySkills returns.
const dummySkillCount = 80

// App is the central coordinator for business operations.
//
// The skill list and selection below are THROWAWAY placeholders so the UI has
// state to render against. They will be replaced when the skills and files
// packages are implemented.
type App struct {
	skills   []Skill
	selected map[string]bool
}

// Skill is a placeholder description of a skill until the skills package
// provides the real type.
type Skill struct {
	ID    string
	Label string
}

// New constructs an App with its dependencies.
func New() *App {
	return &App{
		skills:   dummySkills(),
		selected: map[string]bool{},
	}
}

// Skills returns the known skills in display order.
func (a *App) Skills() []Skill {
	return append([]Skill(nil), a.skills...)
}

// IsSelected reports whether the skill with the given ID is selected.
func (a *App) IsSelected(id string) bool { return a.selected[id] }

// ToggleSkill flips the selection state of the skill with the given ID.
func (a *App) ToggleSkill(id string) {
	if a.selected[id] {
		delete(a.selected, id)
		return
	}
	a.selected[id] = true
}

// dummySkills is placeholder data; delete once skills are discovered on disk.
// It returns dummySkillCount items so the checklist has enough rows to scroll.
func dummySkills() []Skill {
	names := []string{
		"code-review", "simplify", "security-review", "init", "dataviz",
		"artifact-design", "artifact-diagramming", "pdf", "docx", "xlsx",
		"pptx", "skill-creator", "ast-grep", "ci-results",
	}
	skills := make([]Skill, 0, dummySkillCount)
	for i := 0; i < dummySkillCount; i++ {
		name := fmt.Sprintf("%s-%02d", names[i%len(names)], i+1)
		skills = append(skills, Skill{ID: name, Label: name})
	}
	return skills
}
