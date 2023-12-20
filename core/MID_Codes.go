package core

import "encoding/binary"

const (
	CODE_CLIENT_new_ping                      = 1
	CODE_CLIENT_ping                          = 20
	CODE_CLIENT_connect_tunnel_with_handshake = 28
)

var MIDBufferLength = 8

const (
	META_DL_START = 6
	META_DL_END   = 8
)

func CreateMETABuffer(CODE, GROUP, RID, SID, X1, X2 byte, DL uint16) (METAID [8]byte) {

	METAID[0] = CODE
	METAID[1] = GROUP
	METAID[2] = RID
	METAID[3] = SID
	METAID[4] = X1
	METAID[5] = X2
	binary.BigEndian.PutUint16(METAID[6:], DL)

	return
}

func CreateTunnelBuffer() []byte {
	return make([]byte, 70000)
}
