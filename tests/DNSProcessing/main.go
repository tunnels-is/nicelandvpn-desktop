package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/tunnels-is/nicelandvpn-desktop/core"
)

var packet = []byte{

	69, 0, 0, 94, 251, 140, 0, 0, 63, 17, 112, 251, 1, 1, 1, 1, 10, 4, 3, 2, // IP Header
	236, 50, 0, 53, 0, 45, 2, 19, // UDP header
	38, 142, 1, 32, 0, 1, 0, 0, 0, 0, 0, 1, // DNS header
	4, 109, 101, 111, 119, 3, 99, 111, 109, 0, 0, 1, 0, 1, 0, 0, 41, 16, 0, 0, 0, 0, 0, 0, 0, // DNS Question
}
var EP_DNS_Port_Placeholder [2]byte
var IP_InterfaceIP = [4]byte{10, 4, 3, 2}

func main() {

	core.AS = new(core.AdapterSettings)
	core.AS.AP = new(core.AccessPoint)
	core.AS.AP.DNS = make(map[string]*core.DeviceDNSRegistration)
	core.AS.AP.DNS["meow.com"] = new(core.DeviceDNSRegistration)
	core.AS.AP.DNS["meow.com"].IP = []string{"1.1.1.1", "3.3.3.3"}

	PrintPacket(packet, "ORIGINAL")
	fmt.Printf("%p\n", packet)
	ProcessEgressPacket(&packet)
	PrintPacket(packet, "NEW")
	fmt.Printf("%p\n", packet)

}

func ProcessEgressPacket(p *[]byte) {

	packet := *p

	EP_IPv4HeaderLength := (packet[0] << 4 >> 4) * 32 / 8

	EP_IPv4Header := packet[:EP_IPv4HeaderLength]
	EP_TPHeader := packet[EP_IPv4HeaderLength:]

	EP_DNS_Response, OK := core.ProcessEgressDNSQuery(EP_TPHeader[8:])
	if OK {

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

		log.Println("DNS RESPONSE:", EP_DNS_Response)
		// newPacket := make([]byte, int(EP_IPv4HeaderLength)+8+len(EP_DNS_Response))
		newPacket := append(packet[:EP_IPv4HeaderLength+8], EP_DNS_Response...)

		binary.BigEndian.PutUint16(newPacket[2:4], uint16(int(EP_IPv4HeaderLength)+8+len(EP_DNS_Response)))
		log.Println("NEW TTL: ", uint16(int(EP_IPv4HeaderLength)+8+len(EP_DNS_Response)))
		// packet = packet[:EP_IPv4HeaderLength+8]

		binary.BigEndian.PutUint16(newPacket[EP_IPv4HeaderLength+4:EP_IPv4HeaderLength+6], uint16(len(EP_DNS_Response))+8)
		log.Println("NEW TPL: ", uint16(len(EP_DNS_Response))+8)

		// fmt.Printf("%p\n", packet)
		// packet = append(packet, EP_DNS_Response...)
		// fmt.Printf("%p\n", packet)
		// *p = packet

		log.Println("------------------------------- BEFORE CHECKSUM")
		log.Println("EP_TPHeader  ", EP_TPHeader)
		log.Println("EP_IPHEADER  ", EP_IPv4Header)
		// log.Println("FINAL PACKET:", packet[EP_IPv4HeaderLength+8:])
		// log.Println("FINAL P:     ", p)

		core.RecalculateAndReplaceIPv4HeaderChecksum(newPacket[:EP_IPv4HeaderLength])
		core.RecalculateAndReplaceTransportChecksum(newPacket[:EP_IPv4HeaderLength], newPacket[EP_IPv4HeaderLength:])

		// packet[0] = 0
		// packet[1] = 0
		// packet[2] = 0
		log.Println("------------------------------- AFTER CHECKSUM")
		// log.Println("EP_TPHeader  ", EP_TPHeader)
		// fmt.Printf("%p\n", EP_TPHeader)
		// log.Println("EP_IPHEADER  ", EP_IPv4Header)
		// fmt.Printf("%p\n", EP_IPv4Header)
		log.Println("IP:", newPacket[:EP_IPv4HeaderLength])
		log.Println("TP:", newPacket[EP_IPv4HeaderLength:])
		// fmt.Printf("%p\n", packet)

		// log.Println("FINAL P:     ", p)
		// fmt.Printf("%p\n", p)

		*p = newPacket
		// log.Println("FINAL P:", *p)

	}

}

func PrintPacket(packet []byte, label string) {

	// testPacket := gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default)
	// log.Println(label, " ==========================================")
	// fmt.Println(testPacket)
	// log.Println(" =================================================")
}

// QDCOUNT := binary.BigEndian.Uint16(UDPData[4:6])
// log.Println("DNS DATA - QD:", QDCOUNT)

// https://www.rfc-editor.org/rfc/rfc1035
// offset := 12
// offset := 12
// for QDCOUNT != 0 {
// 	UDPContent := UDPData[offset:]

// 	labelEndIndex := int(UDPContent[0])
// 	labelStartIndex := 1
// 	domain := make([]byte, 0)
// 	domainEndIndex := 0

// 	for i := range UDPContent {
// 		if i == labelEndIndex {
// 			domain = append(domain, UDPContent[labelStartIndex:labelEndIndex+1]...)
// 			log.Println(string(UDPContent[labelStartIndex:labelEndIndex+1]), UDPContent[labelStartIndex:labelEndIndex+1])

// 			if UDPContent[i+1] != 0 {
// 				domain = append(domain, '.')
// 			}

// 			// log.Println("SI:", labelStartIndex, "EI:", labelEndIndex)
// 			labelStartIndex = i + 2
// 			labelEndIndex = labelStartIndex + int(UDPContent[i+1]) - 1
// 			// log.Println("NEXT - SI:", labelStartIndex, "EI:", labelEndIndex)
// 		}

// 		if UDPContent[i+1] == 0 {
// 			domainEndIndex = i
// 			offset = domainEndIndex + 5
// 			break
// 		}
// 	}

// 	QTYPE := binary.BigEndian.Uint16(UDPContent[domainEndIndex+1 : domainEndIndex+3])
// 	QCLASS := binary.BigEndian.Uint16(UDPContent[domainEndIndex+3 : domainEndIndex+5])
// 	log.Println("PARSED DNS: ", QTYPE, QCLASS, string(domain))

// 	QDCOUNT--
// }
