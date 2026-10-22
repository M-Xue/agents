// Command skill-selector is the entry point for the TUI.
// It locates the agents workspace, constructs the app and starts the UI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxx/skill-selector-app/internal/app"
	"github.com/maxx/skill-selector-app/internal/skills"
	"github.com/maxx/skill-selector-app/internal/ui"
)

func main() {
	root := flag.String("root", "", "agents workspace root (default: found by walking up from the current directory)")
	flag.Parse()

	repo, err := locate(*root)
	if err != nil {
		fail(err)
	}
	application, err := app.New(repo)
	if err != nil {
		fail(err)
	}
	if _, err := tea.NewProgram(ui.New(application), tea.WithAltScreen()).Run(); err != nil {
		fail(fmt.Errorf("running skill-selector: %w", err))
	}
}

// locate returns the workspace at root, or searches upward from the current
// directory when root is empty.
func locate(root string) (skills.Repo, error) {
	if root != "" {
		abs, err := filepath.Abs(root)
		if err != nil {
			return skills.Repo{}, err
		}
		return skills.Repo{Root: abs}, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return skills.Repo{}, err
	}
	return skills.FindRoot(cwd)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "skill-selector: %v\n", err)
	os.Exit(1)
}
