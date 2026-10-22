package searchbar

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func typeString(m *Model, s string) {
	for _, r := range s {
		if r == ' ' {
			m.Update(tea.KeyMsg{Type: tea.KeySpace})
			continue
		}
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestTypingSpaceAndBackspace(t *testing.T) {
	m := New("Search")
	m.Focus()
	typeString(&m, "code rev")
	if m.Value() != "code rev" {
		t.Fatalf("value = %q", m.Value())
	}
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.Value() != "code re" {
		t.Fatalf("after backspace = %q", m.Value())
	}
}

func TestEnterAndArrowsNotConsumed(t *testing.T) {
	m := New("Search")
	m.Focus()
	for _, k := range []tea.KeyType{tea.KeyEnter, tea.KeyUp, tea.KeyDown, tea.KeyTab} {
		if m.Update(tea.KeyMsg{Type: k}) {
			t.Fatalf("key %v should not be consumed", k)
		}
	}
}

func TestBlurredIgnoresKeys(t *testing.T) {
	m := New("Search")
	typeString(&m, "abc")
	if m.Value() != "" {
		t.Fatalf("blurred field accepted input: %q", m.Value())
	}
}
