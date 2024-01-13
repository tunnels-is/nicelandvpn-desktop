//go:build freebsd || linux || openbsd

package core

import (
	"errors"
	"os/exec"

	"github.com/zveinn/tunnels"
)

func CB_CreateNewTunnelInterface(
	name string,
	address string,
	netmask string,
	txQueueLen int32,
	mtu int32,
	persistent bool,
) (OUTIF *TunInterface, err error) {
	defer RecoverAndLogToFile()

	OUTIF = new(TunInterface)
	OUTIF.LINUX_IF = &tunnels.Interface{
		Name:        name,
		IPv4Address: address,
		NetMask:     netmask,
		TxQueuelen:  txQueueLen,
		MTU:         mtu,
		Persistent:  persistent,
	}

	err = OUTIF.LINUX_IF.Create()
	if err != nil {
		return
	}

	OUTIF.Read = OUTIF.LINUX_IF.RWC.Read
	OUTIF.Write = OUTIF.LINUX_IF.RWC.Write
	OUTIF.Close = OUTIF.LINUX_IF.RWC.Close
	OUTIF.Addr = OUTIF.LINUX_IF.Syscall_Addr
	OUTIF.Up = OUTIF.LINUX_IF.Syscall_UP
	OUTIF.Down = OUTIF.LINUX_IF.Syscall_DOWN
	OUTIF.MTU = OUTIF.LINUX_IF.Syscall_MTU
	OUTIF.TXQueueLen = OUTIF.LINUX_IF.Syscall_TXQueuelen
	OUTIF.Netmask = OUTIF.LINUX_IF.Syscall_NetMask
	OUTIF.Delete = OUTIF.LINUX_IF.Syscall_Delete

	OUTIF.PreConnect = func(_ *VPNConnectionMETA) (err error) {
		if err = OUTIF.LINUX_IF.Syscall_Addr(); err != nil {
			return
		}
		if err = OUTIF.LINUX_IF.Syscall_MTU(); err != nil {
			return
		}
		if err = OUTIF.LINUX_IF.Syscall_TXQueuelen(); err != nil {
			return
		}
		if err = OUTIF.LINUX_IF.Syscall_UP(); err != nil {
			return
		}

		return
	}

	OUTIF.Connect = func(V *VPNConnectionMETA) (err error) {
		// CHANGE DNS ?? (only on windows)

		for _, n := range V.Networks {
			for _, v := range n.Routes {
				var out []byte
				out, err = exec.Command("ip", "route", "add", v.Address, "via", OUTIF.LINUX_IF.IPv4Address, "dev", OUTIF.LINUX_IF.Name, "metric", v.Metric).CombinedOutput()
				if err != nil {
					return errors.New("err: " + err.Error() + " || out: " + string(out))
				}
			}
		}

		return
	}

	OUTIF.Disconnect = func(_ *VPNConnectionMETA) (err error) {
		if !OUTIF.LINUX_IF.Persistent {
			err = OUTIF.Delete()
			if err != nil {
				CreateErrorLog("disconnect", "unable to delete the interface", err)
			}
		} else {
			err = OUTIF.Down()
			if err != nil {
				CreateErrorLog("disconnect", "unable to bring the interface down", err)
			}
		}

		return
	}

	return
}
