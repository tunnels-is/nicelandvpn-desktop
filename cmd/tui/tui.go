package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	stats        []table.Model
	logs         []string
	ready        bool
	status       []string
	keys         keyMap
	help         help.Model
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
		m.serverTable.SetHeight(max(msg.Height-9, 10))

		m.routerTable.SetWidth(max(msg.Width-8, 72))
		m.routerTable.SetHeight(max(msg.Height-9, 10))

		m.logsViewport.Width = max(msg.Width-20, 72)
		m.logsViewport.Height = max(msg.Height-8, 11)

		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			if m.help.ShowAll {
				m.serverTable.SetHeight(m.serverTable.Height() - 5)

				m.routerTable.SetHeight(m.routerTable.Height() - 5)

				m.logsViewport.Height = m.logsViewport.Height - 5
			} else {
				m.serverTable.SetHeight(m.serverTable.Height() + 5)

				m.routerTable.SetHeight(m.routerTable.Height() + 5)

				m.logsViewport.Height = m.logsViewport.Height + 5
			}
			return m, nil
		case key.Matches(msg, m.keys.Disconnect):
			core.Disconnect()
			return m, tea.Println("Disconnected!")
		case key.Matches(msg, m.keys.Quit):
			logout()
			core.CleanupOnClose()
			log.Println("GRACEFULL QUIT!")
			return m, tea.Quit
		case key.Matches(msg, m.keys.Right):
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
		case key.Matches(msg, m.keys.Left):
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
		case key.Matches(msg, m.keys.Select):
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
			}
		}
	default:
		// update the logs and tables
		logs := GetLogs()
		if len(m.logs) != len(logs) {
			m.logs = logs
		}
		var str string
		for _, l := range m.logs {
			str = str + l
		}
		m.logsViewport.SetContent(str)
		// m.logsViewport.SetContent(strings.Join(m.logs, ""))
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

		if m.activeTab == 3 {
			detailedStatsUpdate(m)
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
	case 3: // Breaks if terminal columns < 134 ??? Also if for some reason you change the number the form of the stats your will have to change this too
		tabContent = lipgloss.JoinHorizontal(
			lipgloss.Left, baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, m.stats[0].View(), "\n", m.stats[2].View())),
			baseStyle.Render(lipgloss.JoinVertical(lipgloss.Left, m.stats[1].View(), strings.Repeat("\n", 8), m.stats[3].View())))
	default:
		tabContent = baseStyle.Render("Not implemented yet!")
	}
	doc.WriteString(windowStyle.Render(tabContent))
	doc.WriteString("\n")

	// Status line at the bottom
	var status string
	if core.GLOBAL_STATE.Connected {
		status = "Router: " + core.GLOBAL_STATE.ActiveRouter.Tag + "\tVPN: " + core.GLOBAL_STATE.ActiveAccessPoint.Tag
	} else {
		status = "Router: " + core.GLOBAL_STATE.ActiveRouter.Tag + "\tVPN: Not Connected"
	}

	stats := "Up: " + core.GLOBAL_STATE.UMbpsString + " " + "Down: " + core.GLOBAL_STATE.DMbpsString
	sep := "\t"
	status = lipgloss.JoinHorizontal(lipgloss.Left, status, sep, stats)
	hlpView := m.help.View(m.keys)
	if m.help.ShowAll {
		doc.WriteString(lipgloss.JoinVertical(lipgloss.Left, statusStyle.Render(status), hlpView))
	} else {
		doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, statusStyle.Render(status), sep, hlpView))
	}

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
	tabs := []string{"VPN List", "Router List", "Logs", "Details", "Settings"}

	// Initialize the servers and routers tables
	// Columns for server table
	s_col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 7},
		{Title: "QoS", Width: 3},
	}

	// Columns for routers table
	r_col := []table.Column{
		{Title: "server", Width: 24},
		{Title: "country", Width: 7},
		{Title: "QoS", Width: 3},
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

	detes := detailedStatsInit()
	// Initial tables finished ---

	// Initialize the viewport for the logs
	vp := viewport.New(50, 22)
	vp.SetContent("Loading...")
	vp.Style = baseStyle.UnsetBorderStyle()

	// make the model and give some starting values
	m := model{tabs: tabs, serverTable: s_t, routerTable: r_t, logsViewport: vp, stats: detes}
	m.serverTable.Focus() // focus on the first table since it starts there

	// help & keybinds
	m.help = help.New()
	m.keys = keys

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

// Probably I should refactor detailedStatsUpdate() and detailedStatsInit() too much spaghet
// Updates the stats for the tables in the detailed Stats tab
func detailedStatsUpdate(m model) {
	// The strings for the name of each stat are declared in the globals.go
	app_state_values := [10]string{
		strconv.Itoa(core.GLOBAL_STATE.SecondsUntilAccessPointUpdate), strconv.FormatBool(core.GLOBAL_STATE.ClientReady),
		core.GLOBAL_STATE.Version, strconv.FormatBool(core.GLOBAL_STATE.TunnelInitialized),
		strconv.FormatBool(core.GLOBAL_STATE.IsAdmin), strconv.FormatBool(core.GLOBAL_STATE.ConfigInitialized),
		strconv.FormatBool(core.GLOBAL_STATE.LogFileInitialized), strconv.FormatBool(core.GLOBAL_STATE.BufferError),
		strconv.FormatBool(core.GLOBAL_STATE.BufferError),
	}

	var as_row []table.Row
	for i := 0; i < 10; i++ {
		as_row = append(as_row, table.Row{app_state_str[i], app_state_values[i]})
	}

	interface_values := [3]string{
		core.GLOBAL_STATE.DefaultInterface.IFName,
		strconv.FormatBool(core.GLOBAL_STATE.DefaultInterface.IPV6Enabled),
		core.GLOBAL_STATE.DefaultInterface.DefaultRouter,
	}

	var i_row []table.Row
	for i := 0; i < 3; i++ {
		i_row = append(i_row, table.Row{interface_str[i], interface_values[i]})
	}

	connetion_values := [3]string{
		core.GLOBAL_STATE.ActiveRouter.Tag,
		strconv.FormatUint(core.GLOBAL_STATE.ActiveRouter.MS, 10),
		strconv.Itoa(core.GLOBAL_STATE.ActiveRouter.Score),
	}

	var c_row []table.Row
	for i := 0; i < 3; i++ {
		c_row = append(c_row, table.Row{connection_str[i], connetion_values[i]})
	}

	network_stats_values := [5]string{
		strconv.FormatBool(core.GLOBAL_STATE.Connected),
		core.GLOBAL_STATE.DMbpsString,
		strconv.FormatUint(core.GLOBAL_STATE.IngressPackets, 10),
		core.GLOBAL_STATE.UMbpsString,
		strconv.FormatUint(core.GLOBAL_STATE.EgressPackets, 10),
	}

	var n_row []table.Row
	for i := 0; i < 5; i++ {
		n_row = append(n_row, table.Row{network_stats_str[i], network_stats_values[i]})
	}

	m.stats[0].SetRows(as_row)
	m.stats[1].SetRows(i_row)
	m.stats[2].SetRows(c_row)
	m.stats[3].SetRows(n_row)
}

// Initializes the tables for the Detailed Stats tab
func detailedStatsInit() []table.Model {
	// The strings for the name of each stat are declared in the globals.go

	// App State table
	as_col := []table.Column{
		{Title: "App State", Width: 19},
		{Title: " ", Width: 6},
	}

	app_state_values := [10]string{
		strconv.Itoa(core.GLOBAL_STATE.SecondsUntilAccessPointUpdate), strconv.FormatBool(core.GLOBAL_STATE.ClientReady),
		core.GLOBAL_STATE.Version, strconv.FormatBool(core.GLOBAL_STATE.TunnelInitialized),
		strconv.FormatBool(core.GLOBAL_STATE.IsAdmin), strconv.FormatBool(core.GLOBAL_STATE.ConfigInitialized),
		strconv.FormatBool(core.GLOBAL_STATE.LogFileInitialized), strconv.FormatBool(core.GLOBAL_STATE.BufferError),
		strconv.FormatBool(core.GLOBAL_STATE.BufferError),
	}

	var as_row []table.Row
	for i := 0; i < 10; i++ {
		as_row = append(as_row, table.Row{app_state_str[i], app_state_values[i]})
	}

	as_t := table.New(
		table.WithColumns(as_col),
		table.WithRows(as_row),
		table.WithHeight(10),
	)

	as_t.SetStyles(detailedStatsStyle)
	as_t.Blur()

	// Interface table
	i_col := []table.Column{
		{Title: "Interface", Width: 12},
		{Title: " ", Width: 15},
	}

	interface_values := [3]string{
		core.GLOBAL_STATE.DefaultInterface.IFName,
		strconv.FormatBool(core.GLOBAL_STATE.DefaultInterface.IPV6Enabled),
		core.GLOBAL_STATE.DefaultInterface.DefaultRouter,
	}

	var i_row []table.Row
	for i := 0; i < 3; i++ {
		i_row = append(i_row, table.Row{interface_str[i], interface_values[i]})
	}

	i_t := table.New(
		table.WithColumns(i_col),
		table.WithRows(i_row),
		table.WithHeight(3),
	)

	i_t.SetStyles(detailedStatsStyle)
	i_t.Blur()

	// Connection table
	c_col := []table.Column{
		{Title: "Connection", Width: 12},
		{Title: " ", Width: 12},
	}

	connetion_values := [3]string{
		core.GLOBAL_STATE.ActiveRouter.Tag,
		strconv.FormatUint(core.GLOBAL_STATE.ActiveRouter.MS, 10),
		strconv.Itoa(core.GLOBAL_STATE.ActiveRouter.Score),
	}

	var c_row []table.Row
	for i := 0; i < 3; i++ {
		c_row = append(c_row, table.Row{connection_str[i], connetion_values[i]})
	}

	c_t := table.New(
		table.WithColumns(c_col),
		table.WithRows(c_row),
		table.WithHeight(3),
	)

	c_t.SetStyles(detailedStatsStyle)
	c_t.Blur()

	// Network Stats table
	n_col := []table.Column{
		{Title: "Network Stats", Width: 13},
		{Title: " ", Width: 7},
	}

	network_stats_values := [5]string{
		strconv.FormatBool(core.GLOBAL_STATE.Connected),
		core.GLOBAL_STATE.DMbpsString,
		strconv.FormatUint(core.GLOBAL_STATE.IngressPackets, 10),
		core.GLOBAL_STATE.UMbpsString,
		strconv.FormatUint(core.GLOBAL_STATE.EgressPackets, 10),
	}

	var n_row []table.Row
	for i := 0; i < 5; i++ {
		n_row = append(n_row, table.Row{network_stats_str[i], network_stats_values[i]})
	}

	n_t := table.New(
		table.WithColumns(n_col),
		table.WithRows(n_row),
		table.WithHeight(5),
	)

	n_t.SetStyles(detailedStatsStyle)
	n_t.Blur()

	return []table.Model{as_t, i_t, c_t, n_t}
}
