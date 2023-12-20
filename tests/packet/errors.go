// Copyright 2015 Tim Heckman. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

// Package packetserr is the package which contains the errors for the packets package.
// This package on its own doesn't do much and is depended on by the packets package.
package packets

import (
	"errors"
	"fmt"
)

// TCPDataOffsetInvalid is a type that implements the error interface. It's used for errors
// marshaling the TCPHeader data.
var TCPDataOffsetInvalid = errors.New("DataOffset field must be at least 5 and no more than 15")

// ChecksumInvalidKind is a type that implements the error interface. It's used when
// an invalid packet kind is provided to the ChecksumIPv4 function.
var ChecksumInvalidKind = errors.New("Checksum kind should either be 'tcp' OR 'udp'.")

// TCPDataOffsetTooSmall is a type that implements the error interface. It's used for errors
// marshaling the TCPHeader data. Specifically, this is used when the DataOffset is too small
// for the amount of data in the TCP header.
type TCPDataOffsetTooSmall struct {
	ExpectedSize uint8
}

func (e TCPDataOffsetTooSmall) Error() string {
	return fmt.Sprintf(
		"The DataOffset field is too small for the data provided. It should be at least %d",
		e.ExpectedSize,
	)
}

// TCPOptionsOverflow is a type that implements the error interface. It's used for errors
// marshaling the TCPHeader data. Specifically, this is used when the TCP Options field exceeds
// its maximum length as specified by the RFC.
type TCPOptionsOverflow struct {
	MaxSize int
}

func (e TCPOptionsOverflow) Error() string {
	return fmt.Sprintf("TCP Options are too large, must be less than %d total bytes", e.MaxSize)
}

// TCPOptionDataInvalid is a type that implements the error interface. It's used for errors
// marshaling the TCPHeader data. Specifically, this is used when the TCP Options Length field
// doesn't match the data provided.
type TCPOptionDataInvalid struct {
	Index int
}

func (e TCPOptionDataInvalid) Error() string {
	return fmt.Sprintf("Option %d Length doesn't match length of data", e.Index)
}

// TCPOptionDataTooLong is a type that implements the error interface. It's used for errors
// marshaling the TCPHeader data. Specifically, this is use for when the TCP Options Data field is
// too long for the Options field as per the RFC.
type TCPOptionDataTooLong struct {
	Index int
}

func (e TCPOptionDataTooLong) Error() string {
	return fmt.Sprintf("Option %d Data cannot be larger than 253 bytes", e.Index)
}

// UDPPayloadTooLarge is a type that implements the error interface. It's used for errors
// marshaling the UDPHeader data. Specifically, this is use for when the UDP payload is too large.
type UDPPayloadTooLarge struct {
	MaxSize, Len int
}

func (e UDPPayloadTooLarge) Error() string {
	return fmt.Sprintf(
		"UDP Payload must not be larger than %d byte, was %d bytes",
		e.MaxSize, e.Len,
	)
}
