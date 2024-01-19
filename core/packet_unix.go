//go:build freebsd || linux || openbsd || darwin

package core

import (
	"fmt"
	"log"
	"math"
	"time"
)

func (V *VPNConnection) ReadFromLocalSocket() {
	defer func() {
		RecoverAndLogToFile()
		CreateErrorLog("", "tun tap listener exiting:", V.Meta.Tag)
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
	)

	for {
		packetLength, err = V.Tun.Read(tempBytes)
		// fmt.Println(tempBytes[:packetLength])
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
			log.Println("NOT SENDING EGRESS PACKET - PROTO:", packet[9])
			continue
		} else if sendLocal {

			fmt.Println("SEND LOCAL:", packet)
			writtenBytes, writeError = V.Tun.Write(packet)
			if writeError != nil {
				CreateErrorLog("", "Send: ", writeError)
			}

			continue
		}
		// fmt.Println("OUT", packet)

		writtenBytes, err = V.EVPNS.Write(packet)
		if err != nil {
			_ = V.EVPNS.SOCKET.Close()
			return
		}
		// fmt.Println("OUT", writtenBytes)
		V.EgressBytes += writtenBytes
		// end := time.Since(start).Microseconds()
		// fmt.Println("OUT:", end)
	}
}

func (V *VPNConnection) ReadFromRouterSocket() {
	defer func() {
		RecoverAndLogToFile()
		CreateErrorLog("", "Router tunnel listener exiting:", V.Meta.Tag)
	}()

	var (
		writeErr      error
		readErr       error
		receivedBytes int
		// encryptedReceiver = make([]byte, math.MaxUint16)
		// decryptedReceiver = make([]byte, math.MaxUint16)
		packet []byte
	)

	for {

		receivedBytes, packet, readErr = V.EVPNS.Read()
		if readErr != nil {
			CreateErrorLog("", "error reading from node socket", readErr, receivedBytes, packet)
			return
		}
		// fmt.Println("IN:", packet)

		if packet[0] == CODE_pingPong {
			V.PingReceived = time.Now()
			continue
		}

		V.IngressPackets++
		V.IngressBytes += receivedBytes
		if !V.ProcessIngressPacket(packet) {
			log.Println("NOT SENDING INGRESS PACKET - PROTO:", packet[9])
			continue
		}
		// fmt.Println("INP:", packet)

		_, writeErr = V.Tun.Write(packet)
		if writeErr != nil {
			CreateErrorLog("", "Send: ", writeErr)
			return
		}
		// fmt.Println("IN", receivedBytes)
	}
}
