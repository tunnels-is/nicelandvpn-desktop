//go:build freebsd || linux || openbsd || darwin

package core

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

var PREVDNS net.IP

func ReadFromLocalSocket(MONITOR chan int) {
	defer func() {
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 4
		} else {
			CreateLog("general", "tunnel interface loop has exited")
		}
	}()
	defer RecoverAndLogToFile()

	IS_UNIX = true

	var (
		err            error
		waitFortimeout = time.Now()
		packetLength   int

		tempPacket = make([]byte, 65600)
		packet     []byte

		encryptedPacket []byte
		lengthBytes     = make([]byte, 2)
		nonce           = make([]byte, chacha20poly1305.NonceSizeX)
		writeError      error
		writtenBytes    int

		sendLocal  bool
		sendRemote bool
	)

WAITFORDEVICE:
	if !GLOBAL_STATE.TunnelInitialized {
		if time.Since(waitFortimeout).Seconds() > 120 {
			CreateLog("", "Adapter reader not initialized, waiting for connection")
			waitFortimeout = time.Now()
		}
		time.Sleep(500 * time.Millisecond)
		goto WAITFORDEVICE
	}

	for {

		packetLength, err = A.Interface.Read(tempPacket)
		if err != nil {
			BUFFER_ERROR = true
			CreateLog("general", err, "error in interface reader loop")
			return
		}

		if packetLength == 0 {
			CreateLog("", "Read size was 0")
			continue
		}

		if AS == nil || AS.AP == nil || !GLOBAL_STATE.Connected {
			if time.Since(waitFortimeout).Seconds() > 120 {
				CreateLog("", "ADAPTER: received packet while disconnected. This is most likely a probe packet")
				waitFortimeout = time.Now()
			}
			continue
		}
		packet = tempPacket[:packetLength]

		EGRESS_PACKETS++

		sendRemote, sendLocal = ProcessEgressPacket(&packet)
		if !sendLocal && !sendRemote {
			// log.Println("NOT SENDING EGRESS PACKET - PROTO:", packet[9])
			continue
		} else if sendLocal {

			writtenBytes, writeError = A.Interface.Write(packet)
			if writeError != nil {
				CreateErrorLog("", "Send: ", writeError)
			}

			continue
		}

		if AS.TCPTunnelSocket != nil {

			encryptedPacket = AS.AEAD.Seal(nil, nonce, packet, nil)

			binary.BigEndian.PutUint16(lengthBytes, uint16(len(encryptedPacket)))

			writtenBytes, writeError = AS.TCPTunnelSocket.Write(append(lengthBytes, encryptedPacket...))
			if writeError != nil {
				BUFFER_ERROR = true
				_ = AS.TCPTunnelSocket.Close()
				return
			}

			CURRENT_UBBS += writtenBytes
			lengthBytes = make([]byte, 2)
		} else {
			GLOBAL_STATE.Connected = false
		}

	}
}

func ReadFromRouterSocket(MONITOR chan int) {
	defer func() {
		if !GLOBAL_STATE.Exiting {
			CreateErrorLog("", "Router tunnel listener exiting")
			MONITOR <- 2
		}
	}()
	defer RecoverAndLogToFile()

WAIT_FOR_TUNNEL:
	if GLOBAL_STATE.ActiveRouter == nil {
		time.Sleep(500 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	}

	if AS.TCPTunnelSocket == nil {
		time.Sleep(500 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	} else {
		AS.TCPTunnelSocket.SetReadDeadline(time.Time{})
	}

	var (
		writeErr     error
		readErr      error
		writtenBytes int
		lengthBytes  = make([]byte, 2)
		DL           uint16
		readBytes    int

		tunnelBuffer = CreateTunnelBuffer()
		nonce        = make([]byte, chacha20poly1305.NonceSizeX)
		encErr       error

		packet []byte
	)

	for {

		readBytes, readErr = io.ReadAtLeast(AS.TCPTunnelSocket, lengthBytes[:2], 2)
		if readErr != nil {
			if !IGNORE_NEXT_BUFFER_ERROR {
				CreateErrorLog("", "Read: ", readErr)
				BUFFER_ERROR = true
			} else {
				IGNORE_NEXT_BUFFER_ERROR = false
			}
			return
		}

		if readBytes != 2 {
			CreateErrorLog("", "TUNNEL SMALL READ ERROR: ", AS.TCPTunnelSocket.RemoteAddr())
			return
		}

		INGRESS_PACKETS++
		DL = binary.BigEndian.Uint16(lengthBytes[0:2])

		if DL == CODE_CLIENT_new_ping {
			GLOBAL_STATE.PingReceivedFromRouter = time.Now()
			continue
		}

		_, readErr = io.ReadAtLeast(AS.TCPTunnelSocket, tunnelBuffer[:DL], int(DL))
		if readErr != nil {
			CreateErrorLog("", "TUNNEL DATA READ ERROR: ", readErr)
			return
		}

		// packet = tunnelBuffer[MIDL : MIDL+DL]
		packet, encErr = AS.AEAD.Open(nil, nonce, tunnelBuffer[:DL], nil)
		if encErr != nil {
			CreateErrorLog("", "Encryption: ", encErr)
			continue
		}

		if !ProcessIngressPacket(packet) {
			// log.Println("NOT SENDING INGRESS PACKET - PROTO:", packet[9])
			continue
		}

		writtenBytes, writeErr = A.Interface.Write(packet)
		if writeErr != nil {
			CreateErrorLog("", "Send: ", writeErr)
		}
		CURRENT_DBBS += writtenBytes

		packet = nil

	}
}
