package main

import (
	"log"
	"time"

	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

const VERSION = "1.1.1"
const PRODUCTION = true
const ENABLE_INSTERFACE = true

var MONITOR = make(chan int, 200)

func main() {

	core.PRODUCTION = PRODUCTION
	core.ENABLE_INSTERFACE = ENABLE_INSTERFACE
	core.GLOBAL_STATE.Version = VERSION

	go core.StartService(MONITOR)

	for {
		time.Sleep(1 * time.Second)
		log.Println("TUI GOES HERE")
	}
}
