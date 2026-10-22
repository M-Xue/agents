// Command skill-selector is the entry point for the TUI.
// It wires together the app, skills, files and ui packages.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxx/skill-selector-app/internal/app"
	"github.com/maxx/skill-selector-app/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.New(app.New()), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running skill-selector: %v\n", err)
		os.Exit(1)
	}
}
