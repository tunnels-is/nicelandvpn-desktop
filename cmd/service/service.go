package main

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

var MONITOR = make(chan int, 200)

func main() {
	defer func() {
		if r := recover(); r != nil {
			core.CreateErrorLog("", r, string(debug.Stack()))
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INTERFACE
	core.GLOBAL_STATE.Version = VERSION

	core.StartService(MONITOR)

	MONITOR <- 1
	MONITOR <- 2
	MONITOR <- 3
	MONITOR <- 4
	MONITOR <- 5
	MONITOR <- 6
	MONITOR <- 7
	MONITOR <- 8
	MONITOR <- 9

	for {
		select {
		case ID := <-MONITOR:
			log.Println("ID", ID)
			if ID == 1 {
				go core.StartLogQueueProcessor(MONITOR)
			} else if ID == 2 {
				go core.START_API(MONITOR)
			} else if ID == 3 {
				go core.PingAllVPNConnections(MONITOR)
			} else if ID == 4 {
				go core.GetDefaultGateway(MONITOR)
			} else if ID == 5 {
			} else if ID == 6 {
			} else if ID == 7 {
			} else if ID == 8 {
				// go core.StateMaintenance(MONITOR)
				// go core.GetDefaultGateway(MONITOR)
				// go core.ProbeRouters(MONITOR)
				// go core.CleanPorts(MONITOR)
			}
		default:
			// log.Println("default")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
