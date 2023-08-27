package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

// main model
type model struct {
	tabs         []string
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
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
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
	for i, t := range m.tabs {
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
	// send them to the login form 1st
	// I do not think I can have 2 completely different models
	// in bubbletea this is the only way I could figure out
	// how to do it...
	login()
	time.Sleep(3 * time.Second)

	var CSR core.CONTROLLER_SESSION_REQUEST
	CSR.DeviceToken = user.DeviceToken.DT
	CSR.GROUP = core.GLOBAL_STATE.ActiveRouter.GROUP
	CSR.ROUTERID = core.GLOBAL_STATE.ActiveRouter.ROUTERID

	data, code, err := core.Connect(&CSR, true)
	fmt.Printf("\n%+v\n", data)
	fmt.Println(code)
	if err != nil {
		fmt.Println("Error connecting: ", err)
	}

	time.Sleep(10 * time.Second)
	logout()

	// Configure tabs and their number
	// tabs := []string{"VPN List", "Router List", "Logs", "Settings"}

	// Initial Sever Table
	// Columns for server table
	s_col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 8},
		{Title: "QoS", Width: 4},
	}

	// get the initial values for the servers table
	s_row := []table.Row{}

	s_t := table.New(
		table.WithColumns(s_col),
		table.WithRows(s_row),
		table.WithFocused(true),
	)
	s_t.SetStyles(table_style)

	// Initial Routers Table
	// Columns for routers table
	r_col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 8},
		{Title: "QoS", Width: 4},
	}

	// get the initial values for the routers table
	r_row := []table.Row{}

	r_t := table.New(
		table.WithColumns(r_col),
		table.WithRows(r_row),
		table.WithFocused(true),
	)
	r_t.SetStyles(table_style)

	// Initialize the viewport for the logs
	vp := viewport.New(80, 20)
	vp.Style = baseStyle.UnsetBorderStyle()

	// make the model and give some starting values
	// m := model{tabs: tabs, serverTable: s_t, routerTable: r_t, logsViewport: vp}

	// This is where it actually starts
	// TUI = tea.NewProgram(m)
	// if _, err := TUI.Run(); err != nil {
	// 	fmt.Println("Error running TUI: ", err)
	// 	os.Exit(1)
	// }
}

func logout() {
	// construct the logout form
	var FR core.FORWARD_REQUEST
	FR.Path = "v2/user/logout"
	FR.JSONData = core.LogoutForm{
		Email:       user.Email,
		DeviceToken: user.DeviceToken.DT,
	}

	// Send logout request
	core.Disconnect()
	data, code, err := core.ForwardToController(&FR)
	fmt.Println("Logging out...")
	fmt.Println(data)
	fmt.Println(code)
	fmt.Println(err)
}
