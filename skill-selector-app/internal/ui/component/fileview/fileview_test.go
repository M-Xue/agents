package fileview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestWrapAndScroll(t *testing.T) {
	m := New()
	m.SetContent("one two three four five six\n\nshort")
	m.SetSize(10, 2)
	if m.LineCount() < 4 {
		t.Fatalf("expected wrapping, got %d lines: %q", m.LineCount(), m.lines)
	}
	for _, l := range m.lines {
		if lipgloss.Width(l) > 10 {
			t.Fatalf("line wider than 10: %q", l)
		}
	}
	for i := 0; i < 100; i++ {
		m.ScrollDown()
	}
	if m.Offset() != m.LineCount()-2 {
		t.Fatalf("offset %d, want %d", m.Offset(), m.LineCount()-2)
	}
	for i := 0; i < 100; i++ {
		m.ScrollUp()
	}
	if m.Offset() != 0 {
		t.Fatalf("offset after scrolling up = %d", m.Offset())
	}
}

func TestViewHeight(t *testing.T) {
	m := New()
	m.SetContent("a\nb")
	m.SetSize(5, 4)
	if n := len(strings.Split(m.View(), "\n")); n != 4 {
		t.Fatalf("got %d rows", n)
	}
}

func TestSetContentKeepsOffset(t *testing.T) {
	m := New()
	m.SetContent("a\nb\nc\nd\ne\nf")
	m.SetSize(5, 2)
	m.ScrollDown()
	m.ScrollDown()
	m.SetContent("a\nb\nc\nd\ne\nf\ng")
	if m.Offset() != 2 {
		t.Fatalf("offset after SetContent = %d, want 2", m.Offset())
	}
	m.SetContent("a\nb")
	if m.Offset() != 0 {
		t.Fatalf("offset not clamped to shorter content: %d", m.Offset())
	}
	m.ScrollToTop()
	if m.Offset() != 0 {
		t.Fatal("ScrollToTop failed")
	}
}
