package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Select     key.Binding
	Disconnect key.Binding
	Left       key.Binding
	Right      key.Binding
	Help       key.Binding
	Quit       key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "Move up"),
	),

	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "Move down"),
	),

	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Select"),
	),

	Disconnect: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("shift+d", "Disconnect from server"),
	),

	Left: key.NewBinding(
		key.WithKeys("left", "h", "shift+tab"),
		key.WithHelp("←/h/shift+tab", "Move to the left tab"),
	),

	Right: key.NewBinding(
		key.WithKeys("right", "l", "tab"),
		key.WithHelp("→/l/tab", "Move to the right tab"),
	),

	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "Show help"),
	),

	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "Quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Disconnect, k.Help, k.Quit},
	}
}
