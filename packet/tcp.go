// Copyright 2015 Tim Heckman. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package packets

import (
	"bytes"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

// These are the control (CTRL) bits of the TCP header. We primarily use these
// for calculating how many bits to shift to get/set them within the header.
const (
	nsBit  uint16 = 256 // NS
	cwrBit uint16 = 128 // CWR
	eceBit uint16 = 64  // ECE
	urgBit uint16 = 32  // URG
	ackBit uint16 = 16  // ACK
	pshBit uint16 = 8   // PSH
	rstBit uint16 = 4   // RST
	synBit uint16 = 2   // SYN
	finBit uint16 = 1   // FIN

	tcpHeaderMinSize int = 20
	tcpOptsMaxSize   int = 40
)

// TCPOption is a struct to hold the data for the various options available for the TCP
// header. See this Wikipedia article for more information:
//
// https://en.wikipedia.org/wiki/Transmission_Control_Protocol#TCP_segment_structure
//
// The struct fields should be fairly self-explanatory after reading the above. To note,
// in all cases the Length field should be equal to 2 + len(Data). This is because the
// Length field, per the RFC, must count itself and the Kind field (which are both
// one byte).
//
// As a convenience, if the Length is set to zero it will be automatically
// calculated and set at the time of marshaling.
type TCPOption struct {
	Kind   uint8
	Length uint8
	Data   []byte
}

// TCPOptionSlice is a slice of TCPOptions for use in the TCPHeader
// Options field.
type TCPOptionSlice []*TCPOption

// TCPHeader is a struct representing a TCP header. The options portion
// of the TCP header is not implemented in this struct.
//
// This struct is a simplified representation of a TCP header. This includes
// making the control (CTRL) bits boolean fields, instead of forcing users of
// this package to do their own bitshifting.
type TCPHeader struct {
	SourcePort      uint16
	DestinationPort uint16
	SeqNum          uint32
	AckNum          uint32
	DataOffset      uint8 // should be either 0 or >= 5 or <=15 (default: 5); if 0 will be auto-set
	Reserved        uint8 // this should always be 0
	NS              bool
	CWR             bool
	ECE             bool
	URG             bool
	ACK             bool
	PSH             bool
	RST             bool
	SYN             bool
	FIN             bool
	WindowSize      uint16 // if set to 0 this becomes 65535
	Checksum        uint16 // suggest setting this to 0 thus offloading to the kernel
	UrgentPointer   uint16
	Options         TCPOptionSlice // optional TCP options; see TCPOption comment for more info
}

// UnmarshalTCPHeader is a function that takes a byte slice and parses it in to an
// instance of *TCPHeader. This also assumes the packet is properly formatted.
func UnmarshalTCPHeader(data []byte) (*TCPHeader, error) {
	return unmarshalTCPHeader(data)
}

// Marshal is a function to marshal the *TCPHeader instance to a byte slice
// without explicitly calculating the checksum for the data. Because there is no
// checksumming of the data, the local and remote addresses are not required.
//
// To note, if the checksum is not provided (i.e. 0) the kernel SHOULD automatically
// calculate this for you.
//
// However, if the *TCPHeader instance has the Checksum field set, it will be
// included in the marshaled data.
//
// The error field may be of packetserr.TCPDataOffsetInvalid,
// packetserr.TCPDataOffsetTooSmall, packetserr.TCPOptionDataTooLong, or
// packetserr.TCPOptionDataInvalid types. See their documentation for more information.
func (tcp *TCPHeader) Marshal() ([]byte, error) {
	return tcp.marshalTCPHeader()
}

// MarshalWithChecksum is a function to marshal the TCPHeader to a byte slice.
// This function is almost the same as Marshal() However, this calculates also
// the TCP checksum and adds it to the header / marshaled data.
//
// It's suggested that you use Marshal() instead and offload the
// checksumming to your kernel (which should do it automatically if field is zero).
//
// The error field may be of packetserr.TCPDataOffsetInvalid,
// packetserr.TCPDataOffsetTooSmall, packetserr.TCPOptionDataTooLong, or
// packetserr.TCPOptionDataInvalid types. See their documentation for more information.
func (tcp *TCPHeader) MarshalWithChecksum(laddr, raddr string) ([]byte, error) {
	// marshal the header
	data, err := tcp.marshalTCPHeader()

	if err != nil {
		return nil, err
	}

	// calculate the checksum using the data that was marshaled
	csum, err := ChecksumIPv4(data, "tcp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	tcp.Checksum = csum

	// remarshal again, with a proper Checksum this time
	fullData, err := tcp.marshalTCPHeader()

	if err != nil {
		return nil, err
	}

	return fullData, nil
}

// UnmarshalTCPOptionSlice is a function that takes a byte slice and converts
// it in to a TCPOptionSlice.
func UnmarshalTCPOptionSlice(data []byte) (TCPOptionSlice, error) {
	var d uint8
	opts := make(TCPOptionSlice, 0)
	reader := bytes.NewReader(data)

	for i := 0; i < len(data); i += 0 {
		var opt TCPOption

		// read the Option-Kind field.
		err := binary.Read(reader, binary.BigEndian, &d)
		if err != nil {
			return nil, err
		}

		switch d {
		case 0:
			// push the index over the loop conditional so we're free
			i = len(data)
			break
		case 1:
			// increment the counter by one
			i++
			continue
		default:
			// set the Kind field to the Option-Kind value
			opt.Kind = d

			// read off the Option-Length field and set it
			err = binary.Read(reader, binary.BigEndian, &d)
			if err != nil {
				return nil, err
			}

			opt.Length = d

			// figure out how much data there is to read,
			// allocate a byte slice for it, and read it
			optionData := make([]byte, int(opt.Length-2))

			err = binary.Read(reader, binary.BigEndian, &optionData)
			if err != nil {
				return nil, err
			}

			// set the Data from the Option-Data field
			// and append this option to the TCPOptionSlice
			opt.Data = optionData
			opts = append(opts, &opt)

			// increment the counter with the number of byte reads
			i = int(uint8(i) + opt.Length)
		}
	}

	return opts, nil
}

// Marshal is a method to marshal the TCPOptionSlice to the raw bytes for use
// in the TCPHeader.Marshal() method. It's exported mainly for convenience as
// marsahling uses this function itself.
func (tcpos TCPOptionSlice) Marshal() ([]byte, error) {
	buf := new(bytes.Buffer)

	for index, opt := range tcpos {
		// yeah, make sure this shit isn't nil
		if opt == nil {
			continue
		}

		switch opt.Kind {
		case 0:
			if buf.Len() == 0 {
				return make([]byte, 0), nil
			}
			fallthrough
		case 1:
			binary.Write(buf, binary.BigEndian, opt.Kind)
		default:
			// make sure we're not going to overflow the uint8 Length field
			if len(opt.Data)+2 > 255 {
				return nil, TCPOptionDataTooLong{Index: index}
			}

			// if the option's Length is zero: auto-calculate the value for that field
			// otherwise: validate that the Length is len(opt.Data) + 2
			if opt.Length == 0 {
				opt.Length = uint8(len(opt.Data)) + 2
			} else if uint8(len(opt.Data))+2 != opt.Length {
				return nil, TCPOptionDataInvalid{Index: index}
			}

			binary.Write(buf, binary.BigEndian, opt.Kind)
			binary.Write(buf, binary.BigEndian, opt.Length)
			binary.Write(buf, binary.BigEndian, opt.Data)

			// if there looks to be no more options just continue through
			// to avoid erroneous padding of the data
			if len(tcpos)-1 == index {
				continue
			}

			// if this isn't the last option, pad to the nearest
			// 32-bit boundary using ones (1)
			for i := 0; i < buf.Len()%4; i++ {
				binary.Write(buf, binary.BigEndian, uint8(1))
			}
		}
	}

	return buf.Bytes(), nil
}

// ipv4AddrToBytes converts an IPv4 address to its four individual pieces in bytes
func ipv4AddrToBytes(addr string) []byte {
	o := strings.Split(addr, ".")

	o0, _ := strconv.Atoi(o[0])
	o1, _ := strconv.Atoi(o[1])
	o2, _ := strconv.Atoi(o[2])
	o3, _ := strconv.Atoi(o[3])

	return []byte{byte(o0), byte(o1), byte(o2), byte(o3)}
}

func ctrlBitSet(value bool, bit uint16) uint16 {
	// if the value is false, set it to zero
	if !value {
		return 0
	}

	// flip the bit in a uint16 so that we can bitwise OR it
	// with our existing value
	//
	// figure out how many bits we need to shift to set what we want
	shift := uint(math.Log2(float64(bit)))

	return uint16(1) << shift
}

func ctrlBitValue(ctrl uint16, bit uint16) bool {
	// figure out how many bits we need to shift to get what we want
	shift := uint(math.Log2(float64(bit)))

	// if the bit is one, return true
	if ctrl>>shift&1 == 1 {
		return true
	}

	// otherwise false
	return false
}

func optionsLen(opts []TCPOption) (count int) {
	for _, opt := range opts {
		count += int(opt.Length)
	}
	return
}

func (tcp *TCPHeader) marshalTCPHeader() ([]byte, error) {
	optBytes, err := tcp.Options.Marshal()

	if err != nil {
		return nil, err
	}

	// if the calculated length of the options is too large
	// return an error
	if len(optBytes) > tcpOptsMaxSize {
		return nil, TCPOptionsOverflow{MaxSize: tcpOptsMaxSize}
	}

	// determine how large the tcp.DataOffset field should be by diving the length
	// of the whole TCP header by 4 (4 bytes [32-bits]) and getting the ceiling of
	// that value
	dataOffsetSize := uint8(math.Ceil(float64(tcpHeaderMinSize+len(optBytes)) / 4))

	// if the field is the type's default, and an obviously invalid value
	// then just set it to the bare minimum for the TCP header.
	if tcp.DataOffset == 0 {
		tcp.DataOffset = dataOffsetSize
	}

	// if the offset is outside of the acceptable range
	// fail with a DataOffsetInvalid error
	if tcp.DataOffset > 15 || tcp.DataOffset < 5 {
		return nil, TCPDataOffsetInvalid
	}

	// if the WindowSize field is the default let's set it to something better
	if tcp.WindowSize == 0 {
		tcp.WindowSize = 65535
	}

	// build the DataOffset, Reserved, and Control Flags data
	ctrl := uint16(tcp.DataOffset)<<12 |
		uint16(tcp.Reserved)<<9 |
		ctrlBitSet(tcp.NS, nsBit) |
		ctrlBitSet(tcp.CWR, cwrBit) |
		ctrlBitSet(tcp.ECE, eceBit) |
		ctrlBitSet(tcp.URG, urgBit) |
		ctrlBitSet(tcp.ACK, ackBit) |
		ctrlBitSet(tcp.PSH, pshBit) |
		ctrlBitSet(tcp.RST, rstBit) |
		ctrlBitSet(tcp.SYN, synBit) |
		ctrlBitSet(tcp.FIN, finBit)

	buf := new(bytes.Buffer)

	// write all the data to the byte buffer
	binary.Write(buf, binary.BigEndian, tcp.SourcePort)
	binary.Write(buf, binary.BigEndian, tcp.DestinationPort)
	binary.Write(buf, binary.BigEndian, tcp.SeqNum)
	binary.Write(buf, binary.BigEndian, tcp.AckNum)
	binary.Write(buf, binary.BigEndian, ctrl)
	binary.Write(buf, binary.BigEndian, tcp.WindowSize)
	binary.Write(buf, binary.BigEndian, tcp.Checksum)
	binary.Write(buf, binary.BigEndian, tcp.UrgentPointer)

	buf.Write(optBytes)

	// each offset is 4 bytes long so figure out how many bytes of padding
	// we should have (based on the DataOffset size) to line up with the 32-bit
	// boundary
	totalPad := int(tcp.DataOffset*4) - buf.Len()

	// DataOffset is too small for the amount of data in the header
	if totalPad < 0 {
		return nil, TCPDataOffsetTooSmall{ExpectedSize: dataOffsetSize}
	}

	// pad the end of the packet with null bytes to the 32-bit boundary
	for i := 0; i < totalPad; i++ {
		binary.Write(buf, binary.BigEndian, uint8(0))
	}

	return buf.Bytes(), nil
}

func unmarshalTCPHeader(data []byte) (*TCPHeader, error) {
	var header TCPHeader
	var ctrl uint16

	reader := bytes.NewReader(data)

	// pull all the fields from the data
	err := binary.Read(reader, binary.BigEndian, &header.SourcePort)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.DestinationPort)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.SeqNum)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.AckNum)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &ctrl)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.WindowSize)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.Checksum)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.UrgentPointer)
	if err != nil {
		return nil, err
	}

	header.DataOffset = uint8(ctrl >> 12)
	header.Reserved = uint8(ctrl >> 9 & 7)

	// We need to convert the control flags to their boolean counterparts.
	// Each control flag is one bit in size, so shift that bit to the end
	// and use a bitwise AND of 1 to see if it's enabled.
	header.NS = ctrlBitValue(ctrl, nsBit)
	header.CWR = ctrlBitValue(ctrl, cwrBit)
	header.ECE = ctrlBitValue(ctrl, eceBit)
	header.URG = ctrlBitValue(ctrl, urgBit)
	header.ACK = ctrlBitValue(ctrl, ackBit)
	header.PSH = ctrlBitValue(ctrl, pshBit)
	header.RST = ctrlBitValue(ctrl, rstBit)
	header.SYN = ctrlBitValue(ctrl, synBit)
	header.FIN = ctrlBitValue(ctrl, finBit)

	if header.DataOffset > 5 {
		optsBytes := make([]byte, len(data)-tcpHeaderMinSize)

		_, err = reader.Read(optsBytes)
		if err != nil {
			return nil, err
		}

		opts, err := UnmarshalTCPOptionSlice(optsBytes)
		if err != nil {
			return nil, err
		}

		header.Options = opts
	}

	return &header, nil
}
