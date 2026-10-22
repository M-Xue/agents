package ui

// Model is the root Bubble Tea model for the application.
type Model struct{}

// New returns a Model ready to be passed to tea.NewProgram.
func New() Model {
	return Model{}
}
