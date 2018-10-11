package main

import (
	//"github.com/dedis/protobuf"
	"fmt"
	"net"
	"time"
	"math/rand"
	"github.com/dedis/protobuf"
)

func (node *Node) handleIncomingMessages() {
	fmt.Println("enter message handler..")
	for {
		// may need to be expanded tu support bigger messages..
		udpBuffer := make([]byte, 16)
		packet := &Packet{}
		_, senderAddress, err := node.Connection.ReadFromUDP(udpBuffer)
		//fmt.Println("read bytes: ", n, "from ", senderAddress, "encoded (receiving) packet: ", udpBuffer)

		if err != nil {
			panic(fmt.Sprintf("error in reading UDP data: %s\n", err))
		}

		err = protobuf.Decode(udpBuffer, packet)

		if err != nil {
			panic(fmt.Sprintf("error in decoding UDP data: %s\n", err))
		}
		fmt.Println("new message received: ", packet)

		if packet.Start != nil {
			node.HandleStartMessage(packet, *senderAddress)
		} else if packet.Status != nil {
			node.HandleStatusMessage(packet, *senderAddress)
		}
	}
}

func (node *Node) HandleStartMessage(packet *Packet, senderAddress net.UDPAddr) {
	//fmt.Println("Handling incoming *START* message..")
	//fmt.Println(packet.Start, packet.End, packet.Status)
	if node.CurrentStatus.CurrentState == Finish && node.CurrentStatus.CurrentRound < packet.Start.RoundID {
		node.RemainingPeers = node.InitPeersMap()
		node.CurrentStatus.CurrentRound = packet.Start.RoundID
		node.CurrentStatus.CurrentState = Start

		toPropagate := &Packet {
			Start: &StartMessage{
				RoundID: node.CurrentStatus.CurrentRound,
			},
		}

		node.StartHandler <- toPropagate

		go func() {
			//fmt.Println("finishing (from startHandler)...")
			//fmt.Printf("DEBUG: %s %s\n", node.CurrentStatus, node.RemainingPeers)
			n := time.Duration(rand.Intn(20))
			//fmt.Printf("random wait: %d\n", n)
			time.Sleep(n*time.Second)
			node.CurrentStatus.CurrentState = Finish
			//fmt.Println("finished (from startHandler)!")
			fmt.Printf("DEBUG: %s %s\n", node.CurrentStatus, node.RemainingPeers)
		} ()
	}
}

func (node *Node) HandleStatusMessage(packet *Packet, senderAddress net.UDPAddr) {
	//fmt.Println("Handling incoming *STATUS* message..")

	if node.CurrentStatus.CurrentRound == packet.Status.CurrentRound && packet.Status.CurrentState == Finish {
		if node.GetPeer(senderAddress) != nil {
			node.RemoveRemainingPeer(senderAddress)
		}
	}
}

func (node *Node) triggerPropagators() {
	ticker := time.NewTicker(3*time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			// if the computation is finished and all the peers finished too: send START
			if node.CurrentStatus.CurrentState == Finish && node.RemainingPeers.Length() == 0 {
				// go to the following round
				node.RemainingPeers = node.InitPeersMap()
				node.CurrentStatus.CurrentRound++
				node.CurrentStatus.CurrentState = Start

				toPropagate := &Packet {
					Start: &StartMessage{
						RoundID: node.CurrentStatus.CurrentRound,
					},
				}

				node.StartHandler <- toPropagate

				go func() {
					//fmt.Println("finishing (from triggerPropagators)...")
					//fmt.Printf("DEBUG: %s %s\n", node.CurrentStatus, node.RemainingPeers)
					//time.Sleep(10*time.Second)
					n := time.Duration(rand.Intn(20))
					//fmt.Printf("random wait: %d\n", n)
					time.Sleep(n*time.Second)
					node.CurrentStatus.CurrentState = Finish
					//fmt.Println("finished (from triggerPropagators)!")
					fmt.Printf("DEBUG: %s %s\n", node.CurrentStatus, node.RemainingPeers)
				} ()
			}

			// always send a status packet
			statusPacket := &Packet {
				Status: node.CurrentStatus,
			}

			node.StatusHandler <- statusPacket
		}
	}
}

func (node *Node) broadcast(packet *Packet) {
	node.sendToPeers(packet, node.Peers)
}

func (node *Node) sendToPeers(packet *Packet, peers map[string]Peer) {
	for _, peer := range peers {
		node.sendToPeer(packet, peer)
	}
}

func (node *Node) sendToPeer(packet *Packet, peer Peer) {
	packetBytes, err := protobuf.Encode(packet)
	if err != nil {
		panic(fmt.Sprintf("Error in encoding the message: %s", err))
	}

	//fmt.Println("encoded sending packet: ", packetBytes)
	_, err = node.Connection.WriteToUDP(packetBytes, peer.peerAddress)
	//fmt.Println("sent bytes: ", n)

	if err != nil {
		panic(fmt.Sprintf("Error in sending udp data: %s", err))
	}
}

func (node *Node) propagateStartMessage(channel chan *Packet) {
	for {
		start := <-channel
		fmt.Printf("START message to send to everyone: %s\n", start)

		node.broadcast(start)
	}
}

func (node *Node) propagateStatusMessage(channel chan *Packet) {
	for {
		status := <-channel
		fmt.Printf("STATUS message to send to everyone: %s. Remaining peers: %s\n", status, node.RemainingPeers)

		node.sendToPeers(status, node.RemainingPeers.v)
	}
}