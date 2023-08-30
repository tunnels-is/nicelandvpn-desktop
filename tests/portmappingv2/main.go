package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

// 23422 > 1.1.1.1:53
// 2000 > 1.1.1.1:53

// Original Local port > Mapped Local Port + Remote Port
// var LP_TCP [math.MaxUint16]*O0
// var LP_UDP [math.MaxUint16]*O0

// Mapped Local Port > Original Local Port + Remote Port
// var MLP_TO_RP [math.MaxUint16]*O0

// type RP struct {
// 	P [math.MaxUint16]*O0
// }

// type O0 struct {
// 	o0 [256]*O1
// }

var TCP_o0 [256]*O1
var UDP_o0 [256]*O1

type O1 struct {
	o1 [256]*O2
}

type O2 struct {
	o2 [256]*O3
}

type O3 struct {
	o3 [256]*M
}

type M struct {
	LOCAL  map[uint16]*RemotePort
	REMOTE map[uint16]*RemotePort
	// X []*RemotePort
	Lock sync.Mutex
}

type RemotePort struct {
	Local        uint16
	Mapped       uint16
	Remote       uint16
	LastActivity time.Time
}

var PM_START_PORT uint16 = 1000
var PM_END_PORT uint16 = 2000

func CreatePortMapping(protoMap *[256]*O1, ip [4]byte, lport, rport uint16) *RemotePort {

	if protoMap[ip[0]] == nil {
		protoMap[ip[0]] = new(O1)
	}

	if protoMap[ip[0]].o1[ip[1]] == nil {
		protoMap[ip[0]].o1[ip[1]] = new(O2)
	}

	if protoMap[ip[0]].o1[ip[1]].o2[ip[2]] == nil {
		protoMap[ip[0]].o1[ip[1]].o2[ip[2]] = new(O3)
	}

	var m *M
	if protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]] == nil {
		protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]] = new(M)
		m = protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]
		m.Lock = sync.Mutex{}
		m.LOCAL = make(map[uint16]*RemotePort)
		m.REMOTE = make(map[uint16]*RemotePort)
	} else {
		m = protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]
	}

	m.Lock.Lock()
	mapping, ok := m.LOCAL[lport]
	m.Lock.Unlock()
	if ok {
		log.Println("FOUND MAPPING", mapping)
		return mapping
	}

	for i := PM_START_PORT; i <= PM_END_PORT; i++ {

		m.Lock.Lock()
		XR, ok := m.REMOTE[i]
		m.Lock.Unlock()

		if !ok || XR == nil {

			m.Lock.Lock()
			m.REMOTE[i] = new(RemotePort)
			m.REMOTE[i].LastActivity = time.Now()
			m.REMOTE[i].Local = lport
			m.REMOTE[i].Mapped = i
			m.REMOTE[i].Remote = rport
			m.LOCAL[lport] = m.REMOTE[i]
			m.Lock.Unlock()
			// log.Println("CU:", ip, "L:", LP, "O:", DP, "M:", i)
			return m.REMOTE[i]
		}
	}

	log.Println("", "Create TCP (NO PORTS): ", ip, " L: ", lport, " R: ", rport)
	return nil
}

func GetEgressPortMapping(protoMap *[256]*O1, ip [4]byte, port uint16) (mapping *RemotePort) {

	if protoMap[ip[0]] == nil {
		return nil
	}
	if protoMap[ip[0]].o1[ip[1]] == nil {
		return nil
	}
	if protoMap[ip[0]].o1[ip[1]].o2[ip[2]] == nil {
		return nil
	}
	if protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]] == nil {
		return nil
	}

	m := protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]

	m.Lock.Lock()
	mapping, _ = m.LOCAL[port]
	m.Lock.Unlock()

	return

}

// var h, l uint8 = uint8(i>>8), uint8(i&0xff)
func main() {

	ip2, ip2net, err := net.ParseCIDR("192.168.1.0/22")
	if err != nil {
		log.Println(err)
		return
	}

	ip2 = ip2.Mask(ip2net.Mask)
	var lport uint16 = 1000
	var dport uint16 = 5000

	for ip2net.Contains(ip2) {
		_ = CreatePortMapping(&TCP_o0, [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, lport, dport)
		// lport++
		// _ = CreatePortMapping(&TCP_o0, [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, lport, dport)
		lport++
		dport++
		// m1.Value = ip2.String()
		inc(ip2)
	}

	WALK_PORT_MAP(&TCP_o0)
	// ip2, ip2net, err = net.ParseCIDR("192.168.1.0/22")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// ip2 = ip2.Mask(ip2net.Mask)
	// lport = 1000

	// for ip2net.Contains(ip2) {
	// 	m := GetEgressPortMapping(&TCP_o0, [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, lport)
	// 	lport++
	// 	log.Println([4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, " - L:", m.Local, " - M:", m.Mapped, " - R:", m.Remote)
	// 	m = GetEgressPortMapping(&TCP_o0, [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, lport)
	// 	lport++
	// 	log.Println([4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, " - L:", m.Local, " - M:", m.Mapped, " - R:", m.Remote)
	// 	inc(ip2)
	// }

	for {
		runtime.GC()
		time.Sleep(1 * time.Second)
		PrintMemUsage()
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func getRealSizeOf(v interface{}) (int, error) {
	b := new(bytes.Buffer)
	if err := gob.NewEncoder(b).Encode(v); err != nil {
		return 0, err
	}
	return b.Len(), nil
}

func WALK_PORT_MAP(protoMap *[256]*O1) {
	start := time.Now()
	// LOG(WO)
	// var RP *RemotePort
	var count int = 0

	for i := 0; i < 256; i++ {
		// time.Sleep(1 * time.Microsecond)
		if protoMap[i] == nil {
			continue
		}
		for i1 := 0; i1 < 256; i1++ {
			if protoMap[i].o1[i1] == nil {
				continue
			}
			for i2 := 0; i2 < 256; i2++ {
				if protoMap[i].o1[i1].o2[i2] == nil {
					continue
				}
				for i3 := 0; i3 < 256; i3++ {
					if protoMap[i].o1[i1].o2[i2].o3[i3] == nil {
						continue
					} else {
						count++
					}
					// log.Println(RP.Mapped)
				}
			}
		}
	}

	done := time.Since(start).Nanoseconds()
	log.Println(" @@@@@ WALK >>> Micro @@@@> ", done, " >> count >> ", count)
}
