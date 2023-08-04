package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func initialModel() model {
	return model{
		choices: []string{"Login", "Show Logs", "??"},

		selected: make(map[int]struct{}),
	}
}

func TimedUIUpdate(MONITOR chan int) {
	defer func() {
		time.Sleep(1 * time.Second)
		MONITOR <- 3
	}()
	defer core.RecoverAndLogToFile()

	for {
		time.Sleep(1 * time.Second)
		TUI.Send(&tea.KeyMsg{
			Type: 0,
		})
	}

}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	default:
		// log.Println(msg)
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			core.CleanupOnClose() // THIS FUNCTION NEEDS TO BE CALLED ON EXIT!!
			log.Println("GRACEFULL QUIT")
			log.Println("GRACEFULL QUIT")
			log.Println("GRACEFULL QUIT")
			log.Println("GRACEFULL QUIT")
			log.Println("GRACEFULL QUIT")
			log.Println("GRACEFULL QUIT")
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	var s string = ""
	var selectedIndex int = 0
	for i, _ := range m.choices {
		if _, ok := m.selected[i]; ok {
			selectedIndex = i

		}
	}

	if selectedIndex == 1 {

		s += fmt.Sprint("\n")
		s += fmt.Sprint("\n")
		LR, err := core.GetLogsForCLI()
		if LR != nil && err == nil {
			for i := range LR.Content {
				if LR.Content[i] != "" {
					s += fmt.Sprint(LR.Time[i], " || ", LR.Function[i], " || ", LR.Content[i]+"\n")
				}
			}
		}
	}

	s += "\n\nWhat should we buy at the market?\n\n"
	s += "\nPress q to quit.\n"

	for i, choice := range m.choices {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			selectedIndex = i
			checked = "x"

		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	return s
}
