package service

import (
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

const (
	VERSION          = "1.1.5"
	PRODUCTION       = false
	ENABLE_INTERFACE = true
)

var routineMonitor = make(chan int, 200)

func Start() {
	defer func() {
		if r := recover(); r != nil {
			core.CreateErrorLog("", r, string(debug.Stack()))
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INTERFACE
	core.GLOBAL_STATE.Version = VERSION

	core.StartService(routineMonitor)

	routineMonitor <- 1
	routineMonitor <- 2
	routineMonitor <- 3
	routineMonitor <- 4
	routineMonitor <- 5
	routineMonitor <- 6
	routineMonitor <- 7
	routineMonitor <- 8
	routineMonitor <- 9

	for {
		select {
		case ID := <-routineMonitor:
			log.Println("ID", ID)
			if ID == 1 {
				go core.StartLogQueueProcessor(routineMonitor)
			} else if ID == 2 {
				go core.START_API(routineMonitor)
			} else if ID == 3 {
				go core.PingAllVPNConnections(routineMonitor)
			} else if ID == 4 {
				go core.GetDefaultGateway(routineMonitor)
			} else if ID == 5 {
			} else if ID == 6 {
			} else if ID == 7 {
			} else if ID == 8 {
				// go core.StateMaintenance(routineMonitor)
				// go core.GetDefaultGateway(routineMonitor)
				// go core.ProbeRouters(routineMonitor)
				// go core.CleanPorts(routineMonitor)
			}
		default:
			// log.Println("default")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
