// Command skill-selector-app is the entry point for the TUI; build it with
// go build in this directory, or install it with scripts/install.sh.
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

// defaultRoot is the workspace to fall back on when none is found above the
// current directory. scripts/install.sh sets it at build time with
// -ldflags "-X main.defaultRoot=<path>" so the installed command works from
// anywhere; a plain go build leaves it empty.
var defaultRoot string

func main() {
	root := flag.String("root", "", "agents workspace root (default: found by walking up from the current directory, then the path compiled in by the installer)")
	flag.Parse()

	repo, err := locate(*root)
	if err != nil {
		fail(err)
	}
	application, err := app.New(repo)
	if err != nil {
		fail(err)
	}
	if _, err := tea.NewProgram(ui.New(application), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fail(fmt.Errorf("running skill-selector: %w", err))
	}
}

// locate returns the workspace to operate on, in order of preference: the
// explicit root, the workspace found above the current directory, then the
// one compiled in by the installer.
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
	repo, err := skills.FindRoot(cwd)
	if err == nil {
		return repo, nil
	}
	if defaultRoot != "" {
		return skills.Repo{Root: defaultRoot}, nil
	}
	return skills.Repo{}, err
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "skill-selector: %v\n", err)
	os.Exit(1)
}
