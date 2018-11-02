package main

import (
	"fmt"
	"net"
	"time"
)

// handling round advancement
func (node *Node) HandleRounds() {
	ticker := time.NewTicker(1*time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <- ticker.C:
			if node.CurrentStatus.StatusValue.CurrentState == Finish && node.RemainingPeers.Length() == 0 {
				//node.startRound(1)
				node.startNewRound()
			}
		}
	}
}

// need to make sure that this is really called only once: what about locking the status or the round?
//func (node *Node) startRound(newRound uint64) {
func (node *Node) startNewRound() {
	node.CurrentStatus.mux.Lock()
	defer node.CurrentStatus.mux.Unlock()

	newRound := node.CurrentStatus.StatusValue.CurrentRound+1
	//if node.CurrentStatus.CurrentRound < newRound {

		// keep the status as this until the end of this round (and the beginning of the new one)
	//node.CurrentStatus.CurrentPrediction.Value[0] = node.LocalDecision.scores[node.DetectionClass]
	//node.CurrentStatus.CurrentPrediction.Value[1] = node.LocalDecision.boundingBoxCoefficients[node.DetectionClass]
	node.ReceivedLocalPredictions = 0
	node.CurrentStatus.StatusValue.CurrentRound = newRound
	node.CurrentStatus.StatusValue.CurrentState = Start

	node.ExternalPredictions = make(map[string]SinglePredictionWithSender)
	//}

	// re-initialize the set of peers that should send me the prediction in the current round
	node.RemainingPeers = node.InitPeersMap()
}



// sending the accumulated local prediction
func (node *Node) PropagateLocalPredictions() {
	ticker := time.NewTicker(1*time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <- ticker.C:
			if node.CurrentStatus.StatusValue.CurrentState == Finish {
				statusPacket := &Packet {
					Status: &node.CurrentStatus.StatusValue,
				}
				node.StatusHandler <- statusPacket
			}
		}
	}
}

// wrapper function to handle status message
func (node *Node) propagateStatusMessage(channel chan *Packet) {
	for {
		status := <- channel
		//fmt.Printf("STATUS message to send to everyone: %s. Remaining peers: %s\n", status, node.RemainingPeers)

		node.sendToPeers(status, node.RemainingPeers.v)
	}
}

// wrapper function to handle external predictions
// useful for concurrency issues
func (node *Node) updateExternalPredictions(channel chan SinglePredictionWithSender) {
	for {
		prediction := <- channel

		//fmt.Println("prediction: ", prediction)
		if _, ok := node.ExternalPredictions[prediction.Sender]; !ok {
			fmt.Println("new prediction: ", prediction)
			//fmt.Println("external predictions to consider in the final vote: ", node.ExternalPredictions)
			node.ExternalPredictions[prediction.Sender] = prediction
		}
	}
}

func (node *Node) HandleStatusMessageReceive(packet *Packet, senderAddress net.UDPAddr) {
	// is any lock needed in order to access the status?
	//fmt.Println("Handling status message..")
	//fmt.Println("status prediction: ", packet.Status.CurrentPrediction)

	//if node.CurrentStatus.CurrentRound == packet.Status.CurrentRound && packet.Status.CurrentState == Finish {
	//fmt.Println("problem accessing the status?")
	if node.CurrentStatus.StatusValue.CurrentRound == packet.Status.CurrentRound {
		//fmt.Println("nop, I accessed")
		if node.GetPeer(senderAddress) != nil {
			//node.RemoveRemainingPeer(senderAddress)

			// aggregate the information
			currentPrediction := packet.Status.CurrentPrediction
			node.ExternalPredictionsHandler <- SinglePredictionWithSender{
					Prediction: currentPrediction,
					Sender: senderAddress.String(),
				}
		}
	}
	//fmt.Println("finished handling status message ..")
}

func (node *Node) HandleEndRoundMessageReceive(packet *Packet, senderAddress net.UDPAddr) {
	myCurrentRond := node.CurrentStatus.StatusValue.CurrentRound
	packetRoundID := packet.End.RoundID
	fmt.Println("my current round: ", myCurrentRond, " packetRoundID: ", packetRoundID)
	if myCurrentRond == packetRoundID {
		node.RemoveRemainingPeer(senderAddress)
	}
}

func (node *Node) updateOpinionVector(localPrediction map[int][]float64) {
	alpha := node.LocalDecision.alpha
	//fmt.Println("Updating local opinion vector..")
	//for k, v := range localPrediction {
	for k := 1; k < DetectionClasses; k++ {
		v, exists := localPrediction[k]
		if !exists {
			v = []float64{0, 0}
		}

		//fmt.Println("")
		//fmt.Println("coefficients: ", node.LocalDecision.boundingBoxCoefficients)


		node.LocalDecision.boundingBoxCoefficients[k] = alpha*v[1] + node.LocalDecision.boundingBoxCoefficients[k]*(1-alpha)
		node.LocalDecision.scores[k] = alpha*v[0] + node.LocalDecision.scores[k]*(1-alpha)
	}
}