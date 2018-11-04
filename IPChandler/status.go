package main

import (
	"fmt"
	"net"
)

// need to make sure that this is really called only once: what about locking the status or the round?
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
	node.CurrentStatus.StatusValue.CurrentState = WaitingForLocalPredictions

	//node.ExternalPredictions = make(map[string]SinglePredictionWithSender)
	//}

	// re-initialize the set of peers that should send me the prediction in the current round
	node.RemainingPeers = node.InitPeersMap()
	node.ExternalPredictions = node.InitExternalPredictionsMap()
}

// wrapper function to handle status message
//func (node *Node) propagateStatusMessage(channel chan *Packet) {
func (node *Node) propagateStatusMessage() {
	for {
		//status := <- channel
		status := <- node.StatusHandler
		fmt.Printf("STATUS message to send to everyone: %s. Remaining peers: %s\n", status, node.RemainingPeers)

		node.sendToPeers(status, node.RemainingPeers.v)
	}
}

// wrapper function to handle external predictions
// useful for concurrency issues
//func (node *Node) updateExternalPredictions(channel chan SinglePredictionWithSender) {
func (node *Node) updateExternalPredictions() {
	for {
		predictionPacket := <- node.ExternalPredictionsHandler
		//prediction := predictionPacket.Status.CurrentPrediction
		prediction := predictionPacket.Packet.Status.CurrentPrediction
		sender := predictionPacket.SenderAddress
		fmt.Println("external prediction: ", prediction)
		if node.isLeader() {
			//fmt.Println("external prediction: ", prediction)
			//if _, ok := node.ExternalPredictions[prediction.Sender]; !ok {
			if existingPrediction := node.ExternalPredictions.getPrediction(sender.String()); existingPrediction == nil {
				fmt.Println("new prediction: ", prediction)
				//fmt.Println("external predictions to consider in the final vote: ", node.ExternalPredictions)
				//node.ExternalPredictions[prediction.Sender] = prediction
				node.ExternalPredictions.addPrediction(sender.String(), prediction)
				// send signal to check that all external predictions have been received
				node.PredictionsAggregatorHandler <- struct{}{}
			}
		}
	}
}

func (node *Node) HandleReceivedStatus(packet *Packet, senderAddress net.UDPAddr) {
	// is any lock needed in order to access the status?
	fmt.Println("Handling status message..")
	//fmt.Println("status prediction: ", packet.Status.CurrentPrediction)
	if senderAddress.String() != node.Leader.peerAddress.String() {

		//if node.CurrentStatus.CurrentRound == packet.Status.CurrentRound && packet.Status.CurrentState == LocalPredictionsTerminated {
		//fmt.Println("problem accessing the status?")
		if node.CurrentStatus.StatusValue.CurrentRound == packet.Status.CurrentRound {
			//fmt.Println("nop, I accessed")
			if node.GetPeer(senderAddress) != nil {
				//node.RemoveRemainingPeer(senderAddress)

				// aggregate the information
				//currentPrediction := packet.Status.CurrentPrediction
				//node.ExternalPredictionsHandler <- SinglePredictionWithSender{
				//	Prediction: currentPrediction,
				//	Sender:     senderAddress.String(),
				//}
				//packet.SenderAddress = &senderAddress
				packetWrapper := &PacketWithSender{
					Packet: packet,
					SenderAddress: &senderAddress,
				}
				node.ExternalPredictionsHandler <- packetWrapper
			}
		}
	}
	//fmt.Println("finished handling status message ..")
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