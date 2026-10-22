package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds the bindings rendered in the footer help bar.
type keyMap struct {
	NextPanel key.Binding
	PrevPanel key.Binding
	Quit      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		NextPanel: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		PrevPanel: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Quit:      key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
	}
}

// ShortHelp returns the bindings shown in the single-line footer, in order.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextPanel, k.PrevPanel, k.Quit}
}
