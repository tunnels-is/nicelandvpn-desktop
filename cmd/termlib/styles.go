package termlib

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// colors taken from frontend/src/assets/style/variables.scss
var (
	MainBg     = lipgloss.Color("#141414")
	BodyBg     = lipgloss.Color("#202324")
	BodyDarkBg = lipgloss.Color("#0A0B0E")

	Teal       = lipgloss.Color("#28ad85")
	TealBorder = lipgloss.Color("#3AF4BD")
	TealHover  = lipgloss.Color("#20C997")

	Orange       = lipgloss.Color("#FF922D")
	OrangeBorder = lipgloss.Color("#EF7503")
	OrangeHover  = lipgloss.Color("#EF7503")

	LogError   = lipgloss.Color("#FF0000")
	LogWarning = lipgloss.Color("#FFFF00")

	Lightblue = lipgloss.Color("#20bec9")
	Red       = lipgloss.Color("#FF5858")
	White     = lipgloss.Color("#FFFFFF")
	Black     = lipgloss.Color("#000000")

	SuccessColor = lipgloss.Color("#0AB60A")
	ErrorColor   = lipgloss.Color("#E70808")
)

// tab style
var (
	DocStyle         = lipgloss.NewStyle().Padding(1, 1, 1, 1)
	HighlightColor   = TealHover
	SelectionColor   = Orange
	InactiveTabStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, false, true).UnsetBorderBottom().BorderForeground(HighlightColor)
	ActiveTabStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true, true, false, true).UnsetBorderBottom().BorderForeground(HighlightColor).Background(SelectionColor).Foreground(Black)
	WindowStyle      = lipgloss.NewStyle().BorderForeground(HighlightColor).Padding(0).Border(lipgloss.NormalBorder())
)

// generic content style
var BaseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(White).
	Padding(0, 1)

// table style
var TableStyle = table.Styles{
	Header:   lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(Teal).BorderBottom(true).Bold(true).Width(26),
	Selected: lipgloss.NewStyle().Foreground(Black).Background(TealBorder).Bold(true),
	Cell:     lipgloss.NewStyle().Padding(0, 1).Width(26),
}

// login from styles
var (
	FocusedStyle        = lipgloss.NewStyle().Foreground(Orange)
	BlurredStyle        = lipgloss.NewStyle().Foreground(Teal)
	CursorStyle         = FocusedStyle.Copy()
	NoStyle             = lipgloss.NewStyle()
	HelpStyle           = BlurredStyle.Copy()
	CursorModeHelpStyle = lipgloss.NewStyle().Foreground(TealHover)

	FocusedButton = FocusedStyle.Copy().Render("[ Submit ]")
	BlurredButton = fmt.Sprintf("[ %s ]", BlurredStyle.Render("Submit"))
)

// status line
var StatusStyle = lipgloss.NewStyle().Foreground(OrangeHover).Padding(0, 1).Bold(true)
var StatsStyle = lipgloss.NewStyle().Foreground(OrangeHover).Padding(0).Bold(true)

// Stats table style
var DetailedStatsStyle = table.Styles{
	Header:   lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(Teal).BorderBottom(true).Bold(true).Width(28).Foreground(Teal),
	Selected: lipgloss.NewStyle().Foreground(White),
	Cell:     lipgloss.NewStyle().Padding(0, 1).Width(28),
}
