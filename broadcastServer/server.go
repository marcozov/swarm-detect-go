package main

import (
	"flag"
	"fmt"
	"net"
	"github.com/dedis/protobuf"
	"github.com/marcozov/swarm-detect-go/structures"
)

func main() {
	address := flag.String("address", "192.168.0.100:3000", "ip:port for the node")
	flag.Parse()

	fmt.Println(address)
	udpAddress, err := net.ResolveUDPAddr("udp4", *address)
	//udpAddress, err := net.ResolveUDPAddr("udp4", ":5000")
	if err != nil {
		panic (fmt.Sprintf("Address not valid: %s", err))
	}

	udpConnection, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		panic (fmt.Sprintf("Error in opening UDP listener: %s", err))
	}


	udpBuffer := make([]byte, 64)

	fmt.Println("waiting for udp data..")
	n, senderAddress, err := udpConnection.ReadFromUDP(udpBuffer)

	if err != nil {
		panic(fmt.Sprintf("error in reading UDP data: %s.\nudpBuffer: %v\nsenderAddress: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), n))
	}

	udpBuffer = udpBuffer[:n]

	fmt.Println("udp buffer: ", udpBuffer)
	packet := &structures.Packet{}

	fmt.Println("decoding udp data..")
	err = protobuf.Decode(udpBuffer, packet)

	if err != nil {
		panic(fmt.Sprintf("error in decoding UDP data: %s\nudpBuffer: %v\nsenderAddress: %s\nreceived packet: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), packet, n))
	}

	fmt.Println("packet: ", packet.Probe)
}
