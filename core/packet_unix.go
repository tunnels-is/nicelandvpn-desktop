//go:build freebsd || linux || openbsd || darwin

package core

import (
	"math"
	"time"
)

func (V *VPNConnection) ReadFromLocalSocket() {
	defer func() {
		RecoverAndLogToFile()
		CreateErrorLog("", "tun tap listener exiting:", V.Name)
	}()

	var (
		err            error
		waitFortimeout = time.Now()
		packetLength   int

		packet []byte

		writeError   error
		writtenBytes int

		sendLocal  bool
		sendRemote bool
		tempBytes  = make([]byte, math.MaxUint16)
		encBytes   = make([]byte, math.MaxUint16)
	)

	for {
		packetLength, err = V.Tun.Read(tempBytes)
		if err != nil {
			CreateLog("general", err, "error in interface reader loop")
			return
		}
		// start := time.Now()

		if packetLength == 0 {
			CreateLog("", "Read size was 0")
			continue
		}

		if !V.Connected {
			// fmt.Println(tempPacket[:packetLength])
			if time.Since(waitFortimeout).Seconds() > 120 {
				CreateLog("", "VPN: received packet while disconnected. This is most likely a probe packet")
				waitFortimeout = time.Now()
			}
			continue
		}
		packet = tempBytes[:packetLength]

		V.EgressPackets++

		sendRemote, sendLocal = V.ProcessEgressPacket(&packet)
		if !sendLocal && !sendRemote {
			// log.Println("NOT SENDING EGRESS PACKET - PROTO:", packet[9])
			continue
		} else if sendLocal {

			writtenBytes, writeError = V.Tun.Write(packet)
			if writeError != nil {
				CreateErrorLog("", "Send: ", writeError)
			}

			continue
		}

		writtenBytes, err = V.EVPNS.Write(encBytes, packet, len(packet))
		if err != nil {
			_ = V.EVPNS.SOCKET.Close()
			return
		}
		V.EgressBytes += writtenBytes
		// end := time.Since(start).Microseconds()
		// fmt.Println("OUT:", end)
	}
}

func (V *VPNConnection) ReadFromRouterSocket() {
	defer func() {
		RecoverAndLogToFile()
		CreateErrorLog("", "Router tunnel listener exiting:", V.Name)
	}()

	var (
		writeErr          error
		readErr           error
		writtenBytes      int
		encryptedReceiver = make([]byte, math.MaxUint16)
		decryptedReceiver = make([]byte, math.MaxUint16)
		packet            []byte
	)

	for {

		_, packet, readErr = V.EVPNS.Read(encryptedReceiver, decryptedReceiver)
		if readErr != nil {
			CreateErrorLog("", "")
			return
		}
		// start := time.Now()
		V.IngressPackets++
		// TODO !!!!!!!!!!!!!!!!!
		// 	GLOBAL_STATE.PingReceivedFromRouter = time.Now()
		// 	continue

		if !V.ProcessIngressPacket(packet) {
			// log.Println("NOT SENDING INGRESS PACKET - PROTO:", packet[9])
			continue
		}

		writtenBytes, writeErr = V.Tun.Write(packet)
		if writeErr != nil {
			CreateErrorLog("", "Send: ", writeErr)
			return
		}
		V.IngressBytes += writtenBytes
		// end := time.Since(start).Microseconds()
		// fmt.Println("IN:", end)

	}
}
