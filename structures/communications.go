package structures

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

		fmt.Println("received local predictions: ", node.ReceivedLocalPredictions)

		if node.ReceivedLocalPredictions == 5 {
			node.CurrentStatus.mux.Lock()
			node.ReceivedLocal.mux.Lock()
			fmt.Println("current state: ", node.CurrentStatus.StatusValue.CurrentState)
			// probably no need to save the state, since it is sent only to one host (the leader)
			if node.CurrentStatus.StatusValue.CurrentState == WaitingForLocalPredictions {
				fmt.Println("I finished accumulating data for this round! (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
				node.CurrentStatus.StatusValue.CurrentState = LocalPredictionsTerminated
				node.ReceivedLocal.Value = true

				// handling a local prediction as an external one!
				localScore, _ := node.LocalDecision.getOpinion(node.DetectionClass)
				//presence := 0.0
				//if localScore >= node.ConfidenceThresholds[node.DetectionClass] {
				//	presence = 1
				//}

				localPrediction := &Packet{
					Status: &Status{
						CurrentRound: node.CurrentStatus.StatusValue.CurrentRound,
						CurrentState: node.CurrentStatus.StatusValue.CurrentState,
						//CurrentPrediction: SinglePrediction{ Value: []float64{ presence}},
						CurrentPrediction: SinglePrediction{ Value: []float64{ localScore}},
					},
				}

				counter := &IntWrapper{ v: 0}
				fmt.Println("sending local prediction as external message", localPrediction.Status.CurrentPrediction, localScore)
				node.PacketHandler <- PacketChannelMessage{localPrediction, node.Address, counter}
			}

			node.ReceivedLocalPredictions = 0
			node.CurrentStatus.mux.Unlock()
			node.ReceivedLocal.mux.Unlock()
		}
	}
}

func (node *Node) ProcessMessage(channel chan PacketChannelMessage) {
	for {
		message := <- channel
		packet := message.Packet
		senderAddress := message.SenderAddress
		//n := rand.Intn(100000000)

		if packet.Probe != nil {
			node.HandleReceivedProbe(packet, *senderAddress)
		} else if packet.Status != nil {
			node.HandleReceivedStatus(packet, *senderAddress)
		} else if packet.FinalPrediction != nil {
			node.HandleReceivedFinalPrediction(packet, *senderAddress)
		} else if packet.Ack != nil {
			node.HandleReceivedAcknowledgement(packet, *senderAddress)
		} else if packet.StartRound != nil {
			node.StartRoundHandler <- packet.StartRound.RoundID
		}

		//fmt.Println("************** FINISHED HANDLING THE RECEIVED MESSAGE ", n, " *****************")
	}
}

type IntWrapper struct {
	v int
	mux sync.Mutex
}

func (node *Node) HandleIncomingMessages() {
	counter := &IntWrapper{ v: 0}
	for {
		//fmt.Println("waiting for new message ..")
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

		//fmt.Println("******** new message received: ", packet)
		node.PacketHandler <- PacketChannelMessage{packet, senderAddress, counter}

		//go node.processMessage(packet, senderAddress, counter)
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

func SendToPeer(packet *Packet, peer *net.UDPAddr, connection *net.UDPConn) {
	packetBytes, err := protobuf.Encode(packet)
	if err != nil {
		panic(fmt.Sprintf("Error in encoding the message: %s\npacket: %s\npacketBytes: %s\n", err, packet, packetBytes))
	}

	_, err = connection.WriteToUDP(packetBytes, peer)
	if err != nil {
		panic(fmt.Sprintf("Error in sending udp data: %s\npacket: %s\npacketBytes: %s\n", err, packet, packetBytes))
	}
}

func (node *Node) sendToPeer(packet *Packet, peer Peer) {
	SendToPeer(packet, peer.PeerAddress, node.Connection)
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