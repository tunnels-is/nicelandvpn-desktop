// Copyright 2015 Tim Heckman. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package packets

import (
	"bytes"
	"encoding/binary"
)

const (
	udpHeaderLen int = 8
	maxUint16    int = int(^uint16(0))
)

// UDPHeader is the struct representing a UDP header.
type UDPHeader struct {
	SourcePort      uint16
	DestinationPort uint16
	Length          uint16 // if set to zero it will be automatically set
	Checksum        uint16
	Payload         []byte
}

// UnmarshalUDPHeader is a function that takes a byte slice anf formats it in to an
// instance of *UDPHeader. This also assumes the packet is properly formatted.
func UnmarshalUDPHeader(data []byte) (*UDPHeader, error) {
	return unmarshalUDPHeader(data)
}

// Marshal is a function to marshal the *UDPHeader instance to a byte slice
// without explicitly calculating the checkum for the data.
//
// To note, if the checksum is not provided (i.e., 0) the kernel SHOULD automatically
// calculate this for you.
//
// If the Checksum field is set, it will be included in the marshaled data.
func (udp *UDPHeader) Marshal() ([]byte, error) {
	return udp.marshalUDPHeader()
}

// MarshalWithChecksum is a function to marshal the *UDPHeader instance to a byte
// slice including the calculation of the Checksum field.
//
// It's suggested that you use Marshal() instead and offload the checksumming
// to your kernel (which should do it automatically if field is zero)
//
// Because of the requirement to create a pseudoheader to do the checksumming the
// local IPv4 address and remote IPv4 address must be provided in string form.
func (udp *UDPHeader) MarshalWithChecksum(laddr, raddr string) ([]byte, error) {
	// marshal the header
	data, err := udp.marshalUDPHeader()

	if err != nil {
		return nil, err
	}

	csum, err := ChecksumIPv4(data, "udp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	udp.Checksum = csum

	// remarshal again, with proper Checksum this time
	data, err = udp.marshalUDPHeader()

	if err != nil {
		return nil, err
	}

	return data, nil
}

func unmarshalUDPHeader(data []byte) (*UDPHeader, error) {
	var header UDPHeader

	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.BigEndian, &header.SourcePort)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.DestinationPort)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.Length)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &header.Checksum)
	if err != nil {
		return nil, err
	}

	bytesToRead := int(header.Length) - udpHeaderLen

	payload := make([]byte, bytesToRead)

	err = binary.Read(reader, binary.BigEndian, &payload)
	if err != nil {
		return nil, err
	}

	header.Payload = payload

	return &header, nil
}

func (udp *UDPHeader) marshalUDPHeader() ([]byte, error) {
	packetSize := len(udp.Payload) + udpHeaderLen

	if packetSize > maxUint16 {
		return nil, UDPPayloadTooLarge{
			MaxSize: maxUint16 - udpHeaderLen,
			Len:     len(udp.Payload),
		}
	}

	if udp.Length == 0 {
		udp.Length = uint16(packetSize)
	}

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, udp.SourcePort)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, udp.DestinationPort)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, udp.Length)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, udp.Checksum)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, udp.Payload)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
