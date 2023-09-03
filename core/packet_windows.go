//go:build windows

package core

import (
	"encoding/binary"
	"io"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/sys/windows"
)

// These variables are used by "VALIDATE_AND_SEND_PACKET_TO_ROUTER"
// The reason they are made global is to reduce memory
// allocations per packet sent

func ReadFromLocalTunnel_NEW(MONITOR chan int) {
	defer func() {
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 4
		} else {
			CreateErrorLog("", "Adapter reader exiting")
		}
	}()
	defer RecoverAndLogToFile()

	var (
		waitForTimeout = time.Now()
		readError      error
		packet         []byte
		packetSize     uint16 = 0

		// fullData         []byte
		// buffer gopacket.SerializeBuffer

		// destinationIP = [4]byte{}
		// outgoingPort  *RemotePort

		encryptedPacket []byte
		lengthBytes     = make([]byte, 2)
		nonce           = make([]byte, chacha20poly1305.NonceSizeX)
		writtenBytes    int
		writeError      error
	)

WAITFORDEVICE:
	if !A.Initialized {
		if time.Since(waitForTimeout).Seconds() > 120 {
			CreateLog("", "Adapter reader not initialized, waiting for connection")
			waitForTimeout = time.Now()
		}
		time.Sleep(500 * time.Millisecond)
		goto WAITFORDEVICE
	}

	for {
		// if packetSize > 0 {
		if packet != nil {
			A.ReleaseReceivePacket(packet)
		}

		// packet = nil
		if GLOBAL_STATE.Exiting {
			CreateLog("", "nicelandVPN is exiting, closing adapter reader")
			return
		}

		packet, packetSize, readError = A.ReceivePacket()

		if readError == windows.ERROR_NO_MORE_ITEMS {
			if time.Since(waitForTimeout).Seconds() > 120 {
				CreateLog("file", "ADAPTER: no packets in buffer, waiting for packets")
				waitForTimeout = time.Now()
			}
			time.Sleep(200 * time.Microsecond)
			continue
		} else if readError == windows.ERROR_HANDLE_EOF {
			CreateErrorLog("", "ADAPTER 1: ", readError)
			BUFFER_ERROR = true
			return
		} else if readError == windows.ERROR_INVALID_DATA {
			CreateErrorLog("", "ADAPTER 2: ", readError)
			BUFFER_ERROR = true
			return
		} else if readError != nil {
			CreateErrorLog("", "ADAPTER 3: ", readError)
			BUFFER_ERROR = true
			return
		}

		if packetSize == 0 {
			CreateLog("", "Read size was 0")
			continue
		}

		if AS == nil || AS.AP == nil || !GLOBAL_STATE.Connected {
			if time.Since(waitForTimeout).Seconds() > 120 {
				CreateLog("", "ADAPTER: received packet while disconnected. This is most likely a probe packet")
				waitForTimeout = time.Now()
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		EGRESS_PACKETS++

		// log.Println(" BEFORE ==========================================")
		// fmt.Println(packet)
		if !ProcessEgressPacket(packet) {
			// log.Println("NOT SENDING EGRESS PACKET - PROTO:", packet[9])
			continue
		}
		// log.Println(" AFTER ==========================================")
		// fmt.Println(packet)

		if AS.TCPTunnelSocket != nil {

			// testPacket := gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
			// log.Println(" EGRESS ==========================================")
			// fmt.Println(testPacket)
			// log.Println(" EGRESS ==========================================")

			// encryptedPacket = AS.AEAD.Seal(nil, nonce, buffer.Bytes(), nil)
			encryptedPacket = AS.AEAD.Seal(nil, nonce, packet, nil)

			binary.BigEndian.PutUint16(lengthBytes, uint16(len(encryptedPacket)))

			writtenBytes, writeError = AS.TCPTunnelSocket.Write(append(lengthBytes, encryptedPacket...))
			if writeError != nil {
				CreateErrorLog("", "Write: ", writeError)
				BUFFER_ERROR = true
				_ = AS.TCPTunnelSocket.Close()
				return
			}

			CURRENT_UBBS += writtenBytes
			lengthBytes = make([]byte, 2)
		} else {
			GLOBAL_STATE.Connected = false
		}

		// fullData = nil
	}
}

func ReadFromRouterSocket_NEW(MONITOR chan int) {
	defer func() {
		if !GLOBAL_STATE.Exiting {
			MONITOR <- 2
		} else {
			CreateErrorLog("", "VPN listener exiting")
		}
	}()
	defer RecoverAndLogToFile()

WAIT_FOR_TUNNEL:
	if GLOBAL_STATE.ActiveRouter == nil {
		time.Sleep(100 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	}

	if AS == nil || AS.AP == nil || AS.TCPTunnelSocket == nil {
		time.Sleep(100 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	}

	var (
		err         error
		lengthBytes = make([]byte, 2)
		DL          uint16
		readBytes   int

		TunnelBuffer = CreateTunnelBuffer()
		nonce        = make([]byte, chacha20poly1305.NonceSizeX)

		packet                  []byte
		ingressAllocationBuffer []byte
		writeError              error
		ingressPacketLength     int
	)

	AS.TCPTunnelSocket.SetReadDeadline(time.Time{})

	for {
		// _, DL, err = ReadMIDAndDataFromBuffer(AS.TCPTunnelSocket, TunnelBuffer)
		readBytes, err = io.ReadAtLeast(AS.TCPTunnelSocket, lengthBytes[:2], 2)
		if err != nil {
			if !IGNORE_NEXT_BUFFER_ERROR {
				CreateErrorLog("", "Read: ", err)
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

		_, err = io.ReadAtLeast(AS.TCPTunnelSocket, TunnelBuffer[:DL], int(DL))
		if err != nil {
			CreateErrorLog("", "TUNNEL DATA READ ERROR: ", err)
			return
		}

		packet, err = AS.AEAD.Open(nil, nonce, TunnelBuffer[:DL], nil)
		if err != nil {
			CreateErrorLog("", "Encryption: ", err)
			continue
		}

		if !ProcessIngressPacket(packet) {
			// log.Println("NOT SENDING INGRESS PACKET - PROTO:", packet[9])
			continue
		}

		// testPacket := gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
		// log.Println(" INGRESS ==========================================")
		// fmt.Println(testPacket)
		// log.Println(" INGRESS ==========================================")

		ingressPacketLength = len(packet)

		ingressAllocationBuffer, writeError = A.AllocateSendPacket(ingressPacketLength)
		if writeError != nil {
			BUFFER_ERROR = true
			CreateErrorLog("", "Send: ", writeError)
			return
		}

		copy(ingressAllocationBuffer, packet)
		A.SendPacket(ingressAllocationBuffer)
		CURRENT_DBBS += ingressPacketLength

		packet = nil
	}
}
