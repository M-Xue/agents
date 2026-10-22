// Package fileview renders a block of text inside a fixed area, soft-wrapping
// long lines to the width and scrolling vertically by line. It owns only the
// scroll offset and the wrapped lines; where the text came from is the
// caller's concern.
package fileview

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Model is a scrollable text viewer. Zero value is usable.
type Model struct {
	content string
	lines   []string // content wrapped to width
	offset  int
	width   int
	height  int
}

// New returns an empty viewer.
func New() Model { return Model{} }

// SetContent replaces the text and scrolls back to the top.
func (m *Model) SetContent(content string) {
	m.content = content
	m.offset = 0
	m.rewrap()
}

// SetSize sets the visible area in cells and re-wraps the text to the width.
func (m *Model) SetSize(width, height int) {
	if width != m.width {
		m.width = width
		m.rewrap()
	}
	m.height = height
	m.clamp()
}

// ScrollDown moves the window one line toward the end, stopping there.
func (m *Model) ScrollDown() {
	m.offset++
	m.clamp()
}

// ScrollUp moves the window one line toward the start, stopping there.
func (m *Model) ScrollUp() {
	m.offset--
	m.clamp()
}

// LineCount returns the number of wrapped lines.
func (m Model) LineCount() int { return len(m.lines) }

// Offset returns the index of the first visible wrapped line.
func (m Model) Offset() int { return m.offset }

// rewrap splits the content into display lines no wider than width.
func (m *Model) rewrap() {
	m.lines = m.lines[:0]
	if m.content == "" {
		return
	}
	wrap := lipgloss.NewStyle().Width(m.width)
	for _, raw := range strings.Split(m.content, "\n") {
		if m.width <= 0 || raw == "" {
			m.lines = append(m.lines, raw)
			continue
		}
		m.lines = append(m.lines, strings.Split(wrap.Render(raw), "\n")...)
	}
	m.clamp()
}

// clamp keeps the window within the text: never past the last page and
// never before the first line.
func (m *Model) clamp() {
	maxOffset := max(0, len(m.lines)-max(m.height, 0))
	m.offset = max(0, min(m.offset, maxOffset))
}

// View renders the visible window, one wrapped line per row, each fitted to
// the width. Rows past the end are blank so the output is height lines tall.
func (m Model) View() string {
	if m.height <= 0 {
		return ""
	}
	rows := make([]string, 0, m.height)
	for i := 0; i < m.height; i++ {
		idx := m.offset + i
		line := ""
		if idx < len(m.lines) {
			line = m.lines[idx]
		}
		rows = append(rows, fit(line, m.width))
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
