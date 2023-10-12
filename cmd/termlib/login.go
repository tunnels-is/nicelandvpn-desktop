package termlib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

type loginForm struct {
	focusIndex int
	inputs     []textinput.Model
}

func intialModel() loginForm {
	m := loginForm{
		inputs: make([]textinput.Model, 4),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = CursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Email"
			t.Focus()
			t.PromptStyle = FocusedStyle
			t.TextStyle = FocusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '*'
		case 2:
			t.Placeholder = "Device Name"
			t.CharLimit = 64
		case 3:
			t.Placeholder = "2FA Code if enabled"
		}
		m.inputs[i] = t
	}

	return m
}

func (m loginForm) Init() tea.Cmd {
	return textinput.Blink
}

func (m loginForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// core.CleanupOnClose()
			sendLoginRequest(m.inputs) // I guess this is one way to exit without going into the TUI
			return m, tea.Quit
		case "tab", "shift-tab", "enter", "up", "down":
			s := msg.String()

			// if hit enter while the submit button was focused
			if s == "enter" && m.focusIndex == len(m.inputs) {
				sendLoginRequest(m.inputs)
				return m, tea.Quit
			}

			// cycle indexes
			if s == "up" || s == "shift-tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = FocusedStyle
					m.inputs[i].TextStyle = FocusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = NoStyle
				m.inputs[i].TextStyle = NoStyle
			}
			return m, tea.Batch(cmds...)
		}
	}
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *loginForm) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m loginForm) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &BlurredButton
	if m.focusIndex == len(m.inputs) {
		button = &FocusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

var user *core.User

func Login() (u *core.User) {
	_, err := tea.NewProgram(intialModel()).Run()
	if err != nil {
		fmt.Printf("Could not start the login form: %s\n", err)
		core.CleanupOnClose()
		os.Exit(1)
	}
	u = user
	return
}

func sendLoginRequest(creds []textinput.Model) {
	var FR core.FORWARD_REQUEST

	// fill the login form
	li := core.LoginForm{
		Email:      creds[0].Value(),
		Password:   creds[1].Value(),
		Digits:     creds[3].Value(),
		DeviceName: creds[2].Value(),
	}

	// construct the request
	FR.Path = "v2/user/login"
	FR.Method = "POST"
	FR.Timeout = 20000
	FR.JSONData = li

	// send the request with the creds
	respBytes, code, err := core.SendRequestToControllerProxy(FR.Method, FR.Path, FR.JSONData, "api.atodoslist.net", FR.Timeout)
	if err != nil || code != 200 {
		fmt.Println("\nCode: ", code)
		fmt.Println("Log in error: ", err)
		core.CleanupOnClose()
		os.Exit(1)
	}

	// unfold it in the user global
	err = json.Unmarshal(respBytes, &user)
	if err != nil {
		fmt.Println("Response error: ", err)
		core.CleanupOnClose()
		os.Exit(1)
	}

}
