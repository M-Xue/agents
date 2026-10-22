package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines the key bindings for the screen. It satisfies help.KeyMap
// so the help component can render it.
type keyMap struct {
	NextPanel key.Binding
	Quit      key.Binding
}

// ShortHelp returns the bindings shown in the single-line help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextPanel, k.Quit}
}

// FullHelp returns the bindings shown in the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.NextPanel, k.Quit}}
}

var keys = keyMap{
	NextPanel: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next panel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
