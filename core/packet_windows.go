//go:build windows

package core

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/sys/windows"
)

// These variables are used by "VALIDATE_AND_SEND_PACKET_TO_ROUTER"
// The reason they are made global is to reduce memory
// allocations per packet sent

func ReadFromLocalTunnel(MONITOR chan int) {
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
		packetVersion  byte
		readError      error
		packet         []byte
		packetSize     uint16 = 0

		// fullData         []byte
		serializeOptions = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		applicationLayer gopacket.ApplicationLayer
		buffer           gopacket.SerializeBuffer
		parsedPacket     gopacket.Packet
		parsedTCPLayer   *layers.TCP
		parsedUDPLayer   *layers.UDP
		parsedIPLayer    *layers.IPv4

		destinationIP = [4]byte{}
		outgoingPort  *RemotePort

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

		if AS == nil || !GLOBAL_STATE.Connected {
			if time.Since(waitForTimeout).Seconds() > 120 {
				CreateLog("", "ADAPTER: received packet while disconnected. This is most likely a probe packet")
				waitForTimeout = time.Now()
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		EGRESS_PACKETS++

		packetVersion = packet[0] >> 4
		if packetVersion != 4 {
			continue
		}

		destinationIP[0] = packet[16]
		destinationIP[1] = packet[17]
		destinationIP[2] = packet[18]
		destinationIP[3] = packet[19]

		if packet[9] == 6 {
			parsedPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
			parsedIPLayer = parsedPacket.NetworkLayer().(*layers.IPv4)
			applicationLayer = parsedPacket.ApplicationLayer()
			parsedTCPLayer = parsedPacket.TransportLayer().(*layers.TCP)
			if parsedTCPLayer.RST {
				continue
			}

			outgoingPort = GetOutgoingTCPMapping(destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))
			if outgoingPort == nil {
				outgoingPort = CreateTCPMapping(destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))
				if outgoingPort == nil {
					continue
				}
			}

			parsedTCPLayer.SrcPort = layers.TCPPort(outgoingPort.Mapped)

			AS.TCPHeader.DstIP = parsedIPLayer.DstIP
			parsedIPLayer.SrcIP = AS.TCPHeader.SrcIP
			parsedTCPLayer.SetNetworkLayerForChecksum(&AS.TCPHeader)

			buffer = gopacket.NewSerializeBuffer()
			if applicationLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedTCPLayer, gopacket.Payload(applicationLayer.LayerContents()))

			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedTCPLayer)

			}

		} else if packet[9] == 17 {
			parsedPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
			parsedIPLayer = parsedPacket.NetworkLayer().(*layers.IPv4)
			applicationLayer = parsedPacket.ApplicationLayer()
			parsedUDPLayer = parsedPacket.TransportLayer().(*layers.UDP)

			outgoingPort = GetOutgoingUDPMapping(destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))
			if outgoingPort == nil {
				outgoingPort = GetOrCreateUDPMapping(destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))
				if outgoingPort == nil {
					continue
				}
			}

			parsedUDPLayer.SrcPort = layers.UDPPort(outgoingPort.Mapped)
			AS.UDPHeader.DstIP = parsedIPLayer.DstIP
			parsedIPLayer.SrcIP = AS.UDPHeader.SrcIP
			parsedUDPLayer.SetNetworkLayerForChecksum(&AS.UDPHeader)

			buffer = gopacket.NewSerializeBuffer()
			if applicationLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedUDPLayer, gopacket.Payload(applicationLayer.LayerContents()))

			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedUDPLayer)
			}

		} else {
			continue
		}

		if AS.TCPTunnelSocket != nil {

			encryptedPacket = AS.AEAD.Seal(nil, nonce, buffer.Bytes(), nil)

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

func ReadFromRouterSocket(MONITOR chan int) {
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

	if AS == nil || AS.TCPTunnelSocket == nil {
		time.Sleep(100 * time.Millisecond)
		goto WAIT_FOR_TUNNEL
	}

	var (
		err error
		// MIDL int = MIDBufferLength
		lengthBytes = make([]byte, 2)
		DL          uint16
		readBytes   int

		TunnelBuffer = CreateTunnelBuffer()
		ip           = new(layers.IPv4)
		nonce        = make([]byte, chacha20poly1305.NonceSizeX)

		packet                  []byte
		ingressPacket           gopacket.Packet
		buffer                  gopacket.SerializeBuffer
		serializeOptions        = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		appLayer                gopacket.ApplicationLayer
		TCPLayer                *layers.TCP
		UDPLayer                *layers.UDP
		incomingPort            *RemotePort
		inPacket                []byte
		ingressAllocationBuffer []byte
		writeError              error
		ingressPacketLength     int
		sourceIP                = [4]byte{}
	)

	ip.TTL = 120
	ip.DstIP = TUNNEL_ADAPTER_ADDRESS_IP
	ip.Version = 4

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

		sourceIP[0] = packet[12]
		sourceIP[1] = packet[13]
		sourceIP[2] = packet[14]
		sourceIP[3] = packet[15]
		ip.SrcIP = net.IP{sourceIP[0], sourceIP[1], sourceIP[2], sourceIP[3]}

		ingressPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
		buffer = gopacket.NewSerializeBuffer()
		appLayer = ingressPacket.ApplicationLayer()

		if packet[9] == 6 {
			ip.Protocol = 6
			TCPLayer = ingressPacket.TransportLayer().(*layers.TCP)

			incomingPort = GetTCPMapping(sourceIP, uint16(TCPLayer.DstPort))
			if incomingPort == nil {
				continue
			}

			TCPLayer.DstPort = layers.TCPPort(incomingPort.Local)
			TCPLayer.SetNetworkLayerForChecksum(ip)

			if appLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, TCPLayer, gopacket.Payload(appLayer.LayerContents()))
			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, TCPLayer)
			}

		} else if packet[9] == 17 {
			ip.Protocol = 17
			UDPLayer = ingressPacket.TransportLayer().(*layers.UDP)
			UDPLayer.SetNetworkLayerForChecksum(ip)

			incomingPort = GetUDPMapping(sourceIP, uint16(UDPLayer.DstPort))

			if incomingPort == nil {
				continue
			}

			UDPLayer.DstPort = layers.UDPPort(incomingPort.Local)

			if appLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, UDPLayer, gopacket.Payload(appLayer.LayerContents()))
			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, ip, UDPLayer)
			}

		}

		inPacket = buffer.Bytes()
		ingressPacketLength = len(inPacket)

		ingressAllocationBuffer, writeError = A.AllocateSendPacket(ingressPacketLength)
		if writeError != nil {
			BUFFER_ERROR = true
			CreateErrorLog("", "Send: ", writeError)
			return
		}

		copy(ingressAllocationBuffer, inPacket)
		A.SendPacket(ingressAllocationBuffer)
		CURRENT_DBBS += ingressPacketLength

		inPacket = nil
	}
}
