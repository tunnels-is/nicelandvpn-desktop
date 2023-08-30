package core

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math/big"
	"net"
	"runtime/debug"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
)

func RecoverAndLogToFile() {
	if r := recover(); r != nil {
		if LogFile != nil {
			CreateErrorLog("", r)
			CreateErrorLog("", string(debug.Stack()))
		} else {
			log.Println(r, string(debug.Stack()))
		}
	}
}

func CopySlice(in []byte) (out []byte) {
	out = make([]byte, len(in))
	_ = copy(out, in)
	return
}

func ReadMIDAndDataFromBuffer(CONN net.Conn, TunnelBuffer []byte) (n int, DL int, err error) {

	n, err = io.ReadAtLeast(CONN, TunnelBuffer[:MIDBufferLength], MIDBufferLength)
	if err != nil {
		CreateErrorLog("", "TUNNEL READER ERROR: ", err)
		return
	}

	if n < MIDBufferLength {
		CreateErrorLog("", "TUNNEL SMALL READ ERROR: ", CONN.RemoteAddr())
		err = errors.New("")
		return
	}

	DL = int(binary.BigEndian.Uint16(TunnelBuffer[6:8]))

	if DL > 0 {
		n, err = io.ReadAtLeast(CONN, TunnelBuffer[MIDBufferLength:MIDBufferLength+DL], DL)
		if err != nil {
			CreateErrorLog("", "TUNNEL DATA READ ERROR: ", err)
			return
		}
	}

	return
}

func GenerateEllipticCurveAndPrivateKey() (PK *ecdsa.PrivateKey, R *OTK_REQUEST, err error) {
	defer RecoverAndLogToFile()

	E := elliptic.P521()
	PK, err = ecdsa.GenerateKey(E, crand.Reader)
	if err != nil {
		CreateErrorLog("", "Unable to generate private key: ", err)
		return nil, nil, err
	}

	R = new(OTK_REQUEST)
	R.X = PK.PublicKey.X
	R.Y = PK.PublicKey.Y
	return
}

func GenerateAEADFromPrivateKey(PK *ecdsa.PrivateKey, R *OTK_REQUEST) (AEAD cipher.AEAD, err error) {
	var CCKeyb *big.Int
	var CCKey [32]byte
	defer func() {
		CCKeyb = nil
		CCKey = [32]byte{}
	}()
	defer RecoverAndLogToFile()

	CCKeyb, _ = PK.Curve.ScalarMult(R.X, R.Y, PK.D.Bytes())
	CCKey = sha256.Sum256(CCKeyb.Bytes())
	AEAD, err = chacha20poly1305.NewX(CCKey[:])
	if err != nil {
		CreateErrorLog("", "Unable to generate AEAD: ", err)
	}
	return
}

func SetGlobalStateAsDisconnected() {
	CreateLog("", "App state set to -Disconnected-")
	GLOBAL_STATE.Connected = false
	GLOBAL_STATE.Connecting = false
}

func DNSMapping(domain string) net.IP {
	parts := strings.Split(domain, ".")
	// parts = parts[:len(parts)-1]
	var d string
	var s string
	if len(parts) == 2 {
		d = strings.Join(parts[len(parts)-2:], ".")
	} else if len(parts) > 2 {
		d = strings.Join(parts[len(parts)-2:], ".")
		s = strings.Join(parts[:len(parts)-2], ".")
	} else {
		return nil
	}

	if AS.AP == nil {
		return nil
	}
	// CreateLog("DNS", "PARTS: ", parts)
	// CreateLog("DNS", "DOMAIN: ", d)
	// CreateLog("DNS", "SUBDOMAIN: ", s)
	// CreateLog("DNS", "AVAILABLE DOMAINS: ", AS.AP.DNS)

	var m *DeviceDNSRegistration
	var ok bool
	if s != "" {
		m, ok = AS.AP.DNS[s]
		if ok {
			return net.ParseIP(m.IP)
		}
	}

	m, ok = AS.AP.DNS[d]
	if ok {
		if m.Wildcard {
			return net.ParseIP(m.IP)
		} else if s == "" {
			return net.ParseIP(m.IP)
		}
	}

	return nil
}

// func CraftDNSResponse(domain string, ip net.IP) {
// 	ip4 := ip.To4()

// 	msg := dnsmessage.Message{
// 		Header: dnsmessage.Header{Response: true, Authoritative: true},
// 		Questions: []dnsmessage.Question{
// 			{
// 				Name:  dnsmessage.MustNewName(domain),
// 				Type:  dnsmessage.TypeA,
// 				Class: dnsmessage.ClassINET,
// 			},
// 		},
// 		Answers: []dnsmessage.Resource{
// 			{
// 				Header: dnsmessage.ResourceHeader{
// 					Name:  dnsmessage.MustNewName(domain),
// 					Type:  dnsmessage.TypeA,
// 					Class: dnsmessage.ClassINET,
// 				},
// 				Body: &dnsmessage.AResource{A: [4]byte{ip4[0], ip4[1], ip4[2], ip4[4]}},
// 			},
// 		},
// 	}

// 	buf, err := msg.Pack()
// 	if err != nil {
// 		panic(err)
// 	}
// }
