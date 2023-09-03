package core

import (
	"encoding/binary"
	"log"
)

var (
	EP_Version  byte
	EP_Protocol byte

	EP_DstIP [4]byte

	EP_IPv4HeaderLength byte
	EP_IPv4Header       []byte
	EP_TPHeader         []byte

	EP_SrcPort    [2]byte
	EP_DstPort    [2]byte
	EP_MappedPort *RP

	EP_NAT_IP [4]byte
	EP_NAT_OK bool

	// This IP gets over-written on connect
	EP_VPNSrcIP [4]byte
)

func ProcessEgressPacket(packet []byte) bool {

	EP_Version = packet[0] >> 4
	if EP_Version != 4 {
		return false
	}

	EP_Protocol = packet[9]
	if EP_Protocol != 6 && EP_Protocol != 17 {
		return false
	}

	// TODO: Check for reset flag
	// TODO: Check for reset flag
	// TODO: Check for reset flag

	EP_DstIP[0] = packet[16]
	EP_DstIP[1] = packet[17]
	EP_DstIP[2] = packet[18]
	EP_DstIP[3] = packet[19]

	// Get the full IPv4Header length in bytes
	EP_IPv4HeaderLength = (packet[0] << 4 >> 4) * 32 / 8

	EP_IPv4Header = packet[:EP_IPv4HeaderLength]
	EP_TPHeader = packet[EP_IPv4HeaderLength:]

	EP_SrcPort[0] = EP_TPHeader[0]
	EP_SrcPort[1] = EP_TPHeader[1]

	EP_DstPort[0] = EP_TPHeader[2]
	EP_DstPort[1] = EP_TPHeader[3]

	// CUSTOM DNS
	// https://stackoverflow.com/questions/7565300/identifying-dns-packets
	if EP_DstPort[1] == 53 {
		// log.Println("IS DNS ????", EP_DstIP, EP_DstPort)
		if ProcessDNSQuery(EP_TPHeader) {
			// ????
			// Flip Ports
			// Flip IPs
			// Write packet back to local interface
			// log.Println("RETURNING CUSTOM DNS RESPONSE")
			return false
		}
	}

	if EP_Protocol == 6 {

		EP_MappedPort = CreateOrGetPortMapping(&TCP_o0, EP_DstIP, EP_SrcPort, EP_DstPort)
		if EP_MappedPort == nil {
			log.Println("NO TCP PORT MAPPING", EP_DstIP, EP_SrcPort, EP_DstPort)
			return false
		}

	} else if EP_Protocol == 17 {

		EP_MappedPort = CreateOrGetPortMapping(&UDP_o0, EP_DstIP, EP_SrcPort, EP_DstPort)
		if EP_MappedPort == nil {
			log.Println("NO UDP PORT MAPPING", EP_DstIP, EP_SrcPort, EP_DstPort)
			return false
		}

	}

	EP_TPHeader[0] = EP_MappedPort.Mapped[0]
	EP_TPHeader[1] = EP_MappedPort.Mapped[1]

	EP_NAT_IP, EP_NAT_OK = AS.AP.NAT_CACHE[EP_DstIP]
	if EP_NAT_OK {
		// log.Println("FOUND NAT", EP_DstIP, EP_NAT_IP)
		EP_IPv4Header[16] = EP_NAT_IP[0]
		EP_IPv4Header[17] = EP_NAT_IP[1]
		EP_IPv4Header[18] = EP_NAT_IP[2]
		EP_IPv4Header[19] = EP_NAT_IP[3]
	}

	EP_IPv4Header[12] = EP_VPNSrcIP[0]
	EP_IPv4Header[13] = EP_VPNSrcIP[1]
	EP_IPv4Header[14] = EP_VPNSrcIP[2]
	EP_IPv4Header[15] = EP_VPNSrcIP[3]

	RecalculateAndReplaceIPv4HeaderChecksum(EP_IPv4Header)
	RecalculateAndReplaceTransportChecksum(EP_IPv4Header, EP_TPHeader)

	return true
}

var (
	IP_Version  byte
	IP_Protocol byte

	IP_DstIP [4]byte
	IP_SrcIP [4]byte

	IP_IPv4HeaderLength byte
	IP_IPv4Header       []byte
	IP_TPHeader         []byte

	IP_SrcPort    [2]byte
	IP_DstPort    [2]byte
	IP_MappedPort *RP

	IP_NAT_IP [4]byte
	IP_NAT_OK bool

	// This IP gets over-written on connect
	// IP_VPNSrcIP [4]byte
	IP_InterfaceIP [4]byte
)

func ProcessIngressPacket(packet []byte) bool {

	IP_SrcIP[0] = packet[12]
	IP_SrcIP[1] = packet[13]
	IP_SrcIP[2] = packet[14]
	IP_SrcIP[3] = packet[15]

	IP_DstIP[0] = packet[16]
	IP_DstIP[1] = packet[17]
	IP_DstIP[2] = packet[18]
	IP_DstIP[3] = packet[19]

	IP_Protocol = packet[9]

	IP_IPv4HeaderLength = (packet[0] << 4 >> 4) * 32 / 8
	IP_IPv4Header = packet[:IP_IPv4HeaderLength]
	IP_TPHeader = packet[IP_IPv4HeaderLength:]

	IP_DstPort[0] = IP_TPHeader[2]
	IP_DstPort[1] = IP_TPHeader[3]

	IP_NAT_IP, IP_NAT_OK = AS.AP.REVERSE_NAT_CACHE[IP_SrcIP]
	if IP_NAT_OK {
		// log.Println("FOUND INGRESS NAT", IP_DstIP, IP_NAT_IP)
		IP_IPv4Header[12] = IP_NAT_IP[0]
		IP_IPv4Header[13] = IP_NAT_IP[1]
		IP_IPv4Header[14] = IP_NAT_IP[2]
		IP_IPv4Header[15] = IP_NAT_IP[3]
	}

	if IP_Protocol == 6 {

		IP_MappedPort = GetIngressPortMapping(&TCP_o0, IP_SrcIP, IP_DstPort)
		if IP_MappedPort == nil {
			return false
		}

	} else if IP_Protocol == 17 {

		IP_MappedPort = GetIngressPortMapping(&UDP_o0, IP_SrcIP, IP_DstPort)
		if IP_MappedPort == nil {
			return false
		}

	}

	IP_TPHeader[2] = IP_MappedPort.Local[0]
	IP_TPHeader[3] = IP_MappedPort.Local[1]

	IP_IPv4Header[16] = IP_InterfaceIP[0]
	IP_IPv4Header[17] = IP_InterfaceIP[1]
	IP_IPv4Header[18] = IP_InterfaceIP[2]
	IP_IPv4Header[19] = IP_InterfaceIP[3]

	RecalculateAndReplaceIPv4HeaderChecksum(IP_IPv4Header)
	RecalculateAndReplaceTransportChecksum(IP_IPv4Header, IP_TPHeader)

	return true
}

func ProcessDNSQuery(TPHeader []byte) (shouldSend bool) {
	// x := dns.Msg{}
	// x.Unpack(TPHeader)
	// // x.String()
	// log.Println("DNS:", x.String())

	return
}

func RecalculateAndReplaceIPv4HeaderChecksum(bytes []byte) {
	// Clear checksum bytes
	bytes[10] = 0
	bytes[11] = 0

	// Compute checksum
	var csum uint32
	for i := 0; i < len(bytes); i += 2 {
		csum += uint32(bytes[i]) << 8
		csum += uint32(bytes[i+1])
	}
	for {
		// Break when sum is less or equals to 0xFFFF
		if csum <= 65535 {
			break
		}
		// Add carry to the sum
		csum = (csum >> 16) + uint32(uint16(csum))
	}

	// Flip all the bits and replace checksum
	binary.BigEndian.PutUint16(bytes[10:12], ^uint16(csum))
	return
}

func RecalculateAndReplaceTransportChecksum(IPv4Header []byte, TPPacket []byte) {

	if IPv4Header[9] == 6 {
		TPPacket[16] = 0
		TPPacket[17] = 0
	} else if IPv4Header[9] == 17 {
		TPPacket[6] = 0
		TPPacket[7] = 0
	}

	var csum uint32
	csum += (uint32(IPv4Header[12]) + uint32(IPv4Header[14])) << 8
	csum += uint32(IPv4Header[13]) + uint32(IPv4Header[15])
	csum += (uint32(IPv4Header[16]) + uint32(IPv4Header[18])) << 8
	csum += uint32(IPv4Header[17]) + uint32(IPv4Header[19])
	csum += uint32(uint8(IPv4Header[9]))
	tcpLength := uint32(len(TPPacket))

	csum += tcpLength & 0xffff
	csum += tcpLength >> 16

	length := len(TPPacket) - 1
	for i := 0; i < length; i += 2 {
		// For our test packet, doing this manually is about 25% faster
		// (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
		csum += uint32(TPPacket[i]) << 8
		csum += uint32(TPPacket[i+1])
	}
	if len(TPPacket)%2 == 1 {
		csum += uint32(TPPacket[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}

	if IPv4Header[9] == 6 {
		binary.BigEndian.PutUint16(TPPacket[16:18], ^uint16(csum))
	} else if IPv4Header[9] == 17 {
		binary.BigEndian.PutUint16(TPPacket[6:8], ^uint16(csum))
	}

	return
}
