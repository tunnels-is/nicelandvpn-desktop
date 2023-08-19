package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

const VERSION = "1.1.1"
const PRODUCTION = false
const ENABLE_INSTERFACE = false

var MONITOR = make(chan int, 200)
var TUI *tea.Program

func main() {

	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INSTERFACE
	core.GLOBAL_STATE.Version = VERSION

	go RoutineMonitor()
	go core.StartService(MONITOR)
	go TimedUIUpdate(MONITOR)

	TUI = tea.NewProgram(initialModel())
	if _, err := TUI.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
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
				go core.ReadFromLocalTunnel(MONITOR)
			} else if ID == 6 {
				go core.CalculateBandwidth(MONITOR)
			} else if ID == 8 {
				go core.StartLogQueueProcessor(MONITOR)
			}
		}
	}

}
