package ui

import "github.com/charmbracelet/lipgloss"

var (
	white = lipgloss.Color("15")

	// panelStyle is the border drawn around every panel.
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(white)

	// focusedPanelStyle marks the panel that currently has focus.
	focusedPanelStyle = panelStyle.
				Border(lipgloss.ThickBorder())

	// titleStyle renders the heading inside each panel.
	titleStyle = lipgloss.NewStyle().Bold(true)
)
