package core

import (
	"encoding/binary"
	"log"
	"sync"
	"time"
)

// func LOG_SIGNAL(TCPL layers.TCP, IPL layers.IPv4, PORT RemotePort) {
// 	log.Println(IPL.SrcIP, ":", TCPL.SrcPort, " >> ", IPL.DstIP, ":", TCPL.DstPort, " || A:", TCPL.ACK, "P:", TCPL.PSH, "F:", TCPL.FIN, "R:", TCPL.RST, " || L:", PORT.Local, "M:", PORT.Mapped, "|| IF:", PORT.IFIN, "EF:", PORT.EFIN, "FA:", PORT.FACK)
// }

func GetUDPMapping(ip [4]byte, LP uint16) *RemotePort {
	defer UDP_MAP_LOCK.RUnlock()
	defer RecoverAndLogToFile()
	UDP_MAP_LOCK.RLock()

	IPMap, ok := UDP_MAP[ip]
	if !ok {
		// log.Println("GU:(NO IP,ok):", ip)
		return nil
	}

	if IPMap == nil {
		// log.Println("GU:(NO IP,nil):", ip)
		return nil
	}

	REMOTE_TEMP, ok := IPMap.REMOTE[LP]
	if !ok {
		// log.Println("GU(NO REMOTE,ok):", ip, "L:", LP, "D:", DP)
		return nil
	}

	if REMOTE_TEMP == nil {
		// log.Println("GU(NO REMOTE,nil):", ip, "L:", LP, "D:", DP)
		return nil
	}

	REMOTE_TEMP.LastActivity = time.Now()
	// log.Println("GU:", ip, "L:", LP, "D:", DP, " || L:", REMOTE_TEMP.Local, "O:", REMOTE_TEMP.Original, "M:", REMOTE_TEMP.Mapped)
	return IPMap.REMOTE[LP]
}

func GetTCPMapping(ip [4]byte, LP uint16) *RemotePort {
	defer TCP_MAP_LOCK.RUnlock()
	defer RecoverAndLogToFile()
	TCP_MAP_LOCK.RLock()

	IPMap, ok := TCP_MAP[ip]
	if !ok {
		// log.Println("GT(NO IP,ok):", ip)
		return nil
	}
	if IPMap == nil {
		// log.Println("GT(NO IP,nil):", ip)
		return nil
	}

	REMOTE_TEMP, ok := IPMap.REMOTE[LP]
	if !ok {
		// log.Println("GT(NO REMOTE,ok):", ip, LP, DP)
		return nil
	}

	if REMOTE_TEMP == nil {
		// log.Println("GT(NO REMOTE,nil):", ip, LP, DP)
		return nil
	}

	REMOTE_TEMP.LastActivity = time.Now()

	// log.Println("GT:", ip, "L:", LP, "D:", DP, "|| L:", REMOTE_TEMP.Local, "O:", REMOTE_TEMP.Original, "R:", REMOTE_TEMP.Mapped)

	return IPMap.REMOTE[LP]
}

func GetOutgoingTCPMapping(ip [4]byte, LP uint16, DP uint16) *RemotePort {
	defer TCP_MAP_LOCK.RUnlock()
	defer RecoverAndLogToFile()
	TCP_MAP_LOCK.RLock()

	XM, ok := TCP_MAP[ip]
	if !ok || XM == nil {
		return nil
	}

	LX, ok := TCP_MAP[ip].LOCAL[LP]
	if ok && LX != nil {
		if TCP_MAP[ip].LOCAL[LP].Original == DP {
			// log.Println("CT:(EXISTED)", ip, "L:", LP, "O:", DP, "M:", "L:", TCP_MAP[ip].LOCAL[LP].Local, "O:", TCP_MAP[ip].LOCAL[LP].Original, "M:", TCP_MAP[ip].LOCAL[LP].Mapped)
			TCP_MAP[ip].LOCAL[LP].LastActivity = time.Now()
			return TCP_MAP[ip].LOCAL[LP]
		}
	}

	return nil
}

// SYN and NO ACK
func CreateTCPMapping(ip [4]byte, LP uint16, DP uint16) *RemotePort {
	defer TCP_MAP_LOCK.Unlock()
	defer RecoverAndLogToFile()
	TCP_MAP_LOCK.Lock()

	var i uint16
	// var XR *RemotePort

	XM, ok := TCP_MAP[ip]
	if !ok || XM == nil {
		TCP_MAP[ip] = new(IP)
		TCP_MAP[ip].LOCAL = make(map[uint16]*RemotePort)
		TCP_MAP[ip].REMOTE = make(map[uint16]*RemotePort)
	}

	// LX, ok := TCP_MAP[ip].LOCAL[LP]
	// if ok && LX != nil {
	// 	if TCP_MAP[ip].LOCAL[LP].Original == DP {
	// 		// log.Println("CT:(EXISTED)", ip, "L:", LP, "O:", DP, "M:", "L:", TCP_MAP[ip].LOCAL[LP].Local, "O:", TCP_MAP[ip].LOCAL[LP].Original, "M:", TCP_MAP[ip].LOCAL[LP].Mapped)
	// 		// TCP_MAP[ip].LOCAL[LP].FACK = false
	// 		// TCP_MAP[ip].LOCAL[LP].IFIN = false
	// 		// TCP_MAP[ip].LOCAL[LP].EFIN = false
	// 		// TCP_MAP[ip].LOCAL[LP].FINStarted = time.Time{}
	// 		TCP_MAP[ip].LOCAL[LP].LastActivity = time.Now()
	// 		return TCP_MAP[ip].LOCAL[LP]
	// 	}
	// }

	for i = AS.StartPort; i <= AS.EndPort; i++ {
		XR, ok := TCP_MAP[ip].REMOTE[i]
		if !ok || XR == nil {
			TCP_MAP[ip].REMOTE[i] = new(RemotePort)
			TCP_MAP[ip].REMOTE[i].LastActivity = time.Now()
			TCP_MAP[ip].REMOTE[i].Local = LP
			TCP_MAP[ip].REMOTE[i].Mapped = i
			TCP_MAP[ip].REMOTE[i].Original = DP
			TCP_MAP[ip].LOCAL[LP] = TCP_MAP[ip].REMOTE[i]

			// log.Println("CT:", ip, "L:", LP, "O:", DP, "M:", i)
			return TCP_MAP[ip].REMOTE[i]
		}
	}

	CreateErrorLog("", "Create TCP (NO PORTS): ", ip, " L: ", LP, " O: ", DP)
	return nil

}

func GetOutgoingUDPMapping(ip [4]byte, LP uint16, DP uint16) *RemotePort {
	defer UDP_MAP_LOCK.RUnlock()
	defer RecoverAndLogToFile()
	UDP_MAP_LOCK.RLock()

	XP, ok := UDP_MAP[ip]
	if !ok || XP == nil {
		return nil
	}

	if UDP_MAP[ip].LOCAL[LP] != nil {
		if UDP_MAP[ip].LOCAL[LP].Original == DP {
			// log.Println("CU(existed):", ip, "L:", LP, "D:", DP, "L:", UDP_MAP[ip].LOCAL[LP].Local, "O:", UDP_MAP[ip].LOCAL[LP].Original, "M:", UDP_MAP[ip].LOCAL[LP].Mapped)
			UDP_MAP[ip].LOCAL[LP].LastActivity = time.Now()
			return UDP_MAP[ip].LOCAL[LP]
		}
	}

	return nil
}

// func GetOrCreateUDPMapping(ip [4]byte, LP uint16, DP uint16) *RemotePort {
func GetOrCreateUDPMapping(ip [4]byte, LP uint16, DP uint16) *RemotePort {
	defer UDP_MAP_LOCK.Unlock()
	defer RecoverAndLogToFile()
	UDP_MAP_LOCK.Lock()

	var i uint16

	XP, ok := UDP_MAP[ip]
	if !ok || XP == nil {
		UDP_MAP[ip] = new(IP)
		UDP_MAP[ip].LOCAL = make(map[uint16]*RemotePort)
		UDP_MAP[ip].REMOTE = make(map[uint16]*RemotePort)
	}

	for i = AS.StartPort; i <= AS.EndPort; i++ {
		XR, ok := UDP_MAP[ip].REMOTE[i]
		if !ok || XR == nil {
			UDP_MAP[ip].REMOTE[i] = new(RemotePort)
			UDP_MAP[ip].REMOTE[i].LastActivity = time.Now()
			UDP_MAP[ip].REMOTE[i].Local = LP
			UDP_MAP[ip].REMOTE[i].Mapped = i
			UDP_MAP[ip].REMOTE[i].Original = DP
			UDP_MAP[ip].LOCAL[LP] = UDP_MAP[ip].REMOTE[i]
			// log.Println("CU:", ip, "L:", LP, "O:", DP, "M:", i)
			return UDP_MAP[ip].REMOTE[i]
		}
	}

	CreateErrorLog("", "Create UDP g (NO PORTS): ", ip, " L: ", LP, " O: ", DP)
	return nil
}

func CleanTCPPorts() {
	// var gotLock bool = false
	defer func() {
		// if gotLock {
		TCP_MAP_LOCK.Unlock()
		// }
	}()
	defer RecoverAndLogToFile()

	TCP_MAP_LOCK.Lock()
	// gotLock = true

	for i := range TCP_MAP {
		for ii := range TCP_MAP[i].REMOTE {
			if TCP_MAP[i].REMOTE[ii] == nil {
				continue
			}

			// R := TCP_MAP[i].REMOTE[ii]

			// TCP_MAP_LOCK.Unlock()
			// gotLock = false

			// found := false
			// for si := range CurrentOpenSockets {
			// 	if i == [4]byte{1, 1, 1, 193} {
			// 		log.Println("COMPARISON: ", CurrentOpenSockets[si].RemoteIP, " >> ", i, " || ", CurrentOpenSockets[si].LocalPort, "-", R.Local, CurrentOpenSockets[si].RemotePort, "-", R.Original, "/", R.Mapped)
			// 	}
			// 	if CurrentOpenSockets[si].RemoteIP == i {
			// 		if CurrentOpenSockets[si].LocalPort == R.Local && CurrentOpenSockets[si].RemotePort == R.Original {
			// 			log.Println("FOUND PORT!", CurrentOpenSockets[si].RemoteAddress, CurrentOpenSockets[si].LocalPort, CurrentOpenSockets[si].RemotePort)
			// 			found = true
			// 		}
			// 	}
			// }

			// TCP_MAP_LOCK.Lock()
			// gotLock = true

			// if TCP_MAP[i].REMOTE[ii].FACK {

			// 	log.Println("DEL TCP(FA): ", i, "L:", TCP_MAP[i].REMOTE[ii].Local, "O:", TCP_MAP[i].REMOTE[ii].Original, "M:", TCP_MAP[i].REMOTE[ii].Mapped)
			// 	delete(TCP_MAP[i].LOCAL, TCP_MAP[i].REMOTE[ii].Local)
			// 	delete(TCP_MAP[i].REMOTE, ii)
			// 	continue

			// } else if !TCP_MAP[i].REMOTE[ii].FINStarted.IsZero() {

			// 	if time.Since(TCP_MAP[i].REMOTE[ii].FINStarted).Seconds() > 30 {
			// 		log.Println("DEL TCP(FS): ", i, "L:", TCP_MAP[i].REMOTE[ii].Local, "O:", TCP_MAP[i].REMOTE[ii].Original, "M:", TCP_MAP[i].REMOTE[ii].Mapped)
			// 		delete(TCP_MAP[i].LOCAL, TCP_MAP[i].REMOTE[ii].Local)
			// 		delete(TCP_MAP[i].REMOTE, ii)
			// 		continue
			// 	}

			// }

			// if !found {
			// 	log.Println("DELETING MAPPING: ", i, TCP_MAP[i].REMOTE[ii])
			// 	delete(TCP_MAP[i].LOCAL, TCP_MAP[i].REMOTE[ii].Local)
			// 	delete(TCP_MAP[i].REMOTE, ii)
			// 	continue
			// }

			if time.Since(TCP_MAP[i].REMOTE[ii].LastActivity).Seconds() > 85 {
				log.Println("DEL TCP(A): ", i, "L:", TCP_MAP[i].REMOTE[ii].Local, "O:", TCP_MAP[i].REMOTE[ii].Original, "M:", TCP_MAP[i].REMOTE[ii].Mapped)
				delete(TCP_MAP[i].LOCAL, TCP_MAP[i].REMOTE[ii].Local)
				delete(TCP_MAP[i].REMOTE, ii)
				continue
			}

		}
	}
}

func CleanUDPPorts() {
	defer UDP_MAP_LOCK.Unlock()
	defer RecoverAndLogToFile()
	UDP_MAP_LOCK.Lock()

	for i := range UDP_MAP {
		for ii := range UDP_MAP[i].REMOTE {
			if UDP_MAP[i].REMOTE[ii] == nil {
				continue
			}
			if C.DNS1Bytes == i {
				if time.Since(UDP_MAP[i].REMOTE[ii].LastActivity).Seconds() > 10 {
					// log.Println("DEL UDP(DNS): ", i, "L:", UDP_MAP[i].REMOTE[ii].Local, "O:", UDP_MAP[i].REMOTE[ii].Original, "M:", UDP_MAP[i].REMOTE[ii].Mapped)
					delete(UDP_MAP[i].LOCAL, UDP_MAP[i].REMOTE[ii].Local)
					delete(UDP_MAP[i].REMOTE, ii)
					continue
				}
			}

			if time.Since(UDP_MAP[i].REMOTE[ii].LastActivity).Seconds() > 60 {
				// log.Println("DEL UDP(A): ", i, "L:", UDP_MAP[i].REMOTE[ii].Local, "O:", UDP_MAP[i].REMOTE[ii].Original, "M:", UDP_MAP[i].REMOTE[ii].Mapped)
				delete(UDP_MAP[i].LOCAL, UDP_MAP[i].REMOTE[ii].Local)
				delete(UDP_MAP[i].REMOTE, ii)
				continue
			}
		}
	}
}

func InstantlyCleanAllTCPPorts() {
	defer TCP_MAP_LOCK.Unlock()
	defer RecoverAndLogToFile()
	TCP_MAP_LOCK.Lock()

	CreateLog("", "Cleaning up all TCP ports")

	for i := range TCP_MAP {
		TCP_MAP[i] = nil
	}

	TCP_MAP = make(map[[4]byte]*IP)
}

func InstantlyCleanAllUDPPorts() {
	defer UDP_MAP_LOCK.Unlock()
	defer RecoverAndLogToFile()
	UDP_MAP_LOCK.Lock()

	CreateLog("", "Cleaning up all UDP ports")

	for i := range UDP_MAP {
		UDP_MAP[i] = nil
	}

	UDP_MAP = make(map[[4]byte]*IP)
}

// ==========================================
// ==========================================
// ==========================================
// ==========================================
// ==========================================
// ==========================================
// ==========================================
// NEW

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
		time.Sleep(15 * time.Second)
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 9
		}
	}()

	defer RecoverAndLogToFile()
	CleanPortMap(&TCP_o0)
	CleanPortMap(&UDP_o0)
}

func CleanPortMap(protoMap *[256]*O1) {
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

					m.Lock.Lock()
					for ri := range m.REMOTE {
						if time.Since(m.REMOTE[ri].LastActivity).Seconds() > 85 {
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

	done := time.Since(start).Nanoseconds()
	log.Println(" @@@@@ WALK >>> Micro @@@@> ", done, " >> count >> ", count)
}
