package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

// tab style variables
var (
	docStyle         = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	selectionColor   = lipgloss.Color("#20C997")
	inactiveTabStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, false, true).UnsetBorderBottom().BorderForeground(highlightColor)
	activeTabStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, false, true).UnsetBorderBottom().BorderForeground(highlightColor).Background(selectionColor).Foreground(lipgloss.Color("#000000"))
	windowStyle      = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(0).Border(lipgloss.NormalBorder())
)

// table style
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// main model
type model struct {
	Tabs        []string
	activeTab   int
	serverTable table.Model
	routerTable table.Model
	logs        []string
    // setting I have no idea how to handle them yet...
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
            core.CleanupOnClose()
            log.Println("GRACEFULL QUIT!")
            log.Println("GRACEFULL QUIT!")
            log.Println("GRACEFULL QUIT!")
            log.Println("GRACEFULL QUIT!")
            log.Println("GRACEFULL QUIT!")
			return m, tea.Quit
		case "right", "l", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			// change table too
			return m, nil
		case "left", "h", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			// change table too
			return m, nil
		case "enter":
			// Handle selection for different tabs differently LUL
			if m.activeTab == 0 {
				// Do stuff to connect to selection
				return m, tea.Println("Connecting to: ", m.serverTable.SelectedRow()[1])
			}

			return m, tea.Println("Only VPN server list works for now.")
		}
		// Probably I'll need to handle more msg types for updates, erros, etc...
	}
	m.serverTable, cmd = m.serverTable.Update(msg)

	return m, cmd
}

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	// iterate through Tabs apply the correct style, render them and append them to the renderedTabs slice
	for i, t := range m.Tabs {
		var style lipgloss.Style
		isActive := i == m.activeTab

		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}

		renderedTabs = append(renderedTabs, style.Render(t))
	}

	// Use lipgloss Join to align the tabs horizontaly then build the rest of the string bit by bit
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")

	// This where the table view goes, I could not figure out a better way to do this
    var tabContent string
    switch m.activeTab {
    case 0:
	    tabContent = baseStyle.Render(m.serverTable.View()) 
    case 1:
        tabContent = baseStyle.Render(m.routerTable.View())
    case 2:
        logs, err := core.GetLogsForCLI()
        if logs != nil && err == nil {
            for i := range logs.Content {
                if logs.Content[i] != "" {
                    tabContent = fmt.Sprint(logs.Time[i], " || ", logs.Function[i], " || ", logs.Content[i]+"\n") 
                }
            }
        }
        tabContent = baseStyle.Render(tabContent)
    default:
        tabContent = baseStyle.Render("Not implemented yet!")
    }
	doc.WriteString(windowStyle.Width(80).Render(tabContent))

	return docStyle.Render(doc.String())
}

func StartTui() {
	// Configure tabs and their number
	tabs := []string{"VPN List", "Router List", "Logs", "Settings"}

	// Example table to have something to show inside the tabs
	// This will evetnually be constructed and updated(I haven't figured out how yet) with real data later
	col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 8},
		{Title: "QoS", Width: 4},
	}

	row := []table.Row{
		{"server-01", "SW", "10"},
		{"server-02", "SW", "9"},
		{"anotherserver-01", "US", "5"},
		{"someserver-01", "GR", "7"},
		{"serverlet-01", "FR", "8"},
		{"keybindssecretserver", "IS", "1"},
	}

	t := table.New(
		table.WithColumns(col),
		table.WithRows(row),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#ffffff")).
		BorderBottom(true).
		Bold(true).
		Width(26)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)

	s.Cell.Width(26)

	t.SetStyles(s)
	// example table construction end

	// make the model and give some starting values
	m := model{Tabs: tabs, serverTable: t, routerTable: t}

	// This is where it actually starts
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running TUI: ", err)
		os.Exit(1)
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

// little helpers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
