// Package theme centralizes the colors shared by the app's chrome: panel
// borders, title tabs and de-emphasized text.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	// BorderActive outlines the box that owns the keyboard: the focused panel.
	BorderActive = lipgloss.Color("111")
	// BorderInactive is the muted slate blue for every other section border.
	BorderInactive = lipgloss.Color("60")
	// TextInactive dims de-emphasized interactive text. It shares the
	// inactive-border color so idle text and idle chrome read as one layer.
	TextInactive = BorderInactive
	// TitleFg/TitleBg render section titles as a tab matching the active
	// border color, with dark text.
	TitleFg = lipgloss.Color("235")
	TitleBg = lipgloss.Color("111")
	// Error colours a failure message shown in place of the help footer.
	Error = lipgloss.Color("203")
)
