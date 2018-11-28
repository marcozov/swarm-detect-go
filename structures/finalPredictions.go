package structures

import (
	"fmt"
	"net"
	"time"
	"github.com/dedis/protobuf"
)

// receive final prediction, send ACK to the leader
// done by the follower nodes
func (node *Node) HandleReceivedFinalPrediction(packet *Packet, senderAddress net.UDPAddr) {
	fmt.Println("received final prediction: ", packet.FinalPrediction)

	node.CurrentStatus.mux.Lock()
	defer node.CurrentStatus.mux.Unlock()
	//roundID := node.CurrentStatus.getRoundIDConcurrent()
	roundID := node.CurrentStatus.StatusValue.CurrentRound

	//if !node.timeout() && packet.FinalPrediction.ID >= roundID {
	//	node.TimeoutHandler <- struct{}{}
	//}
	if !node.isLeader() && packet.FinalPrediction.ID == roundID {
		finalPrediction := packet.FinalPrediction

		ack := &Packet{
			Ack: &AcknowledgementMessage{ID: packet.FinalPrediction.ID,},
		}
		node.sendToPeer(ack, *node.GetPeer(senderAddress))

		fmt.Println("************** FINAL PREDICTION RECEIVED FROM THE LEADER: ID: ", finalPrediction.ID, ", Prediction.Value: ", finalPrediction.Prediction.Value, ", normalized value: ", finalPrediction.Prediction.Value[0]/finalPrediction.Prediction.Value[1], ", MAX SCORE: ", finalPrediction.Prediction.Value[2])
		//node.StartNewRound()
		//node.startNewRoundNoLOCK()
		node.startNewRoundNoLOCKwithRound(packet.FinalPrediction.ID)
	}
}

// receive and ACK ===> check whether all the peers have received my final prediction
// done by the leader
func (node *Node) HandleReceivedAcknowledgement(packet *Packet, senderAddress net.UDPAddr) {
	//fmt.Println("receiving ACK from ", senderAddress.String())
	if node.isLeader() {
		//fmt.Println("Before removing peer: ", node.RemainingPeers.v)
		node.RemoveRemainingPeer(senderAddress)
		//fmt.Println("After removing peer: ", node.RemainingPeers.v)
		//node.EndRoundHandler <- struct{}{}
		//fmt.Println("After sending message to EndRoundHandler")
	}
}

const (
	Timeout  = iota+1
	Absent
	Present
)


//func (node *Node) HandleFinalPredictions(channel chan int8) {
func (node *Node) HandleFinalPredictions() {
	for {
		fmt.Println("waiting for channel (final prediction handler) message ..")
		action := <- node.FinalPredictionHandler
		fmt.Println("received channel (final prediction handler) message!")

		// checking whether the base station has already received the prediction of this round
		//if node.ReceivedAcknowledgements[node.BaseStationAddress.String()] {
		//	continue
		//}

		prediction := FinalPredictionMessage{
			ID: node.CurrentStatus.StatusValue.CurrentRound,
		}

		if action == Timeout || action == Absent {
			prediction.Prediction = &SinglePrediction{
				//Value: []float64{newPrediction[0], newPrediction[1], maxScore},
				Value: []float64{0},
			}
		} else if action == Present {
				prediction.Prediction = &SinglePrediction {
					//Value: []float64{newPrediction[0], newPrediction[1], maxScore},
					Value: []float64{1},
				}
		} else {
			panic(fmt.Sprintf("action must be Timout, Absent or Present: %s", action))
		}

		finalPredictionPacket := &Packet { FinalPrediction: &prediction}
		baseStationACK := &Packet{}
		for {
			udpBuffer := make([]byte, 64)

			SendToPeer(finalPredictionPacket, node.BaseStationAddress, node.BaseStationConnection)
			fmt.Println("trying to read from base station ..")

			// the timeout should not be set too high. Otherwise the TriggerTimeout goroutine
			// will try to send multiple messages for node.FinalPredictionHandler channel and
			// that subroutine will be stuck until this function finishes handling the communication
			// with the base station
			err := node.BaseStationConnection.SetReadDeadline(time.Now().Add(500*time.Millisecond))
			if err != nil {
				panic(fmt.Sprintf("Error in setting the deadline: %s", err))
			}


			n, senderAddress, err := node.BaseStationConnection.ReadFromUDP(udpBuffer)

			if err != nil {
				if err.(net.Error).Timeout() {
					//panic("TIMEOUT !!!!!!!!! WILL NOT CRASH!!")
					continue
				}
				panic(fmt.Sprintf("error in reading UDP data: %s.\nudpBuffer: %v\nsenderAddress: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), n))
			}


			udpBuffer = udpBuffer[:n]
			err = protobuf.Decode(udpBuffer, baseStationACK)

			if err != nil {
				panic(fmt.Sprintf("error in decoding UDP data: %s\nudpBuffer: %v\nsenderAddress: %s\nbaseStationACK: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), baseStationACK, n))
			}

			if baseStationACK.Ack.ID == finalPredictionPacket.FinalPrediction.ID {
				node.ReceivedAcknowledgements[node.BaseStationAddress.String()] = true
				break
			}

		}



		//fmt.Println("before triggering start ..")
		node.StartRoundHandler <- prediction.ID + 1
		//fmt.Println("start triggered!")

		startNewRoundMessage := &Packet{
			StartRound: &StartRoundMessage{ RoundID: prediction.ID + 1},
		}
		node.sendToPeers(startNewRoundMessage, node.Peers)
	}
}