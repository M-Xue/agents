// Package panel renders a titled, sharp-cornered box around a section of the
// app, in the style of lazygit panes: the title sits as a highlighted tab in
// the top border and the content is padded one column from each side.
//
// The panel does not size itself. The caller computes width and height from
// the window and its own layout, sets them on the Model, and the panel fills
// exactly that many cells.
package panel

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/maxx/skill-selector-app/internal/ui/theme"
)

// Model is a titled box of a fixed size. Zero value is unusable; use New.
type Model struct {
	title  string
	width  int
	height int
	active bool
}

// New returns a panel with the given title and outer dimensions in cells.
func New(title string, width, height int) Model {
	return Model{title: title, width: width, height: height}
}

// SetSize updates the outer width and height in cells.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetTitle updates the label shown in the top border.
func (m *Model) SetTitle(title string) { m.title = title }

// Focus draws the border in the active color.
func (m *Model) Focus() { m.active = true }

// Blur draws the border in the muted color.
func (m *Model) Blur() { m.active = false }

// Focused reports whether the panel has focus.
func (m Model) Focused() bool { return m.active }

// Width returns the outer width in cells.
func (m Model) Width() int { return m.width }

// Height returns the outer height in cells.
func (m Model) Height() int { return m.height }

// InnerWidth returns the columns available to content: the width minus the
// two border columns and one column of padding on each side.
func (m Model) InnerWidth() int { return max(0, m.width-4) }

// InnerHeight returns the rows available to content: the height minus the
// top and bottom borders.
func (m Model) InnerHeight() int { return max(0, m.height-2) }

// View draws content inside the box. Content is clipped to the interior and
// missing lines are blank-filled so the box is always full height. Degenerate
// sizes return the content unboxed.
func (m Model) View(content string) string {
	if m.width < 4 || m.height < 2 {
		return content
	}
	borderColor := theme.BorderInactive
	if m.active {
		borderColor = theme.BorderActive
	}
	border := lipgloss.NewStyle().Foreground(borderColor)

	inner := m.width - 2 // columns between the vertical borders
	tab := tabWithBackground(m.title, borderColor)
	if lipgloss.Width(tab) > inner-1 {
		tab = lipgloss.NewStyle().MaxWidth(inner - 1).Render(tab)
	}
	dashes := max(0, inner-1-lipgloss.Width(tab))
	top := border.Render("┌─") + tab + border.Render(strings.Repeat("─", dashes)+"┐")
	bottom := border.Render("└" + strings.Repeat("─", inner) + "┘")

	side := border.Render("│")
	lines := strings.Split(content, "\n")
	rows := make([]string, 0, m.height)
	rows = append(rows, top)
	for i := 0; i < m.height-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		rows = append(rows, side+" "+fit(line, inner-2)+" "+side)
	}
	rows = append(rows, bottom)
	return strings.Join(rows, "\n")
}

// Tab renders a standalone title tab (e.g. the app name) in the same style as
// a panel's embedded title.
func Tab(label string) string {
	return tabWithBackground(label, theme.TitleBg)
}

// tabWithBackground renders a title tab on the given background, so a panel's
// tab always matches its border color.
func tabWithBackground(label string, background lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(theme.TitleFg).Background(background).Bold(true).Render(" " + label + " ")
}

// fit pads or truncates line to exactly width display cells, ANSI-aware.
func fit(line string, width int) string {
	if width <= 0 {
		return ""
	}
	lineWidth := lipgloss.Width(line)
	if lineWidth > width {
		return lipgloss.NewStyle().MaxWidth(width).Render(line)
	}
	return line + strings.Repeat(" ", width-lineWidth)
}
