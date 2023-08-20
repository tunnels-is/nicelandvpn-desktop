package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

// main model
type model struct {
	Tabs         []string
	activeTab    int
	serverTable  table.Model
	routerTable  table.Model
	logsViewport viewport.Model
	logs         []string
	ready        bool
	// setting I have no idea how to handle them yet...
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		servCmd tea.Cmd
		routCmd tea.Cmd
		vpCmd   tea.Cmd
	)

	m.serverTable, servCmd = m.serverTable.Update(msg)
	m.routerTable, routCmd = m.routerTable.Update(msg)
	m.logsViewport, vpCmd = m.logsViewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.serverTable.SetWidth(max(msg.Width-8, 72))
		m.serverTable.SetHeight(max(msg.Height-8, 17))

		m.routerTable.SetWidth(max(msg.Width-8, 72))
		m.routerTable.SetHeight(max(msg.Height-8, 17))

		m.logsViewport.Width = max(msg.Width-8, 72)
		m.logsViewport.Height = max(msg.Height-6, 19)

		return m, tea.Println("Resized!")
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			core.CleanupOnClose()
			log.Println("GRACEFULL QUIT!")
			return m, tea.Quit
		case "right", "l", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "h", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case "enter":
			// Handle selection for different tabs differently LUL
			switch m.activeTab {
			case 0:
				return m, tea.Println("Connecting to: ", m.serverTable.SelectedRow()[0])
			default:
				return m, tea.Println("This thing is still work in progress....")
			}
		}
		// update the logs always
	default:
        logs := GetLogs()
        if len(m.logs) != len(logs) {
            m.logs = logs
        }
		m.logsViewport.SetContent(strings.Join(m.logs, ""))
		if m.logsViewport.ScrollPercent() == 1 {
			m.logsViewport.GotoBottom()
		}

		return m, tea.Batch(servCmd, routCmd, vpCmd)
	}

	return m, tea.Batch(servCmd, routCmd, vpCmd)
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
		tabContent = baseStyle.Render(m.logsViewport.View())
	default:
		tabContent = baseStyle.Render("Not implemented yet!")
	}
	doc.WriteString(windowStyle.Render(tabContent))

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
		{"server-01", "SW", "10"},
		{"server-02", "SW", "9"},
		{"anotherserver-01", "US", "5"},
		{"someserver-01", "GR", "7"},
		{"serverlet-01", "FR", "8"},
		{"keybindssecretserver", "IS", "1"},
		{"server-01", "SW", "10"},
		{"server-02", "SW", "9"},
		{"anotherserver-01", "US", "5"},
		{"someserver-01", "GR", "7"},
		{"serverlet-01", "FR", "8"},
		{"keybindssecretserver", "IS", "1"},
		{"server-01", "SW", "10"},
		{"server-02", "SW", "9"},
		{"anotherserver-01", "US", "5"},
		{"someserver-01", "GR", "7"},
		{"serverlet-01", "FR", "8"},
		{"keybindssecretserver", "IS", "1"},
		{"server-01", "SW", "10"},
		{"server-02", "SW", "9"},
		{"anotherserver-01", "US", "5"},
		{"someserver-01", "GR", "7"},
		{"serverlet-01", "FR", "8"},
		{"keybindssecretserver", "IS", "1"},
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
	)
	// example table construction end
	t.SetStyles(table_style)

	// Initialize the viewport for the logs
	vp := viewport.New(80, 20)
	vp.Style = baseStyle.UnsetBorderStyle()

	// make the model and give some starting values
	m := model{Tabs: tabs, serverTable: t, routerTable: t, logsViewport: vp}

	// This is where it actually starts
	TUI = tea.NewProgram(m)
	if _, err := TUI.Run(); err != nil {
		fmt.Println("Error running TUI: ", err)
		os.Exit(1)
	}
}
