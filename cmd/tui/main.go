package main

import (
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

const (
	VERSION           = "1.1.3"
	PRODUCTION        = true
	ENABLE_INSTERFACE = true
)

var (
	MONITOR = make(chan int, 200)
	TUI     *tea.Program
)

func main() {
	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INSTERFACE
	core.GLOBAL_STATE.Version = VERSION

	go RoutineMonitor()
	go core.StartService(MONITOR)

	StartTui()
}

func RoutineMonitor() {
	defer func() {
		if r := recover(); r != nil {
			core.CreateErrorLog("", r, string(debug.Stack()))
			go RoutineMonitor()
		}
	}()

	for {
		select {
		default:
			time.Sleep(500 * time.Millisecond)
		case ID := <-MONITOR:
			if ID == 1 {
				go core.StateMaintenance(MONITOR)
			} else if ID == 2 {
				go core.ReadFromRouterSocket(MONITOR)
			} else if ID == 3 {
				// TUI ONLY .. does not fire on wails GUI
				go TimedUIUpdate(MONITOR)
			} else if ID == 4 {
				go core.ReadFromLocalSocket(MONITOR)
			} else if ID == 6 {
				go core.CalculateBandwidth(MONITOR)
			} else if ID == 8 {
				go core.StartLogQueueProcessor(MONITOR)
			}
		}
	}
}
