package list

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
	m.MoveUp()
	if sel, _ := m.Selected(); sel.ID != "9" || m.offset != 6 {
		t.Fatalf("wrap to end: id=%q offset=%d", sel.ID, m.offset)
	}
	m.MoveDown()
	if sel, _ := m.Selected(); sel.ID != "0" || m.offset != 0 {
		t.Fatalf("wrap to start: id=%q offset=%d", sel.ID, m.offset)
	}
}

func TestSetItemsKeepsCursorOrFallsToTop(t *testing.T) {
	m := New()
	m.SetItems(items(10))
	m.SetSize(20, 3)
	for i := 0; i < 7; i++ {
		m.MoveDown()
	}
	m.SetItems([]Item{{ID: "2"}, {ID: "7"}, {ID: "9"}})
	if sel, _ := m.Selected(); sel.ID != "7" {
		t.Fatalf("cursor not kept: %q", sel.ID)
	}
	m.SetItems([]Item{{ID: "1"}, {ID: "4"}})
	if sel, _ := m.Selected(); sel.ID != "1" || m.offset != 0 {
		t.Fatalf("fallback: id=%q offset=%d", sel.ID, m.offset)
	}
}

func TestViewHeight(t *testing.T) {
	m := New()
	m.SetItems(items(2))
	m.SetSize(10, 5)
	if n := len(strings.Split(m.View(), "\n")); n != 5 {
		t.Fatalf("got %d lines", n)
	}
}

func TestScrollMovesWindowNotCursor(t *testing.T) {
	m := New()
	m.SetItems(items(10))
	m.SetSize(20, 4)

	m.ScrollDown()
	m.ScrollDown()
	if sel, _ := m.Selected(); sel.ID != "0" || m.Offset() != 2 {
		t.Fatalf("after 2 scroll downs: id=%q offset=%d", sel.ID, m.Offset())
	}
	// Re-sizing to the same size every render must not snap back to the cursor.
	m.SetSize(20, 4)
	if m.Offset() != 2 {
		t.Fatalf("same-size SetSize reset offset to %d", m.Offset())
	}
	// A height change does bring the cursor back into view.
	m.SetSize(20, 3)
	if m.Offset() != 0 {
		t.Fatalf("height change left cursor hidden, offset=%d", m.Offset())
	}
	m.SetSize(20, 4)
	for i := 0; i < 20; i++ {
		m.ScrollDown()
	}
	if m.Offset() != 6 {
		t.Fatalf("scroll past end: offset=%d, want 6", m.Offset())
	}
	for i := 0; i < 20; i++ {
		m.ScrollUp()
	}
	if m.Offset() != 0 {
		t.Fatalf("scroll past start: offset=%d, want 0", m.Offset())
	}
	// A cursor move after scrolling away brings the cursor back into view.
	for i := 0; i < 6; i++ {
		m.ScrollDown()
	}
	m.MoveDown()
	if sel, _ := m.Selected(); sel.ID != "1" || m.Offset() != 1 {
		t.Fatalf("MoveDown after scroll: id=%q offset=%d", sel.ID, m.Offset())
	}
}
