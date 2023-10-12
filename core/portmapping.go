package core

import (
	"encoding/binary"
	"net"
	"sync"
	"time"
)

var TCP_o0 [256]*O1
var UDP_o0 [256]*O1

func InstantlyClearPortMaps() {
	for i := range TCP_o0 {
		TCP_o0[i] = nil
		TCP_o0[i] = new(O1)
	}

	for i := range UDP_o0 {
		UDP_o0[i] = nil
		UDP_o0[i] = new(O1)
	}
}

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
	LOCAL  map[[2]byte]*RP
	REMOTE map[[2]byte]*RP
	// X []*RemotePort
	Lock sync.Mutex
}

type RP struct {
	Local        [2]byte
	Mapped       [2]byte
	Remote       [2]byte
	LastActivity time.Time
}

func CreateOrGetPortMapping(protoMap *[256]*O1, ip [4]byte, lport, rport [2]byte) *RP {

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
		m.LOCAL = make(map[[2]byte]*RP)
		m.REMOTE = make(map[[2]byte]*RP)
	} else {
		m = protoMap[ip[0]].o1[ip[1]].o2[ip[2]].o3[ip[3]]
	}

	m.Lock.Lock()
	mapping, ok := m.LOCAL[lport]
	m.Lock.Unlock()
	if ok {
		mapping.LastActivity = time.Now()
		return mapping
	}

	var newPort = [2]byte{}
	for i := AS.StartPort; i <= AS.EndPort; i++ {

		binary.BigEndian.PutUint16(newPort[:], i)
		m.Lock.Lock()
		XR, ok := m.REMOTE[newPort]
		m.Lock.Unlock()

		if !ok || XR == nil {

			m.Lock.Lock()
			m.REMOTE[newPort] = new(RP)
			m.REMOTE[newPort].LastActivity = time.Now()
			m.REMOTE[newPort].Local = lport
			m.REMOTE[newPort].Mapped = newPort
			m.REMOTE[newPort].Remote = rport
			m.LOCAL[lport] = m.REMOTE[newPort]
			m.Lock.Unlock()
			// log.Println("CU:", ip, "L:", lport, "R:", rport, "M:", i)
			return m.REMOTE[newPort]
		}
	}

	CreateLog("", "Unable to create port mapping")
	// log.Println("", "Create (NO PORTS): ", ip, " L: ", lport, " R: ", rport)
	return nil
}

func GetIngressPortMapping(protoMap *[256]*O1, ip [4]byte, port [2]byte) (mapping *RP) {

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
	mapping, _ = m.REMOTE[port]
	if mapping != nil {
		mapping.LastActivity = time.Now()
	}
	m.Lock.Unlock()

	return

}

func CleanPorts(MONITOR chan int) {
	defer func() {
		time.Sleep(10 * time.Second)
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 9
		}
	}()

	defer RecoverAndLogToFile()
	CleanPortMap(&TCP_o0, "tcp")
	CleanPortMap(&UDP_o0, "udp")
}

func CleanPortMap(protoMap *[256]*O1, mapType string) {
	// start := time.Now()

	var count int = 0

	for i := 0; i < 256; i++ {
		// time.Sleep(1 * time.Microsecond)
		if protoMap[i] == nil {
			continue
		}

		for i1 := 0; i1 < 256; i1++ {
			o1Active := 0
			if protoMap[i].o1[i1] == nil {
				continue
			}

			for i2 := 0; i2 < 256; i2++ {
				o2Active := 0
				if protoMap[i].o1[i1].o2[i2] == nil {
					continue
				}

				for i3 := 0; i3 < 256; i3++ {
					o3Active := 0
					if protoMap[i].o1[i1].o2[i2].o3[i3] == nil {
						continue
					}
					// LOCK /???
					m := protoMap[i].o1[i1].o2[i2].o3[i3]

					if mapType == "udp" {

						var timeout float64 = 29
						ip := net.IP{byte(i), byte(i), byte(i), byte(i)}.String()
						if ip == C.DNS1 || ip == C.DNS2 {
							timeout = 4
						}

						m.Lock.Lock()
						for ri := range m.REMOTE {
							if time.Since(m.REMOTE[ri].LastActivity).Seconds() > timeout {
								// log.Println("DELETING PM: ", m.REMOTE)
								delete(m.LOCAL, m.REMOTE[ri].Local)
								delete(m.REMOTE, ri)
							} else {
								o1Active++
								o2Active++
								o3Active++
							}
						}
						m.Lock.Unlock()

					} else if mapType == "tcp" {

						m.Lock.Lock()
						for ri := range m.REMOTE {
							if time.Since(m.REMOTE[ri].LastActivity).Seconds() > 86 {
								// log.Println("DELETING PM: ", m.REMOTE)
								delete(m.LOCAL, m.REMOTE[ri].Local)
								delete(m.REMOTE, ri)
							} else {
								o1Active++
								o2Active++
								o3Active++
							}
						}
						m.Lock.Unlock()

					}

					if o3Active == 0 {
						// log.Println("DEL OCT 3: ", i1, i2, " > ", i3)
						protoMap[i].o1[i1].o2[i2].o3[i3] = nil
					}

					count++

				} // o3

				if o2Active == 0 {
					// log.Println("DEL OCT 2: ", i1, ">", i2)
					protoMap[i].o1[i1].o2[i2] = nil
				}

			} // o2

			if o1Active == 0 {
				// log.Println("DEL OCT 1: ", i1)
				protoMap[i].o1[i1] = nil
			}

		} // o1

	}

	// done := time.Since(start).Nanoseconds()
	// log.Println(" @@@@@ WALK >>> Micro @@@@> ", done, " >> count >> ", count)
}
