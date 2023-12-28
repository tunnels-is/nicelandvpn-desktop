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

func GeneralRoutine(MONITOR chan int) {
	defer func() {
		time.Sleep(5 * time.Second)
		MONITOR <- 5
	}()
	defer RecoverAndLogToFile()

	if DEFAULT_GATEWAY == nil {
		return
	}
	if GLOBAL_STATE.ActiveRouter != nil {
		return
	}

	err := REF_RefreshRouterList()
	if err != nil {
		CreateErrorLog("", "unable to update the router list", err)
	}
}

func PingAllVPNConnections(MONITOR chan int) {
	defer func() {
		time.Sleep(9 * time.Second)
		MONITOR <- 3
	}()
	defer RecoverAndLogToFile()
}

func getDefaultGateway() {
	defer RecoverAndLogToFile()
	var err error

	OLD_GATEWAY := make([]byte, 4)
	copy(OLD_GATEWAY, DEFAULT_GATEWAY)

	DEFAULT_GATEWAY, err = tunnels.FindGateway()
	if err != nil {
		DEFAULT_GATEWAY = nil
		CreateErrorLog("", "default gateway not found", err)
	}

	if !bytes.Equal(OLD_GATEWAY, DEFAULT_GATEWAY) {
		CreateLog("", "new gateway discovered", DEFAULT_GATEWAY)
		err = REF_RefreshRouterList()
		if err != nil {
			CreateErrorLog("", "unable to update the router list", err)
		}
	}
}

func GetDefaultGateway(MONITOR chan int) {
	defer func() {
		if DEFAULT_GATEWAY != nil {
			time.Sleep(5 * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}

		MONITOR <- 4
	}()
	getDefaultGateway()
}
