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

func TestItemsIsACopy(t *testing.T) {
	m := New()
	m.SetItems([]Item{{ID: "a"}, {ID: "b", Checked: true}})
	got := m.Items()
	if len(got) != 2 || got[1].ID != "b" || !got[1].Checked {
		t.Fatalf("Items = %+v", got)
	}
	got[0].ID = "mutated"
	if m.Items()[0].ID != "a" {
		t.Fatal("Items returned the internal slice")
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

func TestSelect(t *testing.T) {
	m := New()
	m.SetItems(items(10))
	m.SetSize(20, 4)
	if !m.Select(7) {
		t.Fatal("Select(7) reported out of range")
	}
	if sel, _ := m.Selected(); sel.ID != "7" || m.Offset() != 4 {
		t.Fatalf("after Select(7): id=%q offset=%d", sel.ID, m.Offset())
	}
	if m.Select(10) || m.Select(-1) {
		t.Fatal("out-of-range Select reported success")
	}
	if sel, _ := m.Selected(); sel.ID != "7" {
		t.Fatalf("out-of-range Select moved the cursor to %q", sel.ID)
	}
}

// Colour is stripped when tests run without a terminal, so the dimming
// itself cannot be asserted here; this checks everything else about a
// disabled row.
func TestDisabledRowKeepsGlyphAndLabel(t *testing.T) {
	dim := New()
	list := items(2)
	list[1].Disabled = true
	list[1].Checked = true // the glyph follows Checked regardless of Disabled
	dim.SetItems(list)
	dim.SetSize(20, 2)

	rows := strings.Split(dim.View(), "\n")
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	if !strings.Contains(rows[1], GlyphChecked) || !strings.Contains(rows[1], "item 1") {
		t.Fatalf("disabled row lost its glyph or label: %q", rows[1])
	}
	list[1].Checked = false
	dim.SetItems(list)
	if rows := strings.Split(dim.View(), "\n"); !strings.Contains(rows[1], GlyphClear) {
		t.Fatalf("unchecked disabled row did not show the clear glyph: %q", rows[1])
	}
	// Disabled is view state only: the cursor still lands on the row and it
	// is reported as selected.
	dim.MoveDown()
	if sel, _ := dim.Selected(); sel.ID != "1" || !sel.Disabled {
		t.Fatalf("cursor did not land on the disabled row: %+v", sel)
	}
}
