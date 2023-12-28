package core

import (
	"encoding/binary"
	"net"

	"github.com/miekg/dns"
)

func (V *VPNConnection) ProcessEgressPacket(p *[]byte) (sendRemote bool, sendLocal bool) {
	packet := *p

	V.EP_Version = packet[0] >> 4
	if V.EP_Version != 4 {
		return false, false
	}

	V.EP_Protocol = packet[9]
	if V.EP_Protocol != 6 && V.EP_Protocol != 17 {
		return false, false
	}

	// Get the full IPv4Header length in bytes
	V.EP_IPv4HeaderLength = (packet[0] << 4 >> 4) * 32 / 8

	V.EP_IPv4Header = packet[:V.EP_IPv4HeaderLength]
	V.EP_TPHeader = packet[V.EP_IPv4HeaderLength:]

	// DROP RST packets
	if V.EP_Protocol == 6 {
		V.EP_RST = V.EP_TPHeader[13] & 0x7 >> 2
		if V.EP_RST == 1 {
			V.EP_NEW_RST = int(V.EP_TPHeader[13])
			V.EP_NEW_RST |= int(0b00010100)
			V.EP_TPHeader[13] = byte(V.EP_NEW_RST)
			// fmt.Printf("%08b - RST:%08b\n", EP_TPHeader[13], EP_RST)
			// fmt.Printf("POST TRANSFORM: %08b\n", EP_TPHeader[13])
			// log.Println("RST PACKET")
			// return false, false
		}
	}

	V.EP_DstIP[0] = packet[16]
	V.EP_DstIP[1] = packet[17]
	V.EP_DstIP[2] = packet[18]
	V.EP_DstIP[3] = packet[19]

	// This drops NETBIOS DNS packets to the VPN interface
	if V.EP_DstIP == [4]byte{10, 4, 3, 255} {
		return false, false
	}

	V.EP_SrcPort[0] = V.EP_TPHeader[0]
	V.EP_SrcPort[1] = V.EP_TPHeader[1]

	V.EP_DstPort[0] = V.EP_TPHeader[2]
	V.EP_DstPort[1] = V.EP_TPHeader[3]

	// CUSTOM DNS
	// https://stackoverflow.com/questions/7565300/identifying-dns-packets
	if V.EP_Protocol == 17 {
		if IsDNSQuery(V.EP_TPHeader[8:]) {
			// log.Println("DNS FOUND!!!!!!")
			// log.Println("UDP HEADER:", EP_TPHeader[:8])
			// log.Println("UDP DATA:", EP_TPHeader[8:])
			// log.Println("UDP HEADER:", EP_TPHeader[:8], EP_DstIP, EP_DstPort)
			V.EP_DNS_Response, V.EP_DNS_OK = V.ProcessEgressDNSQuery(V.EP_TPHeader[8:])
			if V.EP_DNS_OK {
				// Replace Source IP
				V.EP_IPv4Header[12] = V.EP_IPv4Header[16]
				V.EP_IPv4Header[13] = V.EP_IPv4Header[17]
				V.EP_IPv4Header[14] = V.EP_IPv4Header[18]
				V.EP_IPv4Header[15] = V.EP_IPv4Header[19]

				// Replace Destination IP
				V.EP_IPv4Header[16] = V.AddressNetIP[0]
				V.EP_IPv4Header[17] = V.AddressNetIP[1]
				V.EP_IPv4Header[18] = V.AddressNetIP[2]
				V.EP_IPv4Header[19] = V.AddressNetIP[3]

				// Replace Source Port
				V.EP_DNS_Port_Placeholder[0] = V.EP_TPHeader[0]
				V.EP_DNS_Port_Placeholder[1] = V.EP_TPHeader[1]

				V.EP_TPHeader[0] = V.EP_TPHeader[2]
				V.EP_TPHeader[1] = V.EP_TPHeader[3]

				V.EP_TPHeader[2] = V.EP_DNS_Port_Placeholder[0]
				V.EP_TPHeader[3] = V.EP_DNS_Port_Placeholder[1]

				///
				V.EP_DNS_Packet = append(packet[:V.EP_IPv4HeaderLength+8], V.EP_DNS_Response...)
				// Modify the total Length of the IP Header
				binary.BigEndian.PutUint16(V.EP_DNS_Packet[2:4], uint16(int(V.EP_IPv4HeaderLength)+8+len(V.EP_DNS_Response)))

				// Modify the length of the Transport Header
				binary.BigEndian.PutUint16(V.EP_DNS_Packet[V.EP_IPv4HeaderLength+4:V.EP_IPv4HeaderLength+6], uint16(len(V.EP_DNS_Response))+8)

				RecalculateAndReplaceIPv4HeaderChecksum(V.EP_DNS_Packet[:V.EP_IPv4HeaderLength])
				RecalculateAndReplaceTransportChecksum(V.EP_DNS_Packet[:V.EP_IPv4HeaderLength], V.EP_DNS_Packet[V.EP_IPv4HeaderLength:])

				*p = V.EP_DNS_Packet

				return false, true
			} else {
				if V.IS_UNIX {
					V.PREV_DNS_IP[0] = V.EP_IPv4Header[16]
					V.PREV_DNS_IP[1] = V.EP_IPv4Header[17]
					V.PREV_DNS_IP[2] = V.EP_IPv4Header[18]
					V.PREV_DNS_IP[3] = V.EP_IPv4Header[19]

					V.EP_IPv4Header[16] = C.DNS1Bytes[0]
					V.EP_IPv4Header[17] = C.DNS1Bytes[1]
					V.EP_IPv4Header[18] = C.DNS1Bytes[2]
					V.EP_IPv4Header[19] = C.DNS1Bytes[3]
				}
			}

		}
	}

	if V.EP_Protocol == 6 {

		V.EP_MappedPort = V.CreateOrGetPortMapping(&V.TCP_MAP, V.EP_DstIP, V.EP_SrcPort, V.EP_DstPort)
		if V.EP_MappedPort == nil {
			// log.Println("NO TCP PORT MAPPING", EP_DstIP, EP_SrcPort, EP_DstPort)
			return false, false
		}

	} else if V.EP_Protocol == 17 {

		V.EP_MappedPort = V.CreateOrGetPortMapping(&V.UDP_MAP, V.EP_DstIP, V.EP_SrcPort, V.EP_DstPort)
		if V.EP_MappedPort == nil {
			// log.Println("NO UDP PORT MAPPING", EP_DstIP, EP_SrcPort, EP_DstPort)
			return false, false
		}

	}

	V.EP_NAT_IP, V.EP_NAT_OK = V.NAT_CACHE[V.EP_DstIP]
	if V.EP_NAT_OK {
		// log.Println("FOUND NAT", EP_DstIP, EP_NAT_IP)
		V.EP_IPv4Header[16] = V.EP_NAT_IP[0]
		V.EP_IPv4Header[17] = V.EP_NAT_IP[1]
		V.EP_IPv4Header[18] = V.EP_NAT_IP[2]
		V.EP_IPv4Header[19] = V.EP_NAT_IP[3]
	}

	V.EP_TPHeader[0] = V.EP_MappedPort.Mapped[0]
	V.EP_TPHeader[1] = V.EP_MappedPort.Mapped[1]

	V.EP_IPv4Header[12] = V.EP_VPNSrcIP[0]
	V.EP_IPv4Header[13] = V.EP_VPNSrcIP[1]
	V.EP_IPv4Header[14] = V.EP_VPNSrcIP[2]
	V.EP_IPv4Header[15] = V.EP_VPNSrcIP[3]

	RecalculateAndReplaceIPv4HeaderChecksum(V.EP_IPv4Header)
	RecalculateAndReplaceTransportChecksum(V.EP_IPv4Header, V.EP_TPHeader)

	return true, false
}

func (V *VPNConnection) ProcessIngressPacket(packet []byte) bool {
	V.IP_SrcIP[0] = packet[12]
	V.IP_SrcIP[1] = packet[13]
	V.IP_SrcIP[2] = packet[14]
	V.IP_SrcIP[3] = packet[15]

	V.IP_Protocol = packet[9]

	V.IP_IPv4HeaderLength = (packet[0] << 4 >> 4) * 32 / 8
	V.IP_IPv4Header = packet[:V.IP_IPv4HeaderLength]
	V.IP_TPHeader = packet[V.IP_IPv4HeaderLength:]

	V.IP_DstPort[0] = V.IP_TPHeader[2]
	V.IP_DstPort[1] = V.IP_TPHeader[3]

	V.IP_NAT_IP, V.IP_NAT_OK = V.REVERSE_NAT_CACHE[V.IP_SrcIP]
	if V.IP_NAT_OK {
		// log.Println("FOUND INGRESS NAT", IP_SrcIP, IP_NAT_IP)
		V.IP_IPv4Header[12] = V.IP_NAT_IP[0]
		V.IP_IPv4Header[13] = V.IP_NAT_IP[1]
		V.IP_IPv4Header[14] = V.IP_NAT_IP[2]
		V.IP_IPv4Header[15] = V.IP_NAT_IP[3]

		V.IP_SrcIP[0] = V.IP_NAT_IP[0]
		V.IP_SrcIP[1] = V.IP_NAT_IP[1]
		V.IP_SrcIP[2] = V.IP_NAT_IP[2]
		V.IP_SrcIP[3] = V.IP_NAT_IP[3]
	}

	if V.IP_Protocol == 6 {

		V.IP_MappedPort = GetIngressPortMapping(&V.TCP_MAP, V.IP_SrcIP, V.IP_DstPort)
		if V.IP_MappedPort == nil {
			// log.Println("NO PORT MAPPING", IP_SrcIP, binary.BigEndian.Uint16(IP_DstPort[:]))
			return false
		}

	} else if V.IP_Protocol == 17 {

		V.IP_MappedPort = GetIngressPortMapping(&V.UDP_MAP, V.IP_SrcIP, V.IP_DstPort)
		if V.IP_MappedPort == nil {
			// log.Println("NO PORT MAPPING", IP_SrcIP, binary.BigEndian.Uint16(IP_DstPort[:]))
			return false
		}
	}

	V.IP_TPHeader[2] = V.IP_MappedPort.Local[0]
	V.IP_TPHeader[3] = V.IP_MappedPort.Local[1]

	V.IP_IPv4Header[16] = V.AddressNetIP[0]
	V.IP_IPv4Header[17] = V.AddressNetIP[1]
	V.IP_IPv4Header[18] = V.AddressNetIP[2]
	V.IP_IPv4Header[19] = V.AddressNetIP[3]

	if V.EP_Protocol == 17 {
		if V.IP_SrcIP == C.DNS1Bytes && V.IS_UNIX {
			// if IsDNSQuery(EP_TPHeader[8:]) && IS_UNIX {
			V.IP_IPv4Header[12] = V.PREV_DNS_IP[0]
			V.IP_IPv4Header[13] = V.PREV_DNS_IP[1]
			V.IP_IPv4Header[14] = V.PREV_DNS_IP[2]
			V.IP_IPv4Header[15] = V.PREV_DNS_IP[3]
		}
	}

	RecalculateAndReplaceIPv4HeaderChecksum(V.IP_IPv4Header)
	RecalculateAndReplaceTransportChecksum(V.IP_IPv4Header, V.IP_TPHeader)

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

func (V *VPNConnection) ProcessEgressDNSQuery(UDPData []byte) (DNSResponse []byte, shouldProcess bool) {
	q := new(dns.Msg)
	_ = q.Unpack(UDPData)

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

				if GLOBAL_STATE.C.LogBlockedDomains {
					CreateLog("", "Domain blocked:", domain)
				}
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

				IPS, CNAME := V.Node.DNSAMapping(domain)
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

			TXTS := V.Node.DNSTXTMapping(x.Question[i].Name[0 : len(x.Question[i].Name)-1])
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

			CNAME := V.Node.DNSCNameMapping(x.Question[i].Name[0 : len(x.Question[i].Name)-1])
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
