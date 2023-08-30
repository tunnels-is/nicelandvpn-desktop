package core

import (
	"net"
)

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func BUILD_NAT_MAP(AP *AccessPoint) (err error) {
	AP.NAT_CACHE = make(map[[4]byte][4]byte)
	AP.REVERSE_NAT_CACHE = make(map[[4]byte][4]byte)

	for _, v := range AP.NAT {
		ip2, ip2net, err := net.ParseCIDR(v.Nat)
		if err != nil {
			return err
		}
		ip, ipnet, err := net.ParseCIDR(v.Network)
		if err != nil {
			return err
		}

		ip = ip.Mask(ipnet.Mask)
		ip2 = ip2.Mask(ip2net.Mask)

		for ipnet.Contains(ip) && ip2net.Contains(ip2) {

			AP.NAT_CACHE[[4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}] = [4]byte{ip[0], ip[1], ip[2], ip[3]}
			AP.REVERSE_NAT_CACHE[[4]byte{ip[0], ip[1], ip[2], ip[3]}] = [4]byte{ip2[0], ip2[1], ip2[2], ip2[3]}

			inc(ip)
			inc(ip2)
		}

	}
	return
}
