package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
	status       []string
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
		m.serverTable.SetHeight(max(msg.Height-8, 16))

		m.routerTable.SetWidth(max(msg.Width-8, 72))
		m.routerTable.SetHeight(max(msg.Height-8, 16))

		m.logsViewport.Width = max(msg.Width-8, 72)
		m.logsViewport.Height = max(msg.Height-8, 16)

		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "D":
			core.Disconnect()
			return m, tea.Println("Disconnected!")
		case "ctrl+c", "ctrl+d", "q":
			logout()
			core.CleanupOnClose()
			log.Println("GRACEFULL QUIT!")
			return m, tea.Quit
		case "right", "l", "tab":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			if m.activeTab == 0 {
				m.serverTable.Focus()
				m.routerTable.Blur()
			} else if m.activeTab == 1 {
				m.serverTable.Blur()
				m.routerTable.Focus()
      } else if m.activeTab == 2 {
        m.logsViewport.GotoBottom()
			} else {
				m.serverTable.Blur()
				m.routerTable.Blur()
			}
			return m, nil
		case "left", "h", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			if m.activeTab == 0 {
				m.serverTable.Focus()
				m.routerTable.Blur()
			} else if m.activeTab == 1 {
				m.serverTable.Blur()
				m.routerTable.Focus()
      } else if m.activeTab == 2 {
        m.logsViewport.GotoBottom()
      } else {
				m.serverTable.Blur()
				m.routerTable.Blur()
			}
			return m, nil
		case "enter":
			// Handle selection for different tabs differently LUL
			switch m.activeTab {
			case 0:
				// connect to access point
				ConnectToAP(m.serverTable.SelectedRow()[0])
				return m, tea.Println("Connecting to: ", m.serverTable.SelectedRow()[0])
			case 1:
				// change to router
				code, err := core.SwitchRouter(m.routerTable.SelectedRow()[0])
				s := "Switching to " + m.routerTable.SelectedRow()[0]
				if err != nil {
					s = fmt.Sprintf("There was an error switching routers: %s\nCode: %d", err, code)
				}
				return m, tea.Println(s)
			default:
				return m, tea.Println("This thing is still work in progress...")
			}
		}
	default:
		// update the logs and tables
		logs := GetLogs()
		if len(m.logs) != len(logs) {
			m.logs = logs
		}
		m.logsViewport.SetContent(strings.Join(m.logs, ""))
		if m.logsViewport.ScrollPercent() >= 0.9 {
			m.logsViewport.GotoBottom()
		}

		aps := core.GLOBAL_STATE.AccessPoints
		s_row := []table.Row{}
		for _, v := range aps {
			s_row = append(s_row, table.Row{v.Router.Tag, v.GEO.Country, strconv.Itoa(v.Router.Score)})
		}

		routs := core.GLOBAL_STATE.Routers
		r_row := []table.Row{}
		for _, v := range routs {
			r_row = append(r_row, table.Row{v.Tag, v.Country, strconv.Itoa(v.Score)})
		}

		m.serverTable.SetRows(s_row)
		m.routerTable.SetRows(r_row)

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
  doc.WriteString("\n")

	// Status line at the bottom
	var status string
	if core.GLOBAL_STATE.Connected {
		// core.GLOBAL_STATE.ActiveAccessPoint.Tag throws a panic with runtime error: invalid memory address or nil pointer dereference.
		status = "Router: " + core.GLOBAL_STATE.ActiveRouter.Tag + "\tVPN: " + core.GLOBAL_STATE.ActiveAccessPoint.Tag
		// status = statusStyle.Render("Router: " + core.GLOBAL_STATE.ActiveRouter.Tag + "\tVPN: Connected")
	} else {
		status = "Router: " + core.GLOBAL_STATE.ActiveRouter.Tag + "\tVPN: Not Connected"
	}

  stats := "\tUp: " + strconv.Itoa(core.GLOBAL_STATE.UMbps)  + "   " + "Down: " + strconv.Itoa(core.GLOBAL_STATE.DMbps) 
  status = lipgloss.JoinHorizontal(lipgloss.Left, status, stats)
  doc.WriteString(statusStyle.Render(status))

	// return lipgloss.JoinVertical(lipgloss.Left, ret, status)
  return docStyle.Render(doc.String())
}

func StartTui() {
	// send them to the login form 1st
	// I do not think I can have 2 completely different models
	// in bubbletea this is the only way I could figure out
	// how to do it...
	login()

	// Initial VPNs and Routers tables
	// I thought it's a good idea to have the
	// tables ready before I start the TUI
	// Construct the request for it 1st
	jsonData := make(map[string]string)
	jsonData["UID"] = user.ID.Hex()
	jsonData["DeviceToken"] = user.DeviceToken.DT
	PAFR = core.FORWARD_REQUEST{
		Method:   "POST",
		Path:     "devices/private",
		JSONData: jsonData,
	}

	if PAFR.JSONData != nil {
		core.GetRoutersAndAccessPoints(&PAFR)
	}

	// Configure tabs and their number
	tabs := []string{"VPN List", "Router List", "Logs", "Settings"}

	// Initialize the servers and routers tables
	// Columns for server table
	s_col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 8},
		{Title: "QoS", Width: 4},
	}

	// Columns for routers table
	r_col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 8},
		{Title: "QoS", Width: 4},
	}

	var s_row []table.Row
	var r_row []table.Row

	aps := core.GLOBAL_STATE.AccessPoints
	routs := core.GLOBAL_STATE.Routers

	// rows for server table
	for _, v := range aps {
		s_row = append(s_row, table.Row{v.Router.Tag, v.GEO.Country, strconv.Itoa(v.Router.Score)})
	}

	// rows for router table
	for _, v := range routs {
		r_row = append(r_row, table.Row{v.Tag, v.Country, strconv.Itoa(v.Score)})
	}

	// Set tables
	s_t := table.New(
		table.WithColumns(s_col),
		table.WithRows(s_row),
	)
	s_t.SetStyles(table_style)

	r_t := table.New(
		table.WithColumns(r_col),
		table.WithRows(r_row),
	)
	r_t.SetStyles(table_style)
	// Initial tables finished ---

	// Initialize the viewport for the logs
	vp := viewport.New(78, 17)
	vp.Style = baseStyle.UnsetBorderStyle()

	// make the model and give some starting values
	m := model{tabs: tabs, serverTable: s_t, routerTable: r_t, logsViewport: vp}
	m.serverTable.Focus() // focus on the first table since it starts there

	// This is where it actually starts
	TUI = tea.NewProgram(m)
	go TimedUIUpdate(MONITOR)
	if _, err := TUI.Run(); err != nil {
		fmt.Println("Error running TUI: ", err)
		core.CleanupOnClose()
		os.Exit(1)
	}
}

func ConnectToAP(Tag string) {
	var NS core.CONTROLLER_SESSION_REQUEST
	var AP *core.AccessPoint

	for _, v := range core.GLOBAL_STATE.AccessPoints {
		if Tag == v.Tag {
			AP = v
		}
	}

	NS.UserID = user.ID
	NS.DeviceToken = user.DeviceToken.DT
	NS.GROUP = core.GLOBAL_STATE.ActiveRouter.GROUP
	NS.ROUTERID = core.GLOBAL_STATE.ActiveRouter.ROUTERID
	NS.XGROUP = AP.GROUP
	NS.XROUTERID = AP.ROUTERID
	NS.DEVICEID = AP.DEVICEID

	if core.GLOBAL_STATE.Connected {
		_, code, err := core.Connect(&NS, false)
		if err != nil {
			fmt.Println("There was an error: ", err)
			fmt.Println("Code: ", code)
		}
	} else {
		_, code, err := core.Connect(&NS, true)
		if err != nil {
			fmt.Println("There was an error: ", err)
			fmt.Println("Code: ", code)
		}
	}
}
