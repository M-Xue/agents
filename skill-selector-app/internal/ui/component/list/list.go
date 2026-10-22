// Package list renders a vertically scrolling list of labelled rows. The row
// under the cursor is bold. It is the checklist without the checked state:
// the component owns only the cursor and scroll offset.
package list

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Item is one row of the list.
type Item struct {
	ID    string
	Label string
}

// Model is a scrollable list. Zero value is usable.
type Model struct {
	items  []Item
	cursor int
	offset int
	width  int
	height int
}

// New returns an empty list.
func New() Model { return Model{} }

// SetItems replaces the list. The cursor stays on the item with the same ID
// if it is still present, otherwise it moves to the top row.
func (m *Model) SetItems(items []Item) {
	prevID := ""
	if cur, ok := m.Selected(); ok {
		prevID = cur.ID
	}
	m.items = append(m.items[:0:0], items...)
	m.cursor = 0
	for i, it := range m.items {
		if it.ID == prevID {
			m.cursor = i
			break
		}
	}
	m.clamp()
}

// SetSize sets the visible area in cells. height is the number of rows shown
// at once; the list scrolls within it.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.clamp()
}

// MoveDown advances the cursor, wrapping from the last item to the first.
func (m *Model) MoveDown() {
	if len(m.items) == 0 {
		return
	}
	m.cursor = (m.cursor + 1) % len(m.items)
	m.clamp()
}

// MoveUp retreats the cursor, wrapping from the first item to the last.
func (m *Model) MoveUp() {
	if len(m.items) == 0 {
		return
	}
	m.cursor = (m.cursor + len(m.items) - 1) % len(m.items)
	m.clamp()
}

// Selected returns the item under the cursor, or false when the list is empty.
func (m Model) Selected() (Item, bool) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return Item{}, false
	}
	return m.items[m.cursor], true
}

// Len returns the number of items.
func (m Model) Len() int { return len(m.items) }

// clamp keeps the cursor in range and scrolls the window so the cursor is
// visible, snapping the window to the list bounds.
func (m *Model) clamp() {
	n := len(m.items)
	if n == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	m.cursor = max(0, min(m.cursor, n-1))
	if m.height <= 0 {
		m.offset = 0
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
	m.offset = max(0, min(m.offset, max(0, n-m.height)))
}

// View renders the visible window of rows, one item per line, each fitted to
// the component width. Rows past the end of the list are left blank so the
// output is always height lines tall.
func (m Model) View() string {
	if m.height <= 0 {
		return ""
	}
	bold := lipgloss.NewStyle().Bold(true)
	rows := make([]string, 0, m.height)
	for i := 0; i < m.height; i++ {
		idx := m.offset + i
		if idx >= len(m.items) {
			rows = append(rows, fit("", m.width))
			continue
		}
		line := fit(m.items[idx].Label, m.width)
		if idx == m.cursor {
			line = bold.Render(line)
		}
		rows = append(rows, line)
	}
	return strings.Join(rows, "\n")
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
