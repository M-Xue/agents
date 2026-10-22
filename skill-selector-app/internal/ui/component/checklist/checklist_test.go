package checklist

import (
	"fmt"
	"strings"
	"testing"
)

func items(n int) []Item {
	out := make([]Item, n)
	for i := range out {
		out[i] = Item{ID: fmt.Sprint(i), Label: fmt.Sprintf("item %d", i)}
	}
	return out
}

func TestWrapAndScroll(t *testing.T) {
	m := New()
	m.SetItems(items(10))
	m.SetSize(20, 4)

	m.MoveUp() // wrap from top to bottom
	if sel, _ := m.Selected(); sel.ID != "9" {
		t.Fatalf("after MoveUp from 0: %q", sel.ID)
	}
	if m.offset != 6 {
		t.Fatalf("offset after wrap to end = %d, want 6", m.offset)
	}
	m.MoveDown() // wrap from bottom to top
	if sel, _ := m.Selected(); sel.ID != "0" || m.offset != 0 {
		t.Fatalf("after wrap to start: id=%q offset=%d", sel.ID, m.offset)
	}
	for i := 0; i < 5; i++ {
		m.MoveDown()
	}
	if sel, _ := m.Selected(); sel.ID != "5" || m.offset != 2 {
		t.Fatalf("after 5 downs: id=%q offset=%d", sel.ID, m.offset)
	}
}

func TestSetItemsKeepsCursorByID(t *testing.T) {
	m := New()
	m.SetItems(items(5))
	m.SetSize(20, 5)
	m.MoveDown()
	m.MoveDown()
	next := items(5)
	next[2].Checked = true
	m.SetItems(next)
	sel, _ := m.Selected()
	if sel.ID != "2" || !sel.Checked {
		t.Fatalf("cursor not preserved: %+v", sel)
	}
}

func TestViewGlyphsAndHeight(t *testing.T) {
	m := New()
	list := items(3)
	list[1].Checked = true
	m.SetItems(list)
	m.SetSize(20, 5)
	lines := strings.Split(m.View(), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}
	if !strings.Contains(lines[0], GlyphClear) || !strings.Contains(lines[1], GlyphChecked) {
		t.Fatalf("glyphs wrong:\n%s", m.View())
	}
}

func TestSetItemsFallsBackToTopWhenCursorRemoved(t *testing.T) {
	m := New()
	m.SetItems(items(10))
	m.SetSize(20, 3)
	for i := 0; i < 7; i++ {
		m.MoveDown()
	}
	if m.offset == 0 {
		t.Fatal("expected list to have scrolled")
	}
	// Filter out the cursor row (id 7); the cursor should snap to the top row
	// and the window should scroll back with it.
	m.SetItems([]Item{{ID: "1", Label: "one"}, {ID: "4", Label: "four"}})
	sel, ok := m.Selected()
	if !ok || sel.ID != "1" || m.offset != 0 {
		t.Fatalf("got id=%q offset=%d", sel.ID, m.offset)
	}
}
