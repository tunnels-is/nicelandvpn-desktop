package termlib

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)

type model struct {
	spinner spinner.Model
	Program *tea.Program
}

func NewSpinner() (m *model) {
	m = new(model)
	return m
}

func (m *model) Start() {
	m.ResetSpinner()
	m.Program = tea.NewProgram(m)

	if _, err := m.Program.Run(); err != nil {
		fmt.Println("could not run program:", err)
	}
}

func (m *model) Stop() {
	m.Program.Quit()
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.QuitMsg:
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m *model) ResetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[6]
}

func (m model) View() (s string) {
	s += fmt.Sprintf("\n %s%s%s\n\n", m.spinner.View(), " ", textStyle("Spinning..."))
	return
}
