package main

import (
	"log"
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
	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INTERFACE
	core.GLOBAL_STATE.Version = VERSION
	core.StartService(MONITOR)

	MONITOR <- 7

	for {
		select {
		case ID := <-MONITOR:
			log.Println("ID", ID)
			if ID == 1 {
				go core.StateMaintenance(MONITOR)
			} else if ID == 2 {
				go core.ReadFromRouterSocket(MONITOR)
			} else if ID == 4 {
				go core.ReadFromLocalSocket(MONITOR)
			} else if ID == 6 {
				go core.CalculateBandwidth(MONITOR)
			} else if ID == 8 {
				go core.StartLogQueueProcessor(MONITOR)
			} else if ID == 9 {
				go core.CleanPorts(MONITOR)
			} else if ID == 7 {
				go core.START_API(MONITOR)
			} else if ID == 8 {
				go core.EventAndStateManager()
			}
		default:
			log.Println("default")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
