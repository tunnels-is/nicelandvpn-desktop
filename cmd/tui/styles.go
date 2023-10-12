package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// colors taken from frontend/src/assets/style/variables.scss
var (
	main_bg      = lipgloss.Color("#141414")
	body_bg      = lipgloss.Color("#202324")
	body_dark_bg = lipgloss.Color("#0A0B0E")

	teal        = lipgloss.Color("#28ad85")
	teal_border = lipgloss.Color("#3AF4BD")
	teal_hover  = lipgloss.Color("#20C997")

	orange        = lipgloss.Color("#FF922D")
	orange_border = lipgloss.Color("#EF7503")
	orange_hover  = lipgloss.Color("#EF7503")

	log_error   = lipgloss.Color("#FF0000")
	log_warning = lipgloss.Color("#FFFF00")

	lightblue = lipgloss.Color("#20bec9")
	red       = lipgloss.Color("#FF5858")
	white     = lipgloss.Color("#FFFFFF")
	black     = lipgloss.Color("#000000")

	success_color = lipgloss.Color("#0AB60A")
	error_color   = lipgloss.Color("#E70808")
)

// tab style
var (
	docStyle         = lipgloss.NewStyle().Padding(1, 1, 1, 1)
	highlightColor   = teal_hover
	selectionColor   = orange
	inactiveTabStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, false, true).UnsetBorderBottom().BorderForeground(highlightColor)
	activeTabStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, false, true).UnsetBorderBottom().BorderForeground(highlightColor).Background(selectionColor).Foreground(black)
	windowStyle      = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(0).Border(lipgloss.NormalBorder())
)

// generic content style
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(white).
	Padding(0, 1)

// table style
var table_style = table.Styles{
	Header:   lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(teal).BorderBottom(true).Bold(true).Width(26),
	Selected: lipgloss.NewStyle().Foreground(black).Background(teal_border).Bold(true),
	Cell:     lipgloss.NewStyle().Padding(0, 1).Width(26),
}

// login from styles
var (
	focusedStyle        = lipgloss.NewStyle().Foreground(orange)
	blurredStyle        = lipgloss.NewStyle().Foreground(teal)
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(teal_hover)

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

// status line
var statusStyle = lipgloss.NewStyle().Foreground(orange_hover).Padding(0, 1).Bold(true)
var statsStyle = lipgloss.NewStyle().Foreground(orange_hover).Padding(0).Bold(true)

// Stats table style
var detailedStatsStyle = table.Styles{
	Header:   lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(teal).BorderBottom(true).Bold(true).Width(28).Foreground(teal),
	Selected: lipgloss.NewStyle().Foreground(white),
	Cell:     lipgloss.NewStyle().Padding(0, 1).Width(28),
}
