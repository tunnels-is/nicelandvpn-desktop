package core

import (
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
	defer TCP_MAP_LOCK.Unlock()
	defer RecoverAndLogToFile()
	TCP_MAP_LOCK.Lock()

	for i := range TCP_MAP {
		for ii := range TCP_MAP[i].REMOTE {
			if TCP_MAP[i].REMOTE[ii] == nil {
				continue
			}

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

			if time.Since(TCP_MAP[i].REMOTE[ii].LastActivity).Seconds() > 120 {
				// log.Println("DEL TCP(A): ", i, "L:", TCP_MAP[i].REMOTE[ii].Local, "O:", TCP_MAP[i].REMOTE[ii].Original, "M:", TCP_MAP[i].REMOTE[ii].Mapped)
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
