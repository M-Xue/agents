package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap holds the bindings used for matching keys and rendered in the
// footer help bar.
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	ToggleAll key.Binding
	NextPanel key.Binding
	PrevPanel key.Binding
	Quit      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		// Up and down are matched before the search bars see a key, so they
		// are bound to the arrow keys only. Letters and space fall through to
		// the focused search bar, which is why toggle is enter rather than
		// space and why j and k are not navigation keys.
		Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		Down:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		Toggle: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle")),
		// ctrl+enter is not distinguishable from enter in a terminal (both
		// send a carriage return), so toggle all uses ctrl+a. It is matched
		// only while Skills has focus, so ctrl+a still means home in the
		// Skill Libraries search.
		ToggleAll: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "toggle all")),
		NextPanel: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		PrevPanel: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Quit:      key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
	}
}

// helpFor returns the bindings shown in the single-line footer while the
// given panel has focus, in order. Every panel lists up, down, the panel
// switches and quit; Skills alone appends toggle and toggle all at the end.
func (k keyMap) helpFor(focused panelID) []key.Binding {
	common := []key.Binding{k.Up, k.Down, k.NextPanel, k.PrevPanel, k.Quit}
	if focused == panelSkills {
		return append(common, k.Toggle, k.ToggleAll)
	}
	return common
}
