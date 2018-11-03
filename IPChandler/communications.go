package main

import (
	"fmt"
	"time"
	//"math/rand"
	"github.com/dedis/protobuf"
	"encoding/json"
)

// handling the prediction received from the python process
// concurrency is handled by design (theoretically.. need to test)
func (node *Node) HandleLocalPrediction(channel chan []byte) {
	for {
		jsonMessage := <- channel

		// TODO: replace json with gRPC
		//res := make(map[string][]float64)
		res := make(map[int][]float64)
		err := json.Unmarshal(jsonMessage, &res)

		if err != nil {
			panic(fmt.Sprintf("error in decoding json: %s\n", err))
		}

		//fmt.Println("printing converted structure: ")
		//fmt.Println(res)
		node.ReceivedLocalPredictions++
		node.updateOpinionVector(res)

		if node.ReceivedLocalPredictions == 20 {
			node.CurrentStatus.mux.Lock()

			// probably no need to save the state, since it is sent only to one host (the leader)
			if node.CurrentStatus.StatusValue.CurrentState == WaitingForLocalPredictions {
				fmt.Println("I finished accumulating data for this round!")
				node.CurrentStatus.StatusValue.CurrentPrediction.Value[0] = node.LocalDecision.scores[node.DetectionClass]
				node.CurrentStatus.StatusValue.CurrentPrediction.Value[1] = node.LocalDecision.boundingBoxCoefficients[node.DetectionClass]
				node.CurrentStatus.StatusValue.CurrentState = LocalPredictionsTerminated
				node.PredictionsAggregatorHandler <- struct{}{}
			}

			node.ReceivedLocalPredictions = 0
			node.CurrentStatus.mux.Unlock()
		}
	}
}

func (node *Node) handleIncomingMessages() {
	fmt.Println("enter message handler..")
	for {
		// may need to be expanded to support bigger messages..
		udpBuffer := make([]byte, 32)
		packet := &Packet{}
		//fmt.Println("WAT ********************")
		_, senderAddress, err := node.Connection.ReadFromUDP(udpBuffer)
		//fmt.Println("read bytes: ", n, "from ", senderAddress, "encoded (receiving) packet: ", udpBuffer)

		if err != nil {
			panic(fmt.Sprintf("error in reading UDP data: %s\n", err))
		}
		//fmt.Println("debug received MESSAGE: ", udpBuffer)
		err = protobuf.Decode(udpBuffer, packet)

		if err != nil {
			panic(fmt.Sprintf("error in decoding UDP data: %s\n", err))
		}
		//fmt.Println("new message received: ", packet)

		//if packet.WaitingForLocalPredictions != nil {
		//	node.HandleStartMessage(packet, *senderAddress)
		//} else
		if packet.Probe != nil {
			node.HandleReceivedProbe(packet, *senderAddress)
		} else if packet.Status != nil {
			node.HandleReceivedStatus(packet, *senderAddress)
		} else if packet.FinalPrediction != nil {
			node.HandleReceivedFinalPrediction(packet, *senderAddress)
		} else if packet.Ack != nil {
			node.HandleReceivedAcknowledgement(packet, *senderAddress)
		} //else if packet.End != nil {
		//	fmt.Println("*************** END ROUND MESSAGE ***************", packet.End, "sender: ", senderAddress)
		//	node.HandleReceivedEndRound(packet, *senderAddress)
		//}
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
	//fmt.Println("sending packet to peer ", peer.peerAddress.String())
	_, err = node.Connection.WriteToUDP(packetBytes, peer.peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error in sending udp data: %s", err))
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





// alternative: if we want to remove locks for handling the remaining peers
func (node *Node) removeRemainingPeer(channel chan struct{}) {

}

/*
// can be triggered only if this host is the leader
func (node *Node) PropagateEndRoundMessage(channel chan *Packet) {
	for {
		status := <- channel
		fmt.Printf("END ROUND message to send to everyone: %s. Remaining peers: %s\n", status, node.RemainingPeers)

		node.sendToPeers(status, node.RemainingPeers.v)
	}
}
*/

/*
func (node *Node) triggerPropagators() {
	ticker := time.NewTicker(3*time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			// if the computation is finished and all the peers finished too: send START
			if node.CurrentStatus.StatusValue.CurrentState == LocalPredictionsTerminated && node.RemainingPeers.Length() == 0 {
				// go to the following round
				node.RemainingPeers = node.InitPeersMap()
				node.CurrentStatus.StatusValue.CurrentRound++
				node.CurrentStatus.StatusValue.CurrentState = WaitingForLocalPredictions

				toPropagate := &Packet {
					WaitingForLocalPredictions: &StartMessage{
						RoundID: node.CurrentStatus.StatusValue.CurrentRound,
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
					node.CurrentStatus.StatusValue.CurrentState = LocalPredictionsTerminated
					//fmt.Println("finished (from triggerPropagators)!")
					fmt.Printf("DEBUG: %s %s\n", node.CurrentStatus, node.RemainingPeers)
				} ()
			}

			// always send a status packet
			statusPacket := &Packet {
				Status: &node.CurrentStatus.StatusValue,
			}

			node.StatusHandler <- statusPacket
		}
	}
}*/

func computeCoefficient(entropy float32, boundingBoxRatio BoundingBox) float64 {
	return 0
}

func (node *Node) propagateStartMessage(channel chan *Packet) {
	for {
		start := <-channel
		fmt.Printf("START message to send to everyone: %s\n", start)

		node.broadcast(start)
	}
}

func (node *Node) propagatePredictionMessage(channel chan *Packet) {

}

/*
func (node *Node) HandleStartMessage(packet *Packet, senderAddress net.UDPAddr) {
	//fmt.Println("Handling incoming *START* message..")
	//fmt.Println(packet.WaitingForLocalPredictions, packet.End, packet.Status)
	if node.CurrentStatus.CurrentState == LocalPredictionsTerminated && node.CurrentStatus.CurrentRound < packet.WaitingForLocalPredictions.RoundID {
		node.RemainingPeers = node.InitPeersMap()
		node.CurrentStatus.CurrentRound = packet.WaitingForLocalPredictions.RoundID
		node.CurrentStatus.CurrentState = WaitingForLocalPredictions

		toPropagate := &Packet {
			WaitingForLocalPredictions: &StartMessage{
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
			node.CurrentStatus.CurrentState = LocalPredictionsTerminated
			//fmt.Println("finished (from startHandler)!")
			fmt.Printf("DEBUG: %s %s\n", node.CurrentStatus, node.RemainingPeers)
		} ()
	}
}
*/

/*
func (node *Node) TriggerEndRoundMessagePropagation() {
	ticker := time.NewTicker(1*time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <- ticker.C:
			if node.Leader.peerAddress == node.Address {
				res := node.allExternalPredictionsObtained()
				fmt.Println(" ************** result of the check: ", res)
				if res && node.CurrentStatus.StatusValue.CurrentState == LocalPredictionsTerminated {
					fmt.Println("Preparing END message ..")
					endPacket := &Packet{
						End: &EndRoundMessage{
							RoundID: node.CurrentStatus.StatusValue.CurrentRound,
						},
					}

					fmt.Println(" ***** END message prepared: ", endPacket)
					node.EndRoundMessageHandler <- endPacket
				}
			}
		}
	}
}
*/