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
	// LOCAL  map[uint16]*RemotePort
	// REMOTE map[uint16]*RemotePort
	X    []*RemotePort
	Lock sync.Mutex
}

type RemotePort struct {
	Local        uint16
	Original     uint16
	Mapped       uint16
	LastActivity time.Time
}

func GetIngressPortMapping(protoMap *[256]*O1, ip [4]byte, port uint16) (mapping *RemotePort) {

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

	// m := protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]
	return protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]].X[0]

	// m.Lock.Lock()
	// mapping, _ = m.REMOTE[port]
	// m.Lock.Unlock()

	return

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

	// m := protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]
	return protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]].X[0]

	// m.Lock.Lock()
	// mapping, _ = m.LOCAL[port]
	// m.Lock.Unlock()

	return

}

func GET_OR_CREATE_MAPPING(protoMap *[256]*O1, ip [4]byte, port uint16) (m *M) {

	if protoMap[ip[0]] == nil {
		protoMap[ip[0]] = new(O1)
	}

	if protoMap[ip[0]].o1[ip[1]] == nil {
		protoMap[ip[0]].o1[ip[1]] = new(O2)
	}

	if protoMap[ip[0]].o1[ip[1]].o2[ip[2]] == nil {
		protoMap[ip[0]].o1[ip[1]].o2[ip[2]] = new(O3)
	}

	if protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]] == nil {
		protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]] = new(M)
		m = protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]
		m.X = make([]*RemotePort, 2000-1000)
		// m.LOCAL = make(map[uint16]*RemotePort)
		// m.REMOTE = make(map[uint16]*RemotePort)
		m.Lock = sync.Mutex{}
	}

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

	for ip2net.Contains(ip2) {
		_ = GET_OR_CREATE_MAPPING(&TCP_o0, [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, 100)
		// m1.Value = ip2.String()
		inc(ip2)
	}

	ip2, ip2net, err = net.ParseCIDR("192.168.1.0/22")
	if err != nil {
		log.Println(err)
		return
	}

	ip2 = ip2.Mask(ip2net.Mask)

	for ip2net.Contains(ip2) {
		_ = GetEgressPortMapping(&TCP_o0, [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}, 100)
		// log.Println(m1.LOCAL)
		inc(ip2)
	}

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
