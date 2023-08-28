package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

// Update the accesspoints, routers and logs every 3 seconds
func TimedUIUpdate(MONITOR chan int) {
	defer func() {
		time.Sleep(3 * time.Second)
		MONITOR <- 3
	}()
	defer core.RecoverAndLogToFile()

	for {
		time.Sleep(3 * time.Second)
    core.GetRoutersAndAccessPoints()
		TUI.Send(&tea.KeyMsg{
			Type: 0,
		})
	}
}

func GetLogs() []string {
	var logs []string
	LR, err := core.GetLogsForCLI()
	if LR != nil && err == nil {
		logs = LR.Content
		for i := range LR.Content {
			if LR.Content[i] != "" {
				logs[i] = fmt.Sprint(LR.Time[i], "||", LR.Function[i], "||", LR.Content[i]+"\n")
			}
		}
	}

	return logs
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
