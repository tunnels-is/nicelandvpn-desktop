package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
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

	user        *core.User
	output      = termenv.NewOutput(os.Stdout)
	color       = output.ForegroundColor()
	bgcolor     = output.BackgroundColor()
	resetString = output.String("... exiting")

	userLoginInputs = make([]textinput.Model, 4)
)

func main() {
	defer func() {
		// This will reset the forground and background of the terminal when exiting
		resetString.Foreground(color)
		resetString.Background(bgcolor)
		fmt.Print("\033[H\033[2J")
		fmt.Print("\033[H\033[2J")
		fmt.Print("\033[H\033[2J")
		fmt.Println(resetString)
	}()

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
		GetAPIKey()
	case "createConfig":
		CreateDummyConfig()
	default:
		os.Exit(1)
	}
}

func GetAPIKey() {
	s := termlib.NewSpinner()
	go s.Start()

	core.C = new(core.Config)
	core.C.DebugLogging = true
	core.InitPaths()
	core.CreateBaseFolder()
	core.InitLogfile()
	go core.StartLogQueueProcessor(MONITOR)
	err := core.REF_RefreshRouterList()
	time.Sleep(2 * time.Second)
	s.Stop()
	if err != nil {
		core.CreateErrorLog("", "Unable to find the best router for your connection: ", err)
		os.Exit(1)
	}

	fmt.Print("\033[H\033[2J")
	termlib.Login(userLoginInputs)
	log.Println("USER INPUT:", userLoginInputs[0].Value())
	user = termlib.SendLoginRequest(userLoginInputs)
	if user != nil {
		log.Println("API KEY:", user.APIKey)
	} else {
		log.Println("Invalid login..")
	}
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
				// go core.StateMaintenance(MONITOR)
			} else if ID == 2 {
				// go core.ReadFromRouterSocket(MONITOR)
			} else if ID == 3 {
				// TUI ONLY .. does not fire on wails GUI
				// go TimedUIUpdate(MONITOR)
			} else if ID == 4 {
				// go core.ReadFromLocalSocket(MONITOR)
			} else if ID == 6 {
				// go core.CalculateBandwidth(MONITOR)
			} else if ID == 8 {
				go core.StartLogQueueProcessor(MONITOR)
			}
		}
	}
}
