//go:build windows

package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ===========================================================
// ===========================================================
// GLOBALS
// ===========================================================
// ===========================================================

var (
	modwintun               = newLazyDLL("./wintun.dll", setupLogger)
	procWintunCreateAdapter = modwintun.NewProc("WintunCreateAdapter")
	procWintunOpenAdapter   = modwintun.NewProc("WintunOpenAdapter")
	procWintunCloseAdapter  = modwintun.NewProc("WintunCloseAdapter")
	procWintunDeleteDriver  = modwintun.NewProc("WintunDeleteDriver")
	// procWintunGetAdapterLUID          = modwintun.NewProc("WintunGetAdapterLUID")
	procWintunGetRunningDriverVersion = modwintun.NewProc("WintunGetRunningDriverVersion")
	procWintunAllocateSendPacket      = modwintun.NewProc("WintunAllocateSendPacket")
	procWintunEndSession              = modwintun.NewProc("WintunEndSession")
	// procWintunGetReadWaitEvent        = modwintun.NewProc("WintunGetReadWaitEvent")
	procWintunReceivePacket        = modwintun.NewProc("WintunReceivePacket")
	procWintunReleaseReceivePacket = modwintun.NewProc("WintunReleaseReceivePacket")
	procWintunSendPacket           = modwintun.NewProc("WintunSendPacket")
	procWintunStartSession         = modwintun.NewProc("WintunStartSession")
	procWintunGetLastError         = modwintun.NewProc("WintunGetLastError")
	// procWintunSetLogger               = modwintun.NewProc("WintunSetLogger")

	GUID *windows.GUID
	Dev  Device

	// WINDOWS DLL
	IPHLPApi = syscall.NewLazyDLL("iphlpapi.dll")

	GetTCP = IPHLPApi.NewProc("GetExtendedTcpTable")
	GetUDP = IPHLPApi.NewProc("GetExtendedUdpTable")
	SetTCP = IPHLPApi.NewProc("SetTcpEntry")
)

type loggerLevel int

const (
	PacketSizeMax   = 0xffff    // Maximum packet size
	RingCapacityMin = 0x20000   // Minimum ring capacity (128 kiB)
	RingCapacityMax = 0x4000000 // Maximum ring capacity (64 MiB)
	AdapterNameMax  = 128

	// WINDOWS DLL

	MIB_TCP_TABLE_OWNER_PID_ALL = 5
	MIB_TCP_STATE_DELETE_TCB    = 12
)

// ===========================================================
// ===========================================================
// DLL CODE
// ===========================================================
// ===========================================================
func newLazyDLL(name string, onLoad func(d *lazyDLL)) *lazyDLL {
	return &lazyDLL{Name: name, onLoad: onLoad}
}

func (d *lazyDLL) NewProc(name string) *lazyProc {
	return &lazyProc{dll: d, Name: name}
}

func (p *lazyProc) Find() error {
	if atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&p.addr))) != nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.addr != 0 {
		return nil
	}

	err := p.dll.Load()
	if err != nil {
		return fmt.Errorf("error loading %v DLL: %w", p.dll.Name, err)
	}
	addr, err := p.nameToAddr()
	if err != nil {
		return fmt.Errorf("error getting %v address: %w", p.Name, err)
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&p.addr)), unsafe.Pointer(addr))
	return nil
}

func (p *lazyProc) Addr() uintptr {
	// log.Println("FINDING DLL ADDS", p, " >> ADDR:", p.addr)
	err := p.Find()
	if err != nil {
		panic(err)
	}
	return p.addr
}

func (d *lazyDLL) Load() error {
	if atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&d.module))) != nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.module != 0 {
		return nil
	}

	const (
		LOAD_LIBRARY_SEARCH_APPLICATION_DIR = 0x00000200
		LOAD_LIBRARY_SEARCH_SYSTEM32        = 0x00000800
	)
	module, err := windows.LoadLibraryEx(d.Name, 0, LOAD_LIBRARY_SEARCH_APPLICATION_DIR|LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return fmt.Errorf("unable to load library: %w", err)
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&d.module)), unsafe.Pointer(module))
	if d.onLoad != nil {
		d.onLoad(d)
	}
	return nil
}

func (p *lazyProc) nameToAddr() (uintptr, error) {
	return windows.GetProcAddress(p.dll.module, p.Name)
}

// ===========================================================
// ===========================================================
// STRUCTS and interfaces
// ===========================================================
// ===========================================================

type AdapterHandle struct {
	handle uintptr
}

type SessionHandle struct {
	handle uintptr
}

type WinHandle struct {
	// handle uintptr
}

type Adapter struct {
	Initialized bool

	NameU16 *uint16
	TypeU16 *uint16

	// wt == AdapterHandle
	AdapterHandle AdapterHandle
	//
	GUID         *windows.GUID
	TunnelHandle WinHandle

	Name          string
	TunHandle     windows.Handle
	SessionHandle SessionHandle
	// readWait  windows.Handle
	Events chan Event
	// running sync.WaitGroup
	// closeOnce sync.Once
	// close     int32
	ForcedMTU int
}

type lazyProc struct {
	Name string
	mu   sync.Mutex
	dll  *lazyDLL
	addr uintptr
}
type lazyDLL struct {
	Name   string
	mu     sync.Mutex
	module windows.Handle
	onLoad func(d *lazyDLL)
}
type Event int

// Packet with data
type Packet struct {
	Next *Packet              // Pointer to next packet in queue
	Size uint32               // Size of packet (max WINTUN_MAX_IP_PACKET_SIZE)
	Data *[PacketSizeMax]byte // Pointer to layer 3 IPv4 or IPv6 packet
}

type TimestampedWriter interface {
	WriteWithTimestamp(p []byte, ts int64) (n int, err error)
}

type Device interface {
	File() *os.File                 // returns the file descriptor of the device
	Read([]byte, int) (int, error)  // read a packet from the device (without any additional headers)
	Write([]byte, int) (int, error) // writes a packet to the device (without any additional headers)
	Flush() error                   // flush all previous writes to the device
	MTU() (int, error)              // returns the MTU of the device
	Name() (string, error)          // fetches and returns the current name
	Events() chan Event             // returns a constant channel of events related to the device
	Close() error                   // stops the device and closes the event channel
}

// ===========================================================
// ===========================================================
// WG ADAPTER CODE
// ===========================================================
// ===========================================================

func (A *Adapter) Close() (err error) {
	defer RecoverAndLogToFile()

	CreateLog("", "Closing Adapter")
	runtime.SetFinalizer(&A.AdapterHandle, AdapterCleanup)
	r1, _, e1 := syscall.SyscallN(procWintunEndSession.Addr(), A.SessionHandle.handle)
	if r1 == 0 {
		err = e1
		// log.Println(err)
	}

	r1, _, e1 = syscall.SyscallN(procWintunCloseAdapter.Addr(), A.AdapterHandle.handle)
	if r1 == 0 {
		err = e1
		// log.Println(err)
	}
	return
}

func (A *Adapter) Start(cap uint32) (err error) {

	r1, _, err1 := syscall.SyscallN(procWintunStartSession.Addr(), uintptr(A.AdapterHandle.handle), uintptr(cap))
	if r1 == 0 {
		err = err1
		return
	}

	A.SessionHandle = SessionHandle{handle: r1}
	return
}

func (A Adapter) AllocateSendPacket(packetSize int) (packet []byte, err error) {
	r0, _, e1 := syscall.SyscallN(procWintunAllocateSendPacket.Addr(), A.SessionHandle.handle, uintptr(packetSize))
	if r0 == 0 {
		err = e1
		return
	}
	packet = unsafe.Slice((*byte)(unsafe.Pointer(r0)), packetSize)
	return
}

func (A *Adapter) SendPacket(packet []byte) {
	// syscall.Syscall(procWintunSendPacket.Addr(), 2, A.SessionHandle.handle, uintptr(unsafe.Pointer(&packet[0])), 0)
	syscall.SyscallN(procWintunSendPacket.Addr(), A.SessionHandle.handle, uintptr(unsafe.Pointer(&packet[0])))
}

func (A *Adapter) ReleaseReceivePacket(packet []byte) {
	syscall.SyscallN(procWintunReleaseReceivePacket.Addr(), A.SessionHandle.handle, uintptr(unsafe.Pointer(&packet[0])))
}

func (A *Adapter) ReceivePacket() (packet []byte, size uint16, err error) {
	// var packetSize uint32
	r0, _, e1 := syscall.SyscallN(procWintunReceivePacket.Addr(), A.SessionHandle.handle, uintptr(unsafe.Pointer(&size)))
	if r0 == 0 {
		err = e1
		return
	}
	packet = unsafe.Slice((*byte)(unsafe.Pointer(r0)), size)
	return
}

// Uninstall removes the driver from the system if no drivers are currently in use.
func (*Adapter) Uninstall() (err error) {
	CreateErrorLog("", "Uninstalling Adapter")
	r1, _, e1 := syscall.SyscallN(procWintunDeleteDriver.Addr())
	if r1 == 0 {
		err = e1
	}
	return
}

// RunningVersion returns the version of the loaded driver.
func RunningVersion() (version uint32, err error) {
	r0, _, e1 := syscall.SyscallN(procWintunGetRunningDriverVersion.Addr())
	version = uint32(r0)
	if version == 0 {
		err = e1
	}
	return
}

// ===========================================================
// ===========================================================
// WG LOG CODE
// ===========================================================
// ===========================================================

func logMessage(level loggerLevel, timestamp uint64, msg *uint16) int {
	if !PRODUCTION {
		if tw, ok := log.Default().Writer().(TimestampedWriter); ok {
			tw.WriteWithTimestamp([]byte(log.Default().Prefix()+windows.UTF16PtrToString(msg)), (int64(timestamp)-116444736000000000)*100)
		} else {
			// log.Println(windows.UTF16PtrToString(msg))
		}
	}
	return 0
}

func setupLogger(dll *lazyDLL) {
	defer RecoverAndLogToFile()

	var callback uintptr
	// log.Println("SETTING UP", runtime.GOARCH)

	if runtime.GOARCH == "386" {
		callback = windows.NewCallback(func(level loggerLevel, timestampLow, timestampHigh uint32, msg *uint16) int {
			return logMessage(level, uint64(timestampHigh)<<32|uint64(timestampLow), msg)
		})
	} else if runtime.GOARCH == "arm" {
		callback = windows.NewCallback(func(level loggerLevel, _, timestampLow, timestampHigh uint32, msg *uint16) int {
			return logMessage(level, uint64(timestampHigh)<<32|uint64(timestampLow), msg)
		})
	} else if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
		callback = windows.NewCallback(logMessage)
	}

	syscall.SyscallN(dll.NewProc("WintunSetLogger").Addr(), callback)
}

// Version returns the version of the Wintun DLL.
func Version() string {
	if modwintun.Load() != nil {
		return "unknown"
	}
	resInfo, err := windows.FindResource(modwintun.module, windows.ResourceID(1), windows.RT_VERSION)
	if err != nil {
		return "unknown"
	}
	data, err := windows.LoadResourceData(modwintun.module, resInfo)
	if err != nil {
		return "unknown"
	}

	var fixedInfo *windows.VS_FIXEDFILEINFO
	fixedInfoLen := uint32(unsafe.Sizeof(*fixedInfo))
	err = windows.VerQueryValue(unsafe.Pointer(&data[0]), `\`, unsafe.Pointer(&fixedInfo), &fixedInfoLen)
	if err != nil {
		return "unknown"
	}
	version := fmt.Sprintf("%d.%d", (fixedInfo.FileVersionMS>>16)&0xff, (fixedInfo.FileVersionMS>>0)&0xff)
	if nextNibble := (fixedInfo.FileVersionLS >> 16) & 0xff; nextNibble != 0 {
		version += fmt.Sprintf(".%d", nextNibble)
	}
	if nextNibble := (fixedInfo.FileVersionLS >> 0) & 0xff; nextNibble != 0 {
		version += fmt.Sprintf(".%d", nextNibble)
	}
	return version
}

// ===========================================================
// ===========================================================
// CUSTOM ADAPTER CODE
// ===========================================================
// ===========================================================

func InitializeTunnelAdapter() (err error) {
	defer RecoverAndLogToFile()

	if A.Initialized {
		return
	}

	err = initialize_tunnel_adapter()
	if err != nil {
		CreateErrorLog("connect", " !! Unable to turn on VPN interface: ", err)
		return err
	}

	A.Initialized = true

	CreateLog("loader", "VPN interface ready")
	return
}

func AdapterCleanup(AH *AdapterHandle) {
	// syscall.SyscallN(procWintunCloseAdapter.Addr(), 1, AH.handle, 0, 0)
}

func initialize_tunnel_adapter() (err error) {

	A.Name = TUNNEL_ADAPTER_NAME
	A.TunHandle = windows.InvalidHandle
	A.Events = make(chan Event, 10)
	A.ForcedMTU = 1500
	// A.ForcedMTU = 1500

	A.NameU16, err = windows.UTF16PtrFromString(TUNNEL_ADAPTER_NAME)
	if err != nil {
		return
	}

	A.TypeU16, err = windows.UTF16PtrFromString("Routing")
	if err != nil {
		return
	}

	// TRY GENERATING A STATIC UID
	// A.GUID = new(windows.GUID)
	//https://github.com/microsoft/go-winio/blob/main/pkg/guid/guid.go

	//https://github.com/WireGuard/wintun/blob/master/README.md#wintuncreateadapter

	// A.wt = (*Adapter)(&A.Finalizer)

	var r1 uintptr
	var e1 error
	CreateLog("", "Initializing wintun adapter")
	r1, _, e1 = syscall.SyscallN(procWintunOpenAdapter.Addr(), uintptr(unsafe.Pointer(A.NameU16)))
	if r1 == 0 {
		r1, _, e1 = syscall.SyscallN(procWintunCreateAdapter.Addr(), uintptr(unsafe.Pointer(A.NameU16)), uintptr(unsafe.Pointer(A.TypeU16)), uintptr(unsafe.Pointer(A.GUID)))
		if r1 == 0 {
			err = e1
			return
		}

	}

	A.AdapterHandle = AdapterHandle{handle: r1}
	CreateLog("", "Starting buffer with 8MB capacity")
	runtime.SetFinalizer(&A.AdapterHandle, AdapterCleanup)

	//0x4000000
	err = A.Start(0x4000000)
	if err != nil {
		CreateLog("", "Error starting buffer for adapter reader", err)
	}

	return
}

var OUTPacket []byte
var OUTErr error

func VerifyAndBackupSettings(PotentialDefault *CONNECTION_SETTINGS) (err error) {

	GetDnsSettings(PotentialDefault)
	GetIPv6Settings(PotentialDefault)

	if PotentialDefault.DNS1 == "" || PotentialDefault.DNS1 == TUNNEL_ADAPTER_ADDRESS {
		PotentialDefault.DNS1 = "1.1.1.1"
		PotentialDefault.DNS2 = "8.8.8.8"
	}

	return
}

func FindDefaultInterfaceAndGatewayDuringStartup() (err error) {

	PotentialDefault, err := FindDefaultInterfaceAndGateway()
	if err != nil {
		CreateErrorLog("", "Could not find default interface and gateway >> ", err)
		return errors.New("")
	}

	if PotentialDefault == nil {
		CreateErrorLog("", "Could not find default interface and gateway")
		return errors.New("")
	}

	if PotentialDefault.DefaultRouter == "" {
		CreateErrorLog("", "Default interface had invalid Default router", PotentialDefault, " >> ", err)
		return errors.New("")
	}

	GLOBAL_STATE.DefaultInterface = PotentialDefault
	CreateLog("", "Default Interface >> ", GLOBAL_STATE.DefaultInterface)
	return
}

func GetDnsSettings(PotentialDefault *CONNECTION_SETTINGS) {

	cmd := exec.Command("netsh", "interface", "ipv4", "show", "dnsservers", `name=`+PotentialDefault.IFName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		CreateErrorLog("", "Unable to get DNS settings for interface: ", PotentialDefault.IFName, " // output: ", string(out))
		return
	}

	lineSplit := bytes.Split(out, []byte{13, 10})
	for ii, v := range lineSplit {
		if ii == 0 || ii == 1 {
			continue
		}

		if ii == 2 {
			contentSplit := bytes.Split(v, []byte{32})

			for iii, vv := range contentSplit {
				if bytes.Contains(vv, []byte{46}) {
					dnsString := string(contentSplit[iii])
					if dnsString == TUNNEL_ADAPTER_ADDRESS {
						continue
					}
					if dnsString == "" {
						continue
					}

					PotentialDefault.DNS1 = dnsString
				}
			}
		}

		if ii == 3 {
			dnsBytes := bytes.Replace(v, []byte{32}, []byte{}, -1)
			if bytes.Contains(dnsBytes, []byte{46}) {
				dnsString := string(dnsBytes)
				if dnsString == TUNNEL_ADAPTER_ADDRESS {
					continue
				}

				PotentialDefault.DNS2 = string(dnsBytes)
			}
		}
	}

}

func RestoreSettingsFromFile(PotentialDefault *CONNECTION_SETTINGS) {

	CreateLog("", "ADAPTER: ", PotentialDefault)
	CreateLog("", "RESTORING SETTINGS FROM FILE")

	backupFile, err := os.Open(GLOBAL_STATE.BasePath + PotentialDefault.IFName + "_backup")
	if err != nil {
		CreateErrorLog("", "Unable to open backup fule file, please restart the application. If this problem persists contact customer support")
		return
	}

	if backupFile != nil {
		defer backupFile.Close()
	}

	backupBytes, err := io.ReadAll(backupFile)
	if err != nil {
		CreateErrorLog("", "Unable to parse read file, please restart the application. If this problem persists contact customer support")
		return
	}

	RAS := new(CONNECTION_SETTINGS)
	err = json.Unmarshal(backupBytes, RAS)
	if err != nil {
		CreateErrorLog("", "Unable to parse read file, please restart the application. If this problem persists contact customer support")
		return
	}

	if RAS.DNS1 != "" && RAS.DNS1 != TUNNEL_ADAPTER_ADDRESS {
		GLOBAL_STATE.DefaultInterface = RAS
	} else {
		CreateErrorLog("", "Backup file contained broken DNS")
	}

	RestoreDNS()
	RestoreIPv6()

}

func GetIPv6Settings(PotentialDefault *CONNECTION_SETTINGS) {
	defer RecoverAndLogToFile()

	cmd := exec.Command("netsh", "interface", "ipv6", "show", "interface")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		CreateErrorLog("", "Unable to backup IPv6 setting // msg: ", err, " // output: ", string(out))
		return
	}

	lineSplit := bytes.Split(out, []byte{13, 10})

	found := false
	for _, vv := range lineSplit {
		contentSplit := bytes.Split(vv, []byte{32})
		for _, vvv := range contentSplit {
			if bytes.Equal(vvv, []byte(PotentialDefault.IFName)) {
				found = true
				if !PotentialDefault.IPV6Enabled {
					PotentialDefault.IPV6Enabled = true
					// CreateLog("", PotentialDefault.IFName, " || IPv6: ", true)
				}
			}
		}
	}

	if !found {
		if PotentialDefault.IPV6Enabled {
			PotentialDefault.IPV6Enabled = false
			CreateLog("", "IPv6 BACKUP: ", PotentialDefault.IFName, " || IPv6: ", true)
		}
	}

}

func RestoreIPv6() {
	defer RecoverAndLogToFile()

	if !C.DisableIPv6OnConnect {
		CreateLog("", "IPv6 settings unchanged")
		return
	}

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateErrorLog("", "Failed to restore IPv6 settings, interface settings not found")
		return
	}

	if GLOBAL_STATE.DefaultInterface.IPV6Enabled {

		cmd := exec.Command("powershell", "-NoProfile", "Enable-NetAdapterBinding -Name '"+GLOBAL_STATE.DefaultInterface.IFName+"' -ComponentID ms_tcpip6")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.CombinedOutput()
		if err != nil {
			CreateErrorLog("", "Error restoring IPv6 settings on interface: ", GLOBAL_STATE.DefaultInterface.IFName, "|| error: ", err, " || output: ", string(out))
		}

		CreateLog("", "IPv6 Restored on interface: ", GLOBAL_STATE.DefaultInterface.IFName)
	}

}

func ResetAfterFailedConnectionAttempt() {
	CreateLog("connect", "Connection attempt failed, reseting network configurations")
	_ = DisableAdapter()
	RestoreIPv6()
	RestoreDNS()
}

func RestoreDNS() error {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateErrorLog("", "Unable to restore DNS, no interface backup settings found")
		return errors.New("NO ADAPTER BACKUP SETTINGS FOUND")
	}

	_ = ClearDNS(GLOBAL_STATE.DefaultInterface.IFName)
	if GLOBAL_STATE.DefaultInterface.DNS1 == "" {
		return errors.New("NO PRIMARY DNS FOUND IN BACKUP SETTINGS")
	}

	if GLOBAL_STATE.DefaultInterface.DNS1 == "" || GLOBAL_STATE.DefaultInterface.DNS1 == TUNNEL_ADAPTER_ADDRESS {
		_ = SetDNS(GLOBAL_STATE.DefaultInterface.IFName, "1.1.1.1", "1")
	} else {
		_ = SetDNS(GLOBAL_STATE.DefaultInterface.IFName, GLOBAL_STATE.DefaultInterface.DNS1, "1")
	}

	if GLOBAL_STATE.DefaultInterface.DNS2 != "" {
		if GLOBAL_STATE.DefaultInterface.DNS2 == TUNNEL_ADAPTER_ADDRESS {
			_ = SetDNS(GLOBAL_STATE.DefaultInterface.IFName, "8.8.8.8", "2")
		} else {
			_ = SetDNS(GLOBAL_STATE.DefaultInterface.IFName, GLOBAL_STATE.DefaultInterface.DNS2, "2")
		}
	}

	CreateLog("", "DNS Restored on interface: ", GLOBAL_STATE.DefaultInterface.IFName)

	return nil
}

func ChangeDNS() error {
	defer RecoverAndLogToFile()
	start := time.Now()

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateErrorLog("", " !! Unable to change DNS on interfaces, no backup interface settings found")
		return errors.New("NO ADAPTER BACKUP SETTINGS FOUND")
	}

	CreateLog("connect", " Changing DNS on interface: ", GLOBAL_STATE.DefaultInterface.IFName)
	_ = ClearDNS(GLOBAL_STATE.DefaultInterface.IFName)
	errx := SetDNS(GLOBAL_STATE.DefaultInterface.IFName, TUNNEL_ADAPTER_ADDRESS, "1")
	if errx != nil {
		return errx
	}

	CreateLog("connect", "Updated DNS on default interface // time: ", fmt.Sprintf("%.0f", math.Abs(time.Since(start).Seconds())), " seconds")
	return nil
}

func DisableIPv6() error {
	defer RecoverAndLogToFile()
	start := time.Now()

	if !C.DisableIPv6OnConnect {
		CreateLog("connect", "IPv6 settings unchanged")
		return nil
	}

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateErrorLog("connect", " !! Unable to turn off IPv6, no interface settings found")
		return errors.New("NO ADAPTER BACKUP SETTINGS FOUND")
	}

	i := GLOBAL_STATE.DefaultInterface.IFName
	CreateLog("connect", " Disabling IPv6 on interface: ", i)

	cmd := exec.Command("powershell", "-NoProfile", "Disable-NetAdapterBinding -Name '"+i+"' -ComponentID ms_tcpip6")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		CreateErrorLog("", "Error disabling IPv6 on interface: ", i, " || msg: ", err, " || output: ", string(out))
		return err
	}

	CreateLog("connect", "Finished disabling IPv6 // time: ", fmt.Sprintf("%.0f", math.Abs(time.Since(start).Seconds())), " seconds")

	return nil

}

func AddRoute(IP string) (err error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateLog("", "Not adding route, no default interface")
		return errors.New("no default interface")
	}

	_ = DeleteRoute(IP, false)
	cmd := exec.Command("route", "ADD", IP, "MASK", "255.255.255.255", GLOBAL_STATE.DefaultInterface.DefaultRouter, "METRIC", "1")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		CreateErrorLog("", "ROUTE || Unable to add route to: ", IP, " || using Gateway: ", GLOBAL_STATE.DefaultInterface.DefaultRouter, " || msg: ", err, " || output: ", string(out))
		return
	}

	return
}

func DeleteRoute(IP string, ignoreActiveRouter bool) (err error) {
	defer RecoverAndLogToFile()

	if GLOBAL_STATE.DefaultInterface == nil {
		CreateLog("", "Not deleting route, no default interface")
		return errors.New("no default interface")
	}

	if !ignoreActiveRouter {
		if GLOBAL_STATE.ActiveRouter != nil && GLOBAL_STATE.ActiveRouter.IP == IP {
			return
		}
	}

	// CreateLog("", "Deleting route: ", IP, " || Gateway: ", GLOBAL_STATE.DefaultInterface.DefaultRouter)

	cmd := exec.Command("route", "DELETE", IP)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		CreateErrorLog("", "ROUTE || Unable to delete route: ", IP, " || Gateway: ", GLOBAL_STATE.DefaultInterface.DefaultRouter, " || msg: ", err, " || output: ", string(out))
		return err
	}

	return
}

func AssignIPAddressToAdapter() (err error) {
	defer RecoverAndLogToFile()

	cmd := exec.Command("netsh", "interface", "ipv4", "set", "address", `name="`+TUNNEL_ADAPTER_NAME+`"`, "static", TUNNEL_ADAPTER_ADDRESS, "255.255.255.0", "0.0.0.0")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		CreateErrorLog("", "Error enabling tunnel adapter // msg: ", err, " // output: ", string(out))
		return errors.New("We could not enable the tunnel adapter:" + err.Error())
	}

	CreateLog("connect", " Assigned IP to VPN interface")
	return nil
}

func SetInterfaceStateToDown() (err error) {
	return DisableAdapter()
}

func DisableAdapter() (err error) {
	defer RecoverAndLogToFile()

	cmd := exec.Command("netsh", "interface", "ipv4", "delete", "address", `name="`+TUNNEL_ADAPTER_NAME+`"`, "addr=", TUNNEL_ADAPTER_ADDRESS, "gateway=", "All")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()

	if err != nil {
		CreateErrorLog("", "Error disabling adapter || msg: ", err, " || output: ", string(out))
		return err
	}

	CreateLog("file", "Interface disabled")

	return
}

func EnablePacketRouting() (err error) {
	defer RecoverAndLogToFile()

	err = AssignIPAddressToAdapter()
	if err != nil {
		return
	}

	err = CloseAllOpenSockets()
	if err != nil {
		return
	}

	err = DisableIPv6()
	if err != nil {
		return
	}

	err = ChangeDNSOnTunnelInterface()
	if err != nil {
		return
	}

	err = ChangeDNS()
	if err != nil {
		return
	}
	return
}

func ChangeDNSOnTunnelInterface() error {
	defer RecoverAndLogToFile()

	start := time.Now()

	_ = ClearDNS(TUNNEL_ADAPTER_NAME)
	errx := SetDNS(TUNNEL_ADAPTER_NAME, C.DNS1, "1")
	if errx != nil {
		return errx
	}

	if C.DNS2 != "" && C.DNS2 != C.DNS1 {
		errx = SetDNS(TUNNEL_ADAPTER_NAME, C.DNS2, "2")
		if errx != nil {
			return errx
		}
	}

	CreateLog("connect", "Updated DNS on VPN interface // time: ", fmt.Sprintf("%.0f", math.Abs(time.Since(start).Seconds())), " Seconds")
	return nil
}

func ChangeDNSWhileConnected() error {
	defer RecoverAndLogToFile()
	CreateLog("loader", "Updating DNS on tunnel interface")
	err := ChangeDNSOnTunnelInterface()
	if err != nil {
		CreateLog("loader", "Unable to change DNS on tunnel interface || msg: ", err)
		return err
	}
	return nil
}

func ClearDNS(Interface string) error {
	defer RecoverAndLogToFile()

	cmd := exec.Command("netsh", "interface", "ipv4", "delete", "dns", "name="+Interface, "addr=all")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	out, err := cmd.Output()
	if err != nil {
		CreateErrorLog("", "NETSH || Error clearing DNS on interface: ", Interface, " || msg: ", err, " || output: ", string(out))
		return err
	}
	CreateLog("file", "DNS cleared on interface: ", Interface)
	return nil

}

func SetDNS(Interface, IP string, index string) error {
	defer RecoverAndLogToFile()

	cmd := exec.Command("netsh", "interface", "ipv4", "add", "dnsservers", `name=`+Interface, "address="+IP, "index="+index)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()

	if err != nil {
		CreateErrorLog("", "NETSH || Error setting DNS: ", IP, " || interface: ", Interface, " || msg: ", err, " || output: ", string(out))
		return err
	}

	CreateLog("file", "DNS set on interface: ", Interface, " ", IP, " ", index)
	return nil
}

func LaunchPreperation() (err error) {
	return InitializeTunnelAdapter()
}

func FindDefaultInterfaceAndGateway() (NEW_DEFAULT *CONNECTION_SETTINGS, err error) {
	defer RecoverAndLogToFile()

	INTERFACE_SETTINGS := FindAllInterfaces()

	cmd := exec.Command("powershell", "-NoProfile", "Get-NetRoute -DestinationPrefix '0.0.0.0/0' | ConvertTo-Json -Depth 1")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		CreateErrorLog("", "Error fetching default routes || msg: ", err, " || output: ", string(out))
		return nil, err
	}

	// log.Println(string(out))
	isList := false
	PS_IF := new(PS_DEFAULT_ROUTES)
	if len(out) > 0 {
		if string(out[0]) == "[" {
			isList = true
		}
	}

	PS_IFLIST = nil
	PS_IFLIST = make([]*PS_DEFAULT_ROUTES, 0)

	if isList {
		err = json.Unmarshal(out, &PS_IFLIST)
		if err != nil {
			CreateErrorLog("", "Error parsing powershell output for default route LIST", err)
			return nil, err
		}
	} else {
		err = json.Unmarshal(out, PS_IF)
		if err != nil {
			CreateErrorLog("", "Error parsing powershell output for default route", err)
			return nil, err
		}
		PS_IFLIST = append(PS_IFLIST, PS_IF)
	}

	defaultIndex := 0
	lowestMetric := 999999
	for i := range PS_IFLIST {
		if PS_IFLIST[i] == nil {
			continue
		}

		IF := PS_IFLIST[i]

		for i := range INTERFACE_SETTINGS {
			if INTERFACE_SETTINGS[i] == nil {
				continue
			}

			if i == TUNNEL_ADAPTER_NAME {
				continue
			}

			if INTERFACE_SETTINGS[i].Index == IF.InterfaceIndex {
				INTERFACE_SETTINGS[i].Hop = IF.NextHop
				INTERFACE_SETTINGS[i].Metric = IF.RouteMetric

				if IF.InterfaceMetric < lowestMetric {
					lowestMetric = IF.RouteMetric
					defaultIndex = IF.InterfaceIndex
				}
			}
		}

	}

	for i := range INTERFACE_SETTINGS {
		if INTERFACE_SETTINGS[i] == nil {
			continue
		}

		if INTERFACE_SETTINGS[i].Index == defaultIndex {

			NEW_DEFAULT = new(CONNECTION_SETTINGS)
			NEW_DEFAULT.IFName = i
			NEW_DEFAULT.DefaultRouter = INTERFACE_SETTINGS[i].Hop
			return
		}
	}

	if lowestMetric == 999999 || defaultIndex == 0 {
		return nil, errors.New("")
	}

	return
}

func PrintRouters() ([]byte, error) {
	cmd := exec.Command("route", "print")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func PrintInterfaces() ([]byte, error) {
	cmd := exec.Command("ipconfig")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func PrintDNS() ([]byte, error) {
	cmd := exec.Command("netsh", "interface", "ipv4", "show", "dnsservers")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// =============================================
// =============================================
// =============================================
// =============================================
// =============================================
// WINDOWS DLL STUFF

func CloseAllOpenSockets() error {

	if !C.CloseConnectionsOnConnect {
		CreateLog("", "Leaving open sockets intact")
		return nil
	}

	CreateLog("", "Closing all open socket")

	tcpTable, err := GetOpenSockets()
	if err != nil {
		CreateErrorLog("", "Unable to get TCP Socket table", err)
		return err
	}

	for _, row := range tcpTable {
		localAddr := fmt.Sprintf("%d.%d.%d.%d", byte(row.dwLocalAddr), byte(row.dwLocalAddr>>8), byte(row.dwLocalAddr>>16), byte(row.dwLocalAddr>>24))
		remoteAddr := fmt.Sprintf("%d.%d.%d.%d", byte(row.dwRemoteAddr), byte(row.dwRemoteAddr>>8), byte(row.dwRemoteAddr>>16), byte(row.dwRemoteAddr>>24))

		if localAddr != "0.0.0.0" && localAddr != "127.0.0.1" {
			if remoteAddr == GLOBAL_STATE.ActiveRouter.IP {
				continue
			}
			_ = DeleteSocket(row)
		}

	}

	return nil
}

func GetOpenSockets() ([]MIB_TCPROW_OWNER_PID, error) {
	var tcpTable MIB_TCPTABLE_OWNER_PID
	size := uintptr(unsafe.Sizeof(tcpTable))

	r, _, err := GetTCP.Call(
		uintptr(unsafe.Pointer(&tcpTable)),
		uintptr(unsafe.Pointer(&size)),
		1,
		syscall.AF_INET,
		MIB_TCP_TABLE_OWNER_PID_ALL,
		0)

	if r == 0 {
		entries := tcpTable.dwNumEntries
		return tcpTable.table[:entries], nil
	}

	return nil, err
}

func DeleteSocket(row MIB_TCPROW_OWNER_PID) error {
	row.dwState = MIB_TCP_STATE_DELETE_TCB
	r, _, err := SetTCP.Call(
		uintptr(unsafe.Pointer(&row)))

	if r == 0 {
		return nil
	}

	return err
}
