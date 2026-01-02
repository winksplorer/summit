package main

import (
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

type N_NicStats struct {
	Name    string `msgpack:"name"`     // NIC name
	RxBytes uint64 `msgpack:"rx_bytes"` // bytes recieved
	TxBytes uint64 `msgpack:"tx_bytes"` // bytes transmitted
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

func N_GetAndSendStats(conn *websocket.Conn, id uint32) error {
	// get nics
	files, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return err
	}

	var nics []N_NicStats

	for _, file := range files {
		// get name and path
		name := file.Name()
		if name == "lo" || name == "bonding_masters" {
			continue
		}

		netPath := filepath.Join("/sys/class/net", name)
		dest, err := os.Readlink(netPath)
		if err != nil {
			return err
		}

		if !filepath.IsAbs(dest) {
			dest = filepath.Join(filepath.Dir(netPath), dest)
		}

		// rx bytes
		rxBytesRaw, err := os.ReadFile(filepath.Join(dest, "statistics/rx_bytes"))
		if err != nil {
			return err
		}

		rxBytes, err := strconv.ParseUint(strings.TrimSpace(string(rxBytesRaw)), 10, 64)
		if err != nil {
			return err
		}

		// tx bytes
		txBytesRaw, err := os.ReadFile(filepath.Join(dest, "statistics/tx_bytes"))
		if err != nil {
			return err
		}

		txBytes, err := strconv.ParseUint(strings.TrimSpace(string(txBytesRaw)), 10, 64)
		if err != nil {
			return err
		}

		// assemble
		nics = append(nics, N_NicStats{
			Name:    name,
			RxBytes: rxBytes,
			TxBytes: txBytes,
		})
	}

	// senc
	return Comm_Send(Comm_Message{
		ID:   id,
		T:    "net.stats",
		Data: nics,
	}, conn)
}

func Comm_NetGetnics(data Comm_Message, keyCookie string) (any, error) {
	return N_GetNics()
}

func Comm_NetStats(ctx context.Context, conn *websocket.Conn, id uint32) {
	// 0s
	if err := N_GetAndSendStats(conn, id); err != nil {
		log.Println("Comm_NetStats: Couldn't send stats:", err)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := N_GetAndSendStats(conn, id); err != nil {
				log.Println("Comm_NetStats: Couldn't send stats:", err)
				return
			}
		}
	}
}
