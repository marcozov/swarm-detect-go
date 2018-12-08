package main

import (
	"flag"
	"fmt"
	"net"
	"github.com/marcozov/swarm-detect-go/structures"
)

func main() {

	address := flag.String("address", "127.0.0.1:5000", "ip:port for the node")
	port := flag.Int("outputPort", 5000, "ip:port for the node")
	flag.Parse()

	udpAddress, err := net.ResolveUDPAddr("udp4", *address)
	if err != nil {
		panic (fmt.Sprintf("Address not valid: %s", err))
	}

	localListenerConnection, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		panic(fmt.Sprintf("Error in opening UDP listener: %s", err))
	}

	packet := &structures.Packet{
		Probe: &structures.ProbeMessage{
			RoundID: 10,
		},
	}

	udpAddress.IP = net.ParseIP("255.255.255.255")
	udpAddress.Port = *port

	structures.RealBroadcast(packet, localListenerConnection, udpAddress)
}
