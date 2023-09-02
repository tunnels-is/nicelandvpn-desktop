//go:build windows

package core

import (
	"encoding/base64"
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	packets "github.com/tunnels-is/nicelandvpn-desktop/packet"
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

		parsedDNSLayer      *layers.DNS
		domain              string
		isCustomDNS         bool
		domainIPS           []net.IP
		domainTXTS          []string
		dnsOldIP            net.IP
		dnsOldPort          layers.UDPPort
		dnsPacket           []byte
		dnsAllocationBuffer []byte

		shouldDropDNS bool

		natOK  bool
		NAT_IP [4]byte

		isDNSLayer bool = false

		destinationIP = [4]byte{}
		// outgoingPort  *RemotePort
		OP *RP

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

		packetVersion = packet[0] >> 4
		if packetVersion != 4 {
			continue
		}

		destinationIP[0] = packet[16]
		destinationIP[1] = packet[17]
		destinationIP[2] = packet[18]
		destinationIP[3] = packet[19]

		if packet[9] == 6 {

			parsedPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.NoCopy)

			if destinationIP == [4]byte{22, 22, 22, 22} {
				log.Println(parsedPacket)
			}

			parsedIPLayer = parsedPacket.NetworkLayer().(*layers.IPv4)
			applicationLayer = parsedPacket.ApplicationLayer()
			parsedTCPLayer = parsedPacket.TransportLayer().(*layers.TCP)
			if parsedTCPLayer.RST {
				continue
			}

			OP = CreateOrGetPortMapping(&TCP_o0, destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))
			if OP == nil {
				continue
			}

			// outgoingPort = GetOutgoingTCPMapping(destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))
			// if outgoingPort == nil {
			// 	outgoingPort = CreateTCPMapping(destinationIP, uint16(parsedTCPLayer.SrcPort), uint16(parsedTCPLayer.DstPort))
			// 	if outgoingPort == nil {
			// 		continue
			// 	}
			// }

			NAT_IP, natOK = AS.AP.NAT_CACHE[destinationIP]
			if natOK {
				// CreateLog("NAT", "FOUND NAT: ", NAT_IP)
				AS.TCPHeader.DstIP = net.IP{NAT_IP[0], NAT_IP[1], NAT_IP[2], NAT_IP[3]}
				parsedIPLayer.DstIP = net.IP{NAT_IP[0], NAT_IP[1], NAT_IP[2], NAT_IP[3]}
			} else {
				AS.TCPHeader.DstIP = parsedIPLayer.DstIP
			}

			if destinationIP == [4]byte{22, 22, 22, 22} {
				packets.ParsePacket(packet, OP.Mapped, AS.TCPHeader.SrcIP, NAT_IP)
			}

			// parsedTCPLayer.SrcPort = layers.TCPPort(outgoingPort.Mapped)
			parsedTCPLayer.SrcPort = layers.TCPPort(OP.Mapped)

			parsedIPLayer.SrcIP = AS.TCPHeader.SrcIP

			parsedTCPLayer.SetNetworkLayerForChecksum(&AS.TCPHeader)

			buffer = gopacket.NewSerializeBuffer()
			if applicationLayer != nil {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedTCPLayer, gopacket.Payload(applicationLayer.LayerContents()))

			} else {
				gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedTCPLayer)
			}
			// if destinationIP == [4]byte{184, 186, 76, 193} {
			// 	testPacket := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeIPv4, gopacket.Default)
			// 	CreateLog("NAT", testPacket)
			// 	continue
			// }

		} else if packet[9] == 17 {

			parsedPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
			parsedIPLayer = parsedPacket.NetworkLayer().(*layers.IPv4)
			applicationLayer = parsedPacket.ApplicationLayer()
			parsedUDPLayer = parsedPacket.TransportLayer().(*layers.UDP)

			parsedDNSLayer, isDNSLayer = applicationLayer.(*layers.DNS)
			if isDNSLayer {
				shouldDropDNS = false
				// DNS BLOCK LIST PARSING
				// log.Println(parsedDNSLayer)
				// if len(parsedDNSLayer.Questions) > 0 {
				// 	DNSQuestionDomain = string(parsedDNSLayer.Questions[0].Name)
				// 	// log.Println("Searching in blocklist: ", DNSQuestionDomain)
				// 	_, DomainIsBlocked = BlockedDomainMap[DNSQuestionDomain]
				// 	if DomainIsBlocked {
				// 		log.Println("IS BLOCKED: ", DNSQuestionDomain)
				// 		// DomainIsBlocked = false
				// 		continue
				// 	}
				// }

				for _, v := range parsedDNSLayer.Questions {
					domain = string(v.Name)
					if GLOBAL_STATE.DNSWhitelistEnabled {
						if !IsDomainAllowed(domain) {
							// TODO .. reply with bogus IP
							shouldDropDNS = true
							goto DROP
						}
					}

					if v.Type == 28 { // AAA RECORD

						// CreateLog("DNS", "Question AAAA Record: ", string(v.Name))
						if GLOBAL_STATE.DNSCaptureEnabled {
							CaptureDNS(string(v.Name))
						}

					} else if v.Type == 1 { // A RECORD
						if GLOBAL_STATE.DNSCaptureEnabled {
							CaptureDNS(string(v.Name))
						}

						// CreateLog("DNS", "Question A Record: ", string(v.Name))
						domainIPS = DNSAMapping(domain)
						for _, ip := range domainIPS {
							isCustomDNS = true
							parsedDNSLayer.Answers = append(parsedDNSLayer.Answers, layers.DNSResourceRecord{
								Name:  v.Name,
								Type:  v.Type,
								Class: v.Class,
								TTL:   30,
								IP:    ip,
							})
							parsedDNSLayer.ANCount++
						}

					} else if v.Type == 16 { // TXT RECORD
						// CreateLog("DNS", "Question TXT Record: ", string(v.Name))
						domainTXTS = DNSTXTMapping(domain)
						for _, txt := range domainTXTS {
							// txtb := make([][]byte, 1)
							// txtb[0] = []byte(txt)
							isCustomDNS = true
							parsedDNSLayer.Answers = append(parsedDNSLayer.Answers, layers.DNSResourceRecord{
								Name:  v.Name,
								Type:  v.Type,
								Class: v.Class,
								TTL:   30,
								TXTs:  [][]byte{[]byte(txt)},
								TXT:   []byte(base64.StdEncoding.EncodeToString([]byte(txt))),
							})
							parsedDNSLayer.ANCount++
						}

					}

				}

			DROP:
				if shouldDropDNS {
					continue
				}

				if isCustomDNS {
					isCustomDNS = false
					parsedDNSLayer.QR = true

					dnsOldIP = parsedIPLayer.DstIP
					parsedIPLayer.DstIP = parsedIPLayer.SrcIP
					parsedIPLayer.SrcIP = dnsOldIP

					dnsOldPort = parsedUDPLayer.DstPort
					parsedUDPLayer.DstPort = parsedUDPLayer.SrcPort
					parsedUDPLayer.SrcPort = dnsOldPort

					parsedUDPLayer.SetNetworkLayerForChecksum(parsedIPLayer)
					buffer = gopacket.NewSerializeBuffer()

					gopacket.SerializeLayers(buffer, serializeOptions, parsedIPLayer, parsedUDPLayer, parsedDNSLayer)
					dnsPacket = buffer.Bytes()

					dnsAllocationBuffer, writeError = A.AllocateSendPacket(len(dnsPacket))
					if writeError != nil {
						BUFFER_ERROR = true
						CreateErrorLog("", "Send: ", writeError)
						return
					}

					copy(dnsAllocationBuffer, dnsPacket)
					A.SendPacket(dnsAllocationBuffer)

					continue
				}

			}

			OP = CreateOrGetPortMapping(&UDP_o0, destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))
			if OP == nil {
				continue
			}
			// outgoingPort = GetOutgoingUDPMapping(destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))
			// if outgoingPort == nil {
			// 	outgoingPort = GetOrCreateUDPMapping(destinationIP, uint16(parsedUDPLayer.SrcPort), uint16(parsedUDPLayer.DstPort))
			// 	if outgoingPort == nil {
			// 		continue
			// 	}
			// }

			NAT_IP, natOK = AS.AP.NAT_CACHE[destinationIP]
			if natOK {
				AS.UDPHeader.DstIP = net.IP{NAT_IP[0], NAT_IP[1], NAT_IP[2], NAT_IP[3]}
				parsedIPLayer.DstIP = net.IP{NAT_IP[0], NAT_IP[1], NAT_IP[2], NAT_IP[3]}
			} else {
				AS.UDPHeader.DstIP = parsedIPLayer.DstIP
			}

			// parsedUDPLayer.SrcPort = layers.UDPPort(outgoingPort.Mapped)
			parsedUDPLayer.SrcPort = layers.UDPPort(OP.Mapped)
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

			// var OUT = make([]byte, 0)
			OUT := buffer.Bytes()
			// //185.186.76.193
			if destinationIP == [4]byte{22, 22, 22, 22} {
				// log.Println("OUT IP: ", destinationIP)
				// log.Println(AS.AP.NAT_CACHE[destinationIP])
				testPacket := gopacket.NewPacket(OUT, layers.LayerTypeIPv4, gopacket.Default)
				log.Println(testPacket)
			}

			// encryptedPacket = AS.AEAD.Seal(nil, nonce, buffer.Bytes(), nil)
			encryptedPacket = AS.AEAD.Seal(nil, nonce, OUT, nil)

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

	if AS == nil || AS.AP == nil || AS.TCPTunnelSocket == nil {
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

		packet           []byte
		ingressPacket    gopacket.Packet
		buffer           gopacket.SerializeBuffer
		serializeOptions = gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
		appLayer         gopacket.ApplicationLayer
		TCPLayer         *layers.TCP
		UDPLayer         *layers.UDP
		// incomingPort            *RemotePort
		incP                    *RP
		inPacket                []byte
		ingressAllocationBuffer []byte
		writeError              error
		ingressPacketLength     int
		sourceIP                = [4]byte{}

		natOK  bool
		NAT_IP [4]byte
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

		// if sourceIP == [4]byte{184, 186, 76, 193} || sourceIP == [4]byte{185, 186, 76, 193} {
		// 	log.Println("IN IP: ", sourceIP)
		// 	log.Println(AS.AP.REVERSE_NAT_CACHE[sourceIP])
		// }

		NAT_IP, natOK = AS.AP.REVERSE_NAT_CACHE[sourceIP]
		if natOK {
			// CreateLog("NAT", "FOUND REVERSE NAT: ", NAT_IP)
			ip.SrcIP = net.IP{NAT_IP[0], NAT_IP[1], NAT_IP[2], NAT_IP[3]}
			sourceIP[0] = NAT_IP[0]
			sourceIP[1] = NAT_IP[1]
			sourceIP[2] = NAT_IP[2]
			sourceIP[3] = NAT_IP[3]
		} else {
			ip.SrcIP = net.IP{sourceIP[0], sourceIP[1], sourceIP[2], sourceIP[3]}
		}

		ingressPacket = gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.NoCopy)
		buffer = gopacket.NewSerializeBuffer()
		appLayer = ingressPacket.ApplicationLayer()

		if packet[9] == 6 {
			ip.Protocol = 6
			TCPLayer = ingressPacket.TransportLayer().(*layers.TCP)

			incP = GetIngressPortMapping(&TCP_o0, sourceIP, uint16(TCPLayer.DstPort))
			if incP == nil {
				continue
			}
			// incomingPort = GetTCPMapping(sourceIP, uint16(TCPLayer.DstPort))
			// if incomingPort == nil {
			// 	continue
			// }

			// TCPLayer.DstPort = layers.TCPPort(incomingPort.Local)
			TCPLayer.DstPort = layers.TCPPort(incP.Local)
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

			incP = GetIngressPortMapping(&UDP_o0, sourceIP, uint16(UDPLayer.DstPort))
			if incP == nil {
				continue
			}
			// incomingPort = GetUDPMapping(sourceIP, uint16(UDPLayer.DstPort))
			// if incomingPort == nil {
			// 	continue
			// }

			// UDPLayer.DstPort = layers.UDPPort(incomingPort.Local)
			UDPLayer.DstPort = layers.UDPPort(incP.Local)

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
