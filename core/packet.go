package core

import (
	"encoding/binary"
	"net"

	"github.com/miekg/dns"
)

var (
	PREV_DNS_IP [4]byte
	IS_UNIX     bool = false
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

	EP_RST byte

	EP_DNS_Response         []byte
	EP_DNS_OK               bool
	EP_DNS_Port_Placeholder [2]byte
	EP_DNS_Packet           []byte

	// This IP gets over-written on connect
	EP_VPNSrcIP [4]byte
)

func ProcessEgressPacket(p *[]byte) (sendRemote bool, sendLocal bool) {

	packet := *p

	EP_Version = packet[0] >> 4
	if EP_Version != 4 {
		return false, false
	}

	EP_Protocol = packet[9]
	if EP_Protocol != 6 && EP_Protocol != 17 {
		return false, false
	}

	// Get the full IPv4Header length in bytes
	EP_IPv4HeaderLength = (packet[0] << 4 >> 4) * 32 / 8

	EP_IPv4Header = packet[:EP_IPv4HeaderLength]
	EP_TPHeader = packet[EP_IPv4HeaderLength:]

	// DROP RST packets
	if EP_Protocol == 6 {
		EP_RST = EP_TPHeader[13] & 0x7 >> 2
		// fmt.Printf("%08b - RST:%08b\n", EP_TPHeader[13], EP_RST)
		if EP_RST == 1 {
			// log.Println("RST PACKET")
			return false, false
		}
	}

	EP_DstIP[0] = packet[16]
	EP_DstIP[1] = packet[17]
	EP_DstIP[2] = packet[18]
	EP_DstIP[3] = packet[19]

	// This drops NETBIOS DNS packets to the VPN interface
	if EP_DstIP == [4]byte{10, 4, 3, 255} {
		return false, false
	}

	EP_SrcPort[0] = EP_TPHeader[0]
	EP_SrcPort[1] = EP_TPHeader[1]

	EP_DstPort[0] = EP_TPHeader[2]
	EP_DstPort[1] = EP_TPHeader[3]

	// CUSTOM DNS
	// https://stackoverflow.com/questions/7565300/identifying-dns-packets
	if EP_Protocol == 17 {
		if IsDNSQuery(EP_TPHeader[8:]) {
			// log.Println("DNS FOUND!!!!!!")

			// log.Println("UDP HEADER:", EP_TPHeader[:8])
			// log.Println("UDP DATA:", EP_TPHeader[8:])
			// log.Println("UDP HEADER:", EP_TPHeader[:8], EP_DstIP, EP_DstPort)
			EP_DNS_Response, EP_DNS_OK = ProcessEgressDNSQuery(EP_TPHeader[8:])
			if EP_DNS_OK {
				// Replace Source IP
				EP_IPv4Header[12] = EP_IPv4Header[16]
				EP_IPv4Header[13] = EP_IPv4Header[17]
				EP_IPv4Header[14] = EP_IPv4Header[18]
				EP_IPv4Header[15] = EP_IPv4Header[19]

				// Replace Destination IP
				EP_IPv4Header[16] = IP_InterfaceIP[0]
				EP_IPv4Header[17] = IP_InterfaceIP[1]
				EP_IPv4Header[18] = IP_InterfaceIP[2]
				EP_IPv4Header[19] = IP_InterfaceIP[3]

				// Replace Source Port
				EP_DNS_Port_Placeholder[0] = EP_TPHeader[0]
				EP_DNS_Port_Placeholder[1] = EP_TPHeader[1]

				EP_TPHeader[0] = EP_TPHeader[2]
				EP_TPHeader[1] = EP_TPHeader[3]

				EP_TPHeader[2] = EP_DNS_Port_Placeholder[0]
				EP_TPHeader[3] = EP_DNS_Port_Placeholder[1]

				///
				EP_DNS_Packet = append(packet[:EP_IPv4HeaderLength+8], EP_DNS_Response...)
				// Modify the total Length of the IP Header
				binary.BigEndian.PutUint16(EP_DNS_Packet[2:4], uint16(int(EP_IPv4HeaderLength)+8+len(EP_DNS_Response)))

				// Modify the length of the Transport Header
				binary.BigEndian.PutUint16(EP_DNS_Packet[EP_IPv4HeaderLength+4:EP_IPv4HeaderLength+6], uint16(len(EP_DNS_Response))+8)

				RecalculateAndReplaceIPv4HeaderChecksum(EP_DNS_Packet[:EP_IPv4HeaderLength])
				RecalculateAndReplaceTransportChecksum(EP_DNS_Packet[:EP_IPv4HeaderLength], EP_DNS_Packet[EP_IPv4HeaderLength:])

				*p = EP_DNS_Packet

				return false, true
			} else {

				if IS_UNIX {
					PREV_DNS_IP[0] = EP_IPv4Header[16]
					PREV_DNS_IP[1] = EP_IPv4Header[17]
					PREV_DNS_IP[2] = EP_IPv4Header[18]
					PREV_DNS_IP[3] = EP_IPv4Header[19]

					EP_IPv4Header[16] = C.DNS1Bytes[0]
					EP_IPv4Header[17] = C.DNS1Bytes[1]
					EP_IPv4Header[18] = C.DNS1Bytes[2]
					EP_IPv4Header[19] = C.DNS1Bytes[3]
				}

			}

		}
	}

	if EP_Protocol == 6 {

		EP_MappedPort = CreateOrGetPortMapping(&TCP_o0, EP_DstIP, EP_SrcPort, EP_DstPort)
		if EP_MappedPort == nil {
			// log.Println("NO TCP PORT MAPPING", EP_DstIP, EP_SrcPort, EP_DstPort)
			return false, false
		}

	} else if EP_Protocol == 17 {

		EP_MappedPort = CreateOrGetPortMapping(&UDP_o0, EP_DstIP, EP_SrcPort, EP_DstPort)
		if EP_MappedPort == nil {
			// log.Println("NO UDP PORT MAPPING", EP_DstIP, EP_SrcPort, EP_DstPort)
			return false, false
		}

	}

	EP_NAT_IP, EP_NAT_OK = AS.AP.NAT_CACHE[EP_DstIP]
	if EP_NAT_OK {
		// log.Println("FOUND NAT", EP_DstIP, EP_NAT_IP)
		EP_IPv4Header[16] = EP_NAT_IP[0]
		EP_IPv4Header[17] = EP_NAT_IP[1]
		EP_IPv4Header[18] = EP_NAT_IP[2]
		EP_IPv4Header[19] = EP_NAT_IP[3]
	}

	EP_TPHeader[0] = EP_MappedPort.Mapped[0]
	EP_TPHeader[1] = EP_MappedPort.Mapped[1]

	EP_IPv4Header[12] = EP_VPNSrcIP[0]
	EP_IPv4Header[13] = EP_VPNSrcIP[1]
	EP_IPv4Header[14] = EP_VPNSrcIP[2]
	EP_IPv4Header[15] = EP_VPNSrcIP[3]

	RecalculateAndReplaceIPv4HeaderChecksum(EP_IPv4Header)
	RecalculateAndReplaceTransportChecksum(EP_IPv4Header, EP_TPHeader)

	return true, false
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

	IP_Protocol = packet[9]

	IP_IPv4HeaderLength = (packet[0] << 4 >> 4) * 32 / 8
	IP_IPv4Header = packet[:IP_IPv4HeaderLength]
	IP_TPHeader = packet[IP_IPv4HeaderLength:]

	IP_DstPort[0] = IP_TPHeader[2]
	IP_DstPort[1] = IP_TPHeader[3]

	IP_NAT_IP, IP_NAT_OK = AS.AP.REVERSE_NAT_CACHE[IP_SrcIP]
	if IP_NAT_OK {
		// log.Println("FOUND INGRESS NAT", IP_SrcIP, IP_NAT_IP)
		IP_IPv4Header[12] = IP_NAT_IP[0]
		IP_IPv4Header[13] = IP_NAT_IP[1]
		IP_IPv4Header[14] = IP_NAT_IP[2]
		IP_IPv4Header[15] = IP_NAT_IP[3]

		IP_SrcIP[0] = IP_NAT_IP[0]
		IP_SrcIP[1] = IP_NAT_IP[1]
		IP_SrcIP[2] = IP_NAT_IP[2]
		IP_SrcIP[3] = IP_NAT_IP[3]
	}

	if IP_Protocol == 6 {

		IP_MappedPort = GetIngressPortMapping(&TCP_o0, IP_SrcIP, IP_DstPort)
		if IP_MappedPort == nil {
			// log.Println("NO PORT MAPPING", IP_SrcIP, binary.BigEndian.Uint16(IP_DstPort[:]))
			return false
		}

	} else if IP_Protocol == 17 {

		IP_MappedPort = GetIngressPortMapping(&UDP_o0, IP_SrcIP, IP_DstPort)
		if IP_MappedPort == nil {
			// log.Println("NO PORT MAPPING", IP_SrcIP, binary.BigEndian.Uint16(IP_DstPort[:]))
			return false
		}

	}

	IP_TPHeader[2] = IP_MappedPort.Local[0]
	IP_TPHeader[3] = IP_MappedPort.Local[1]

	IP_IPv4Header[16] = IP_InterfaceIP[0]
	IP_IPv4Header[17] = IP_InterfaceIP[1]
	IP_IPv4Header[18] = IP_InterfaceIP[2]
	IP_IPv4Header[19] = IP_InterfaceIP[3]

	if EP_Protocol == 17 {
		if IP_SrcIP == C.DNS1Bytes && IS_UNIX {
			// if IsDNSQuery(EP_TPHeader[8:]) && IS_UNIX {
			IP_IPv4Header[12] = PREV_DNS_IP[0]
			IP_IPv4Header[13] = PREV_DNS_IP[1]
			IP_IPv4Header[14] = PREV_DNS_IP[2]
			IP_IPv4Header[15] = PREV_DNS_IP[3]
		}
	}

	RecalculateAndReplaceIPv4HeaderChecksum(IP_IPv4Header)
	RecalculateAndReplaceTransportChecksum(IP_IPv4Header, IP_TPHeader)

	return true
}

func IsDNSQuery(UDPData []byte) bool {

	if len(UDPData) < 12 {
		// log.Println("NOT ENOUGH UDP DATA")
		return false
	}

	// QR == 0 when making a DNS Query
	QR := UDPData[2] >> 7
	if QR != 0 {
		return false
	}

	// AN Count is always 0 for queries
	if UDPData[6] != 0 || UDPData[7] != 0 {
		// log.Println("AN COUNT OFF", UDPData[6:8])
		return false
	}

	// NS Count is always 0 for queries
	if UDPData[8] != 0 || UDPData[9] != 0 {
		// log.Println("NS COUNT OFF", UDPData[8:10])
		return false
	}

	return true
}
func ProcessEgressDNSQuery(UDPData []byte) (DNSResponse []byte, shouldProcess bool) {

	q := new(dns.Msg)
	q.Unpack(UDPData)

	x := new(dns.Msg)
	x.SetReply(q)
	x.Authoritative = true
	x.Compress = true

	isCustomDNS := false
	for i := range x.Question {

		if x.Question[i].Qtype == dns.TypeA {
			domain := x.Question[i].Name[0 : len(x.Question[i].Name)-1]

			_, ok := GLOBAL_BLOCK_LIST[domain]
			// CreateLog("", "DNS Q: ", domain, len(GLOBAL_BLOCK_LIST), ok)
			if ok {

				CreateLog("", "Domain blocked:", domain)
				isCustomDNS = true
				x.Answer = append(x.Answer, &dns.A{
					Hdr: dns.RR_Header{
						Class:  dns.TypeA,
						Rrtype: dns.ClassINET,
						Name:   x.Question[i].Name,
						Ttl:    5,
					},
					A: net.ParseIP("127.0.0.1"),
				})

			} else {

				IPS, CNAME := DNSAMapping(domain)
				if CNAME != "" {

					isCustomDNS = true
					x.Answer = append(x.Answer, &dns.CNAME{
						Hdr: dns.RR_Header{
							Class:  dns.ClassNONE,
							Rrtype: dns.TypeCNAME,
							Name:   x.Question[i].Name,
							Ttl:    5,
						},
						Target: CNAME + ".",
					})

				} else if IPS != nil {
					isCustomDNS = true

					for ii := range IPS {
						x.Answer = append(x.Answer, &dns.A{
							Hdr: dns.RR_Header{
								Class:  dns.TypeA,
								Rrtype: dns.ClassINET,
								Name:   x.Question[i].Name,
								Ttl:    5,
							},
							A: IPS[ii].To4(),
						})
					}
				}

			}

		} else if x.Question[i].Qtype == dns.TypeTXT {

			TXTS := DNSTXTMapping(x.Question[i].Name[0 : len(x.Question[i].Name)-1])
			if TXTS != nil {
				isCustomDNS = true
				for ii := range TXTS {
					x.Answer = append(x.Answer, &dns.TXT{
						Hdr: dns.RR_Header{
							Class:  dns.ClassNONE,
							Rrtype: dns.TypeTXT,
							Name:   x.Question[i].Name,
							Ttl:    30,
						},
						Txt: []string{TXTS[ii]},
					})
				}
			}

		} else if x.Question[i].Qtype == dns.TypeCNAME {

			CNAME := DNSCNameMapping(x.Question[i].Name[0 : len(x.Question[i].Name)-1])
			if CNAME != "" {
				isCustomDNS = true
				x.Answer = append(x.Answer, &dns.CNAME{
					Hdr: dns.RR_Header{
						Class:  dns.ClassNONE,
						Rrtype: dns.TypeCNAME,
						Name:   x.Question[i].Name,
						Ttl:    30,
					},
					Target: CNAME + ".",
				})
			}

		}

	}

	if isCustomDNS {

		var err error
		DNSResponse, err = x.Pack()
		if err != nil {
			// log.Println("UNABLE TO PICK DNS RESPONSE: ", err)
			return
		}

		shouldProcess = true
		return
	}

	return
}

func ProcessIngressDNSQuery(TPHeader []byte) bool {

	return true
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
