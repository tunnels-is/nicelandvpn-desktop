package packets

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
)

func prettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	log.Println(string(s))
}

func CopySlice(in []byte) (out []byte) {
	out = make([]byte, len(in))
	_ = copy(out, in)
	return
}

func CreateSudoIPHeader(SrcIP, DstIP []byte, protocol uint8, length uint32) (csum uint32) {
	csum += (uint32(SrcIP[0]) + uint32(SrcIP[2])) << 8
	csum += uint32(SrcIP[1]) + uint32(SrcIP[3])
	csum += (uint32(DstIP[0]) + uint32(DstIP[2])) << 8
	csum += uint32(DstIP[1]) + uint32(DstIP[3])
	csum += uint32(protocol)
	csum += length & 0xffff
	csum += length >> 16
	return
}

func tcpipChecksum(data []byte, csum uint32) uint16 {
	// to handle odd lengths, we loop to length - 1, incrementing by 2, then
	// handle the last byte specifically by checking against the original
	// length.
	length := len(data) - 1
	for i := 0; i < length; i += 2 {
		// For our test packet, doing this manually is about 25% faster
		// (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	return ^uint16(csum)
}

func RecalculateTCPChecksum(IPv4Header []byte, TCPPacket []byte) {

	TCPPacket[16] = 0
	TCPPacket[17] = 0
	var csum uint32
	csum += (uint32(IPv4Header[12]) + uint32(IPv4Header[14])) << 8
	csum += uint32(IPv4Header[13]) + uint32(IPv4Header[15])
	csum += (uint32(IPv4Header[16]) + uint32(IPv4Header[18])) << 8
	csum += uint32(IPv4Header[17]) + uint32(IPv4Header[19])
	csum += uint32(uint8(IPv4Header[9]))
	tcpLength := uint32(len(TCPPacket))

	csum += tcpLength & 0xffff
	csum += tcpLength >> 16

	length := len(TCPPacket) - 1
	for i := 0; i < length; i += 2 {
		// For our test packet, doing this manually is about 25% faster
		// (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
		csum += uint32(TCPPacket[i]) << 8
		csum += uint32(TCPPacket[i+1])
	}
	if len(TCPPacket)%2 == 1 {
		csum += uint32(TCPPacket[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	log.Println("NEW CHECKSUM:", ^uint16(csum))
	binary.BigEndian.PutUint16(TCPPacket[16:18], ^uint16(csum))

	return
}

func ParsePacket(packet []byte, NP uint16, SRCIP []byte, NIP [4]byte) {

	// Src IP for new check is always the same, can we save it somewhere ?

	packet2 := CopySlice(packet)

	// packet2[16] = NIP[0]
	// packet2[17] = NIP[1]
	// packet2[18] = NIP[2]
	// packet2[19] = NIP[3]

	packet2[12] = SRCIP[0]
	packet2[13] = SRCIP[1]
	packet2[14] = SRCIP[2]
	packet2[15] = SRCIP[3]

	binary.BigEndian.PutUint16(packet2[20:22], NP)
	ipL := (packet2[0] << 4 >> 4) * 32 / 8
	RecalculateIPv4HeaderChecksum(packet2[:ipL])
	RecalculateTCPChecksum(packet2[0:20], packet2[20:])

	log.Println("IP CHECKSUM:", binary.BigEndian.Uint16(packet2[10:12]))
	log.Println("TCP CHECKSUM:", binary.BigEndian.Uint16(packet2[20+16:20+18]))

	// IPHeader, _ := extractIPHeaderFromPacket(packet2)

	// TCPHeader, err := unmarshalTCPHeader(packet2[ipL:])
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	// prettyPrint(IPHeader)
	// prettyPrint(TCPHeader)

	// TCPHeader.SourcePort = 1000

}

func RecalculateIPv4HeaderChecksum(bytes []byte) {
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

type IPHeader struct {
	Bytes []byte

	Version     byte
	Length      byte
	TotalLength uint16
	Protocol    byte
	Checksum    uint16
	ID          uint16

	ToS byte

	FlagReserved   uint16
	FlagMF         uint16
	FlagDF         uint16
	FragmentOffset uint16

	TTL byte

	DestinationIPBytes []byte
	SourceIPBytes      []byte

	Options []byte
}

// https://en.wikipedia.org/wiki/Internet_Protocol_version_4
func extractIPHeaderFromPacket(data []byte) (IPH *IPHeader, err error) {

	IPH = new(IPHeader)
	IPH.Version = data[0] >> 4
	IPH.Length = (data[0] << 4 >> 4) * 32 / 8
	// IPH.Length = uint8(data[0]) & 0x0F
	IPH.TotalLength = binary.BigEndian.Uint16(data[2:4])
	IPH.ToS = data[1]

	IPH.ID = binary.BigEndian.Uint16(data[4:6])

	IPH.FragmentOffset = binary.BigEndian.Uint16(data[6:8])
	IPH.FlagReserved = (IPH.FragmentOffset >> 15) & 0x01
	IPH.FlagDF = (IPH.FragmentOffset >> 14) & 0x01
	IPH.FlagMF = (IPH.FragmentOffset >> 13) & 0x01
	IPH.FragmentOffset = IPH.FragmentOffset & 0x1FFF

	IPH.TTL = data[8]
	IPH.Protocol = data[9]

	IPH.SourceIPBytes = data[12:16]
	IPH.DestinationIPBytes = data[16:20]
	IPH.Options = data[20 : 20+IPH.Length]
	IPH.Checksum = binary.BigEndian.Uint16(data[10:12])

	IPH.Bytes = data[:IPH.Length]

	// This code is added for the following enviroment:
	// * Windows 10 with TSO option activated. ( tested on Hyper-V, RealTek ethernet driver )
	if IPH.Length == 0 {
		// If using TSO(TCP Segmentation Offload), length is zero.
		// The actual packet length is the length of data.
		IPH.Length = byte(len(data))
	}
	return
}

func checksum(data []byte) uint16 {
	dataSize := len(data) - 1

	var sum uint32

	for i := 0; i+1 < dataSize; i += 2 {
		sum += uint32(data[i+1])<<8 | uint32(data[i])
	}

	if dataSize&1 == 1 {
		sum += uint32(data[dataSize])
	}

	sum = sum>>16 + sum&0xffff
	sum = sum + sum>>16

	return ^uint16(sum)
}

// ChecksumIPv4 is a function for computing the TCP checksum of an IPv4 packet. The kind
// field is either 'tcp' or 'udp' and returns an error if invalid input is given.
//
// The returned error type may be packetserr.ChecksumInvalidKind if an invalid
// kind field is provided.
func ChecksumIPv4(data []byte, kind, laddr, raddr string) (uint16, error) {
	// convert the IP address strings to their byte equivalents
	srcBytes, dstBytes := ipv4AddrToBytes(laddr), ipv4AddrToBytes(raddr)

	var protocol uint8

	switch kind {
	case "tcp", "TCP":
		protocol = 6
	case "udp", "UDP":
		protocol = 17
	default:
		return 0, ChecksumInvalidKind
	}

	// create a pseudo header for the packet checksumming
	pHeader := new(bytes.Buffer)

	binary.Write(pHeader, binary.BigEndian, srcBytes[0])
	binary.Write(pHeader, binary.BigEndian, srcBytes[1])
	binary.Write(pHeader, binary.BigEndian, srcBytes[2])
	binary.Write(pHeader, binary.BigEndian, srcBytes[3])
	binary.Write(pHeader, binary.BigEndian, dstBytes[0])
	binary.Write(pHeader, binary.BigEndian, dstBytes[1])
	binary.Write(pHeader, binary.BigEndian, dstBytes[2])
	binary.Write(pHeader, binary.BigEndian, dstBytes[3])
	binary.Write(pHeader, binary.BigEndian, uint8(0))
	binary.Write(pHeader, binary.BigEndian, protocol)
	binary.Write(pHeader, binary.BigEndian, uint16(len(data)))
	pHeader.Write(data)

	return checksum(pHeader.Bytes()), nil
}
