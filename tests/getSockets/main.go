package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

const (
	MIB_TCP_TABLE_OWNER_PID_ALL = 5
	MIB_TCP_STATE_DELETE_TCB    = 12
	ERROR_INSUFFICIENT_BUFFER   = 122
)

type MIB_UDPTABLE_OWNER_PID struct {
	dwNumEntries uint32
	table        [30000]MIB_UDPROW_OWNER_PID
}

type MIB_UDPROW_OWNER_PID struct {
	dwLocalAddr uint32
	dwLocalPort uint32
	dwOwningPid uint32
}

type MIB_TCPROW_OWNER_PID struct {
	dwState      uint32
	dwLocalAddr  uint32
	dwLocalPort  uint32
	dwRemoteAddr uint32
	dwRemotePort uint32
	dwOwningPid  uint32
}

type MIB_TCPTABLE_OWNER_PID struct {
	dwNumEntries uint32
	table        [30000]MIB_TCPROW_OWNER_PID
}

var (
	modIphlpapi = syscall.NewLazyDLL("iphlpapi.dll")

	GetTCP = modIphlpapi.NewProc("GetExtendedTcpTable")
	GetUDP = modIphlpapi.NewProc("GetExtendedUdpTable")
	SetTCP = modIphlpapi.NewProc("SetTcpEntry")
)

func getExtendedUdpTable() ([]MIB_UDPROW_OWNER_PID, error) {
	var tcpTable MIB_UDPTABLE_OWNER_PID
	size := uintptr(unsafe.Sizeof(tcpTable))

	r, _, err := GetUDP.Call(
		uintptr(unsafe.Pointer(&tcpTable)),
		uintptr(unsafe.Pointer(&size)),
		1,
		syscall.AF_INET,
		1,
		0)

	// log.Println(r, err, tcpTable)
	for _, row := range tcpTable.table[:tcpTable.dwNumEntries] {
		localAddr := fmt.Sprintf("%d.%d.%d.%d", byte(row.dwLocalAddr), byte(row.dwLocalAddr>>8), byte(row.dwLocalAddr>>16), byte(row.dwLocalAddr>>24))

		if row.dwOwningPid == 4 {
			fmt.Printf("Local Address: %s:%d, PID: %d\n", localAddr, row.dwLocalPort, row.dwOwningPid)

		}

	}

	if r == 0 {
		entries := tcpTable.dwNumEntries
		return tcpTable.table[:entries], nil
	}

	return nil, err
}

func getExtendedTcpTable() ([]MIB_TCPROW_OWNER_PID, error) {
	var tcpTable MIB_TCPTABLE_OWNER_PID
	size := uintptr(unsafe.Sizeof(tcpTable))

	r, _, err := GetTCP.Call(
		uintptr(unsafe.Pointer(&tcpTable)),
		uintptr(unsafe.Pointer(&size)),
		1,
		syscall.AF_INET,
		MIB_TCP_TABLE_OWNER_PID_ALL,
		0)

	for _, row := range tcpTable.table[:tcpTable.dwNumEntries] {
		localAddr := fmt.Sprintf("%d.%d.%d.%d", byte(row.dwLocalAddr), byte(row.dwLocalAddr>>8), byte(row.dwLocalAddr>>16), byte(row.dwLocalAddr>>24))
		remoteAddr := fmt.Sprintf("%d.%d.%d.%d", byte(row.dwRemoteAddr), byte(row.dwRemoteAddr>>8), byte(row.dwRemoteAddr>>16), byte(row.dwRemoteAddr>>24))

		// if localAddr == "192.168.2.10" {
		// if localAddr != "0.0.0.0" && localAddr != "127.0.0.1" {
		// }
		if row.dwOwningPid == 4 {
			fmt.Printf("Local Address: %s:%d, Remote Address: %s:%d, PID: %d\n", localAddr, row.dwLocalPort, remoteAddr, row.dwRemotePort, row.dwOwningPid)

			// setTcpEntry(row)
		}

	}

	if r == 0 {
		entries := tcpTable.dwNumEntries
		return tcpTable.table[:entries], nil
	}

	return nil, err
}

func main() {
	_, err := getExtendedUdpTable()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

}

func setTcpEntry(row MIB_TCPROW_OWNER_PID) error {
	row.dwState = 12
	r, r1, err := SetTCP.Call(
		uintptr(unsafe.Pointer(&row)))

	log.Println(r, r1, err)

	if r == 0 {
		return nil
	}

	return err
}

func getHostname() {
	DLL := syscall.NewLazyDLL("Ws2_32.dll")
	DLL.Load()

	var host = make([]byte, 100)
	lenx := len(host)

	proc2 := DLL.NewProc("gethostname")
	r1, r2, err := syscall.SyscallN(
		proc2.Addr(),
		uintptr(unsafe.Pointer(&host[0])),
		uintptr(unsafe.Pointer(&lenx)),
	)

	log.Println(r1, r2, err)
	log.Println(string(host))
}
