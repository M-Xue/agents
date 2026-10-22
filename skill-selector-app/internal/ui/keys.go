package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds the bindings rendered in the footer help bar.
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	NextPanel key.Binding
	PrevPanel key.Binding
	Quit      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Toggle:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle")),
		NextPanel: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		PrevPanel: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Quit:      key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
	}
}

// ShortHelp returns the bindings shown in the single-line footer, in order.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.NextPanel, k.PrevPanel, k.Quit}
}
