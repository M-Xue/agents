// Package searchbar renders a single-line text field on a dark background
// with a block caret, in the style of grove's text input. It owns only the
// typed value and caret position; what the value is used for is the caller's
// concern.
package searchbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/maxx/skill-selector-app/internal/ui/theme"
)

var (
	fieldBg    = lipgloss.Color("#1e2030")
	valueColor = lipgloss.Color("252")
	// The placeholder shares the inactive-border color so hint text and idle
	// chrome read as one muted layer.
	placeholderColor = theme.TextInactive
)

// Model is a single-line text field. Zero value is usable; New sets a hint.
type Model struct {
	value       string
	placeholder string
	focused     bool
	cursor      int
	width       int
	// offset is the index of the first visible character when the value is
	// longer than the field width. Indexes are byte-safe: values are ASCII.
	offset int
}

// New returns a field showing placeholder while empty.
func New(placeholder string) Model {
	return Model{placeholder: placeholder}
}

// SetPlaceholder changes the hint shown while the value is empty.
func (m *Model) SetPlaceholder(value string) { m.placeholder = value }

// SetValue replaces the value and moves the caret to its end.
func (m *Model) SetValue(value string) {
	m.value = filterASCII(value)
	m.cursor = len(m.value)
	m.offset = m.scrollStart()
}

// Value returns the current text.
func (m Model) Value() string { return m.value }

// Clear empties the field.
func (m *Model) Clear() {
	m.value = ""
	m.cursor = 0
	m.offset = 0
}

// Focus shows the caret and lets Update accept keys.
func (m *Model) Focus() { m.focused = true }

// Blur hides the caret and makes Update ignore keys.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the field accepts input.
func (m Model) Focused() bool { return m.focused }

// SetWidth fixes the rendered width in cells; the background pads out to it.
// Zero renders at the content's natural width.
func (m *Model) SetWidth(width int) { m.width = width }

// Update applies a key to the field and reports whether it was consumed.
// Printable characters and space insert; backspace, delete, arrows, home and
// end edit or move. Everything else, including enter and up/down, is left
// for the caller.
func (m *Model) Update(msg tea.KeyMsg) bool {
	if !m.focused {
		return false
	}
	consumed := m.handleKey(msg)
	if consumed {
		m.offset = m.scrollStart()
	}
	return consumed
}

func (m *Model) handleKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyLeft:
		if m.cursor > 0 {
			m.cursor--
		}
		return true
	case tea.KeyRight:
		if m.cursor < len(m.value) {
			m.cursor++
		}
		return true
	case tea.KeyHome, tea.KeyCtrlA:
		m.cursor = 0
		return true
	case tea.KeyEnd, tea.KeyCtrlE:
		m.cursor = len(m.value)
		return true
	case tea.KeyBackspace:
		if m.cursor > 0 {
			m.value = m.value[:m.cursor-1] + m.value[m.cursor:]
			m.cursor--
		}
		return true
	case tea.KeyDelete:
		if m.cursor < len(m.value) {
			m.value = m.value[:m.cursor] + m.value[m.cursor+1:]
		}
		return true
	case tea.KeyRunes, tea.KeySpace:
		if msg.Alt {
			return false
		}
		runes := filterASCII(msg.String())
		m.value = m.value[:m.cursor] + runes + m.value[m.cursor:]
		m.cursor += len(runes)
		return true
	default:
		return false
	}
}

// View renders the field. When focused, the window of the value around the
// caret is shown so long text scrolls rather than overflowing; when blurred,
// the head of the value is shown clipped.
func (m Model) View() string {
	base := lipgloss.NewStyle().Background(fieldBg)
	text := base.Foreground(valueColor)
	placeholder := base.Foreground(placeholderColor)
	// The caret is value colour and field background reversed, so it reads as
	// a light block regardless of terminal defaults. It is the only focus cue.
	caret := base.Foreground(valueColor).Reverse(true)

	value, at := m.visibleWindow()
	var parts []string
	switch {
	case m.value == "" && m.focused:
		parts = append(parts, caret.Render(" "), placeholder.Render(m.placeholder))
	case m.value == "":
		parts = append(parts, placeholder.Render(m.placeholder))
	case !m.focused:
		parts = append(parts, text.Render(clipTo(m.value, m.width)))
	case at >= len(value):
		parts = append(parts, text.Render(value), caret.Render(" "))
	default:
		parts = append(parts,
			text.Render(value[:at]),
			caret.Render(value[at:at+1]),
			text.Render(value[at+1:]),
		)
	}
	return m.fill(strings.Join(parts, ""), base)
}

// scrollStart returns the offset of the visible window, keeping the caret in
// view: unchanged while the caret stays visible, shifted minimally otherwise.
func (m Model) scrollStart() int {
	if m.width <= 0 {
		return 0
	}
	start := m.offset
	if m.cursor-start > m.width-1 {
		start = m.cursor - (m.width - 1)
	}
	if m.cursor < start {
		start = m.cursor
	}
	return max(0, start)
}

// visibleWindow returns the slice of the value that fits the width with the
// caret in view, and the caret's index within that slice.
func (m Model) visibleWindow() (string, int) {
	if m.width <= 0 {
		return m.value, m.cursor
	}
	start := m.scrollStart()
	end := min(len(m.value), start+m.width)
	return m.value[start:end], m.cursor - start
}

// clipTo truncates value to width characters; zero width means unlimited.
func clipTo(value string, width int) string {
	if width <= 0 || len(value) <= width {
		return value
	}
	return value[:width]
}

// fill pads the rendered content out to the configured width with
// field-background spaces so the box reads as one solid field.
func (m Model) fill(content string, base lipgloss.Style) string {
	if m.width <= 0 {
		return content
	}
	gap := m.width - lipgloss.Width(content)
	if gap <= 0 {
		return content
	}
	return content + base.Render(strings.Repeat(" ", gap))
}

func filterASCII(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= 32 && r <= 126 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
