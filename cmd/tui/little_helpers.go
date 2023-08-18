package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)


func TimedUIUpdate(MONITOR chan int) {
	defer func() {
		time.Sleep(3 * time.Second)
		MONITOR <- 3
	}()
	defer core.RecoverAndLogToFile()

	for {
		time.Sleep(3 * time.Second)
		TUI.Send(&tea.KeyMsg{
			Type: 0,
		})
	}
}

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
