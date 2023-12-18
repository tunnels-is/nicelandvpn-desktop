package core

import (
	"bytes"
	"net"
	"time"

	"github.com/zveinn/tunnels"
)

var (
	DEFAULT_GATEWAY         net.IP
	ROUTER_PROBE_TIMEOUT_MS = 60000
	LAST_ROUTER_PROBE       = time.Now().AddDate(0, 0, -1)
)

type TunInterface struct {
	Read  func([]byte) (int, error)
	Write func([]byte) (int, error)
	Close func() error

	PreConnect func() error
	Connect    func() error
	Disconnect func() error

	// Function to apply new parameters
	Addr       func() error
	Up         func() error
	Down       func() error
	MTU        func() error
	TXQueueLen func() error
	Netmask    func() error
	Delete     func() error

	// DANGER ZONE
	LINUX_IF *tunnels.Interface

	// TODO
	DARWIN_IF  *Adapter
	WINDOWS_IF *Adapter
}

func PingAllVPNConnections(MONITOR chan int) {
	defer func() {
		MONITOR <- 3
	}()
	defer RecoverAndLogToFile()
}

func GetDefaultGateway(MONITOR chan int) {
	defer func() {
		// MONITOR <- 4
	}()
	defer RecoverAndLogToFile()
	var err error

	var OLD_GATEWAY net.IP
	copy(OLD_GATEWAY, DEFAULT_GATEWAY)

	DEFAULT_GATEWAY, err = tunnels.FindGateway()
	if err != nil {
		CreateErrorLog("connect", "default gateway not found", err)
	}
	if bytes.Compare(OLD_GATEWAY, DEFAULT_GATEWAY) != 0 {
		if time.Since(LAST_ROUTER_PROBE).Milliseconds() > int64(ROUTER_PROBE_TIMEOUT_MS) {
			err := RefreshRouterList()
			if err != nil {
				CreateErrorLog("", "Unable to find the best router for your connection: ", err)
			} else {
				LAST_ROUTER_PROBE = time.Now()
			}
		}
	}

	if DEFAULT_GATEWAY != nil {
		time.Sleep(5 * time.Second)
	} else {
		time.Sleep(2 * time.Second)
	}
}
