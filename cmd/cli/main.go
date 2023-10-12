package main

import (
	"log"
	"os"
	"runtime/debug"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tunnels-is/nicelandvpn-desktop/cmd/termlib"
	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

const (
	VERSION           = "1.1.3"
	PRODUCTION        = true
	ENABLE_INSTERFACE = true
)

var (
	FLAG_COMMAND  string
	FLAG_USER     string
	FLAG_PASSWORD string
	MONITOR       = make(chan int, 200)
	TUI           *tea.Program
	user          *core.User
)

func main() {
	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INSTERFACE
	core.GLOBAL_STATE.Version = VERSION

	// log.Println(os.Args)
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	switch os.Args[1] {
	case "connect":
		Connect()
	case "getApiKey":
		GetAPIKEy()
	case "createConfig":
		CreateDummyConfig()
	default:
		os.Exit(1)
	}

}

func GetAPIKEy() {
	log.Println("LOADING ROUTER LIST AND STUFF ...")
	core.C = new(core.Config)
	core.C.DebugLogging = true
	core.InitPaths()
	core.CreateBaseFolder()
	core.InitLogfile()
	go core.StartLogQueueProcessor(MONITOR)
	err := core.RefreshRouterList()
	if err != nil {
		core.CreateErrorLog("", "Unable to find the best router for your connection: ", err)
	}

	user = termlib.Login()
	log.Println("API KEY:", user.APIKey)
}

func CreateDummyConfig() {

}
func Connect() {

	go core.StartService(MONITOR)
	RoutineMonitor()
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
				// go TimedUIUpdate(MONITOR)
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
