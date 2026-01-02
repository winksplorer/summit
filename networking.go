package main

import (
	"net"

	"github.com/jaypipes/ghw"
)

type N_NIC struct {
	Name    string   `msgpack:"name"`    // name (enp3s0, lo, etc.)
	MAC     string   `msgpack:"mac"`     // MAC address
	Virtual bool     `msgpack:"virtual"` // is this NIC virtual?
	Speed   string   `msgpack:"speed"`   // speed, human readable
	Duplex  string   `msgpack:"duplex"`  // duplex
	IPs     []string `msgpack:"ips"`     // ip addr(s)

}

func N_GetNics() ([]N_NIC, error) {
	var nics []N_NIC

	netinfo, err := ghw.Network()
	if err != nil {
		return nil, err
	}

	for _, nic := range netinfo.NICs {
		// get ip
		usage, err := net.InterfaceByName(nic.Name)
		if err != nil {
			return nil, err
		}

		addrs, err := usage.Addrs()
		if err != nil {
			return nil, err
		}

		var ips []string
		for _, v := range addrs {
			ips = append(ips, v.String())
		}

		// assemble
		nics = append(nics, N_NIC{
			Name:    nic.Name,
			MAC:     nic.MACAddress,
			Virtual: nic.IsVirtual,
			Speed:   nic.Speed + "Mb/s",
			Duplex:  nic.Duplex,
			IPs:     ips,
		})
	}

	return nics, nil
}

func Comm_NetGetnics(data Comm_Message, keyCookie string) (any, error) {
	return N_GetNics()
}
