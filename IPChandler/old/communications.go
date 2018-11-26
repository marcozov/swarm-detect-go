package main

import (
	"fmt"
	"time"
	"github.com/dedis/protobuf"
	"encoding/json"
	"net"
	"sync"
)

// handling the prediction received from the python process
// concurrency is handled by design (theoretically.. need to test)
func (node *Node) HandleLocalPrediction(channel chan []byte) {
	for {
		jsonMessage := <- channel

		// TODO: replace json with gRPC
		res := make(map[int][]float64)
		err := json.Unmarshal(jsonMessage, &res)

		if err != nil {
			panic(fmt.Sprintf("error in decoding json: %s\n", err))
		}

		//fmt.Println("printing converted structure: ")
		//fmt.Println(res)
		node.ReceivedLocalPredictions++
		node.LocalDecision.updateOpinionVector(res)

		if node.ReceivedLocalPredictions == 5 {
			node.CurrentStatus.mux.Lock()
			// probably no need to save the state, since it is sent only to one host (the leader)
			if node.CurrentStatus.StatusValue.CurrentState == WaitingForLocalPredictions {
				fmt.Println("I finished accumulating data for this round!")
				node.CurrentStatus.StatusValue.CurrentState = LocalPredictionsTerminated
				node.CurrentStatus.mux.Unlock()
				node.PredictionsAggregatorHandler <- struct{}{}
			} else {
				node.CurrentStatus.mux.Unlock()
			}

			node.ReceivedLocalPredictions = 0
		}
	}
}

func (node *Node) processMessage(packet *Packet, senderAddress *net.UDPAddr, counter *IntWrapper) {
	//n := rand.Intn(100000000)
	//fmt.Println("******** n: ", n, ", new message received: ", packet)

	if packet.Probe != nil {
		node.HandleReceivedProbe(packet, *senderAddress)
	} else if packet.Status != nil {
		node.HandleReceivedStatus(packet, *senderAddress)
	} else if packet.FinalPrediction != nil {
		node.HandleReceivedFinalPrediction(packet, *senderAddress)
	} else if packet.Ack != nil {
		node.HandleReceivedAcknowledgement(packet, *senderAddress)
	}

	//fmt.Println("************** FINISHED HANDLING THE RECEIVED MESSAGE ", n, " *****************")
}

type IntWrapper struct {
	v int
	mux sync.Mutex
}

func (node *Node) handleIncomingMessages() {
	counter := &IntWrapper{ v: 0}
	for {
		// may need to be expanded to support bigger messages..
		udpBuffer := make([]byte, 64)
		packet := &Packet{}
		n, senderAddress, err := node.Connection.ReadFromUDP(udpBuffer)

		if err != nil {
			panic(fmt.Sprintf("error in reading UDP data: %s.\nudpBuffer: %v\nsenderAddress: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), n))
		}

		udpBuffer = udpBuffer[:n]
		err = protobuf.Decode(udpBuffer, packet)

		if err != nil {
			panic(fmt.Sprintf("error in decoding UDP data: %s\nudpBuffer: %v\nsenderAddress: %s\npacket: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), packet, n))
		}

		go node.processMessage(packet, senderAddress, counter)
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
		panic(fmt.Sprintf("Error in encoding the message: %s\npacket: %s\npacketBytes: %s\n", err, packet, packetBytes))
	}

	_, err = node.Connection.WriteToUDP(packetBytes, peer.peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error in sending udp data: %s\npacket: %s\npacketBytes: %s\n", err, packet, packetBytes))
	}
}

func (node *Node) opinionVectorDEBUG() {
	ticker := time.NewTicker(3*time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			fmt.Println("Person score: ", node.LocalDecision.scores[1])
			fmt.Println("Person coefficient: ", node.LocalDecision.boundingBoxCoefficients[1])
			//fmt.Println("Bottle score: ", node.LocalDecision.scores[44])
			//fmt.Println("Bottle coefficient: ", node.LocalDecision.boundingBoxCoefficients[44])
			//fmt.Println("Cell phone score: ", node.LocalDecision.scores[77])
			//fmt.Println("Cell phone coefficient: ", node.LocalDecision.boundingBoxCoefficients[77])
			//fmt.Println("Keyboard score: ", node.LocalDecision.scores[76])
			//fmt.Println("Keyboard coefficient: ", node.LocalDecision.boundingBoxCoefficients[76])
			fmt.Println("remaining neighbors: ", node.RemainingPeers)
			fmt.Println("received external predictions: ", node.ExternalPredictions)
			fmt.Println("current round: ", node.CurrentStatus.StatusValue.CurrentRound, " current status: ", node.CurrentStatus.StatusValue.CurrentState)
			fmt.Println()
		}
	}
}