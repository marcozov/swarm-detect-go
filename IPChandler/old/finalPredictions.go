package main

import (
	"fmt"
	"net"
	"time"
)

// receive final prediction, send ACK to the leader
// done by the follower nodes
func (node *Node) HandleReceivedFinalPrediction(packet *Packet, senderAddress net.UDPAddr) {
	node.CurrentStatus.mux.Lock()
	defer node.CurrentStatus.mux.Unlock()
	//roundID := node.CurrentStatus.getRoundIDConcurrent()
	roundID := node.CurrentStatus.StatusValue.CurrentRound
	if !node.isLeader() && packet.FinalPrediction.ID == roundID {
		finalPrediction := packet.FinalPrediction

		ack := &Packet{
			Ack: &AcknowledgementMessage{ID: packet.FinalPrediction.ID,},
		}
		node.sendToPeer(ack, *node.GetPeer(senderAddress))

		fmt.Println("************** FINAL PREDICTION RECEIVED FROM THE LEADER: ID: ", finalPrediction.ID, ", Prediction.Value: ", finalPrediction.Prediction.Value, ", normalized value: ", finalPrediction.Prediction.Value[0]/finalPrediction.Prediction.Value[1], ", MAX SCORE: ", finalPrediction.Prediction.Value[2])
		//node.startNewRound()
		node.startNewRoundNoLOCK()
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
		node.EndRoundHandler <- struct{}{}
		//fmt.Println("After sending message to EndRoundHandler")
	}
}

// send stuff to the channel whenever an element is added to ExternalPredictions
// this should be done only by the leader node
func (node *Node) AggregateAllPredictions() *SinglePrediction {
	for {
		// wait until I received all the external predictions and my state is LocalPredictionsTerminated
		//fmt.Println("AggregateAllPredictions (beginning) ..")
		// 2 sources: communications.go (finish local prediction), status.go (received external prediction)
		<- node.PredictionsAggregatorHandler
		if   node.isLeader() &&
			(node.allExternalPredictionsObtained() && node.CurrentStatus.StatusValue.CurrentState == LocalPredictionsTerminated) {
			// aggregate predictions, send the final

			N := len(node.Peers)
			majority := N/3

			// node.LocalDecision can be accessed concurrently through this function, by updateOpinionVector, by HandleReceivedProbe
			localScore, localBBcoefficient, localEntropy := node.LocalDecision.getOpinion(node.DetectionClass)
			newPrediction := []float64{localScore*localBBcoefficient, localBBcoefficient}
			if newPrediction[0] >= node.ConfidenceThresholds[node.DetectionClass] {
				majority = majority-1
			}

			maxScore := localScore
			fmt.Println("local opinion for data fusion: score=", localScore, ", localBBcoefficient=", localBBcoefficient, ", localEntropy=", localEntropy)
			for host, _ := range node.Peers {
				externalPrediction := node.GetPrediction(host)
				if externalPrediction != nil {
					if externalPrediction.Value[0] >= node.ConfidenceThresholds[node.DetectionClass] {
						majority = majority-1
					}

					// not really useful ....
					fmt.Println("external prediction for data fusion, coming from ", host, ": ", externalPrediction)
					newPrediction[0] = newPrediction[0] + externalPrediction.Value[0]*externalPrediction.Value[1]
					newPrediction[1] = newPrediction[1] + externalPrediction.Value[1]
					if externalPrediction.Value[0] > maxScore {
						maxScore = externalPrediction.Value[0]
					}
				}
			}

			decision := 0.0
			if majority < 0 {
				decision = 1.0
			}
			prediction := FinalPredictionMessage{
				ID: node.CurrentStatus.StatusValue.CurrentRound,
				Prediction: &SinglePrediction {
					//Value: []float64{newPrediction[0], newPrediction[1], maxScore},
					Value: []float64{decision, newPrediction[1], maxScore},
				},
			}

			fmt.Println("************** FINAL PREDICTION THAT IS PROPAGATED TO THE OTHER HOSTS: ID: ", prediction.ID, ", Prediction.Value: ", prediction.Prediction.Value, ", normalized value: ", prediction.Prediction.Value[0]/prediction.Prediction.Value[1], ", MAX SCORE: ", maxScore)
			// trigger the final predictions forwarding

			// this is executed
			go node.propagateFinalPredictions(prediction)

			// wait until I receive all acks
			for {
				// triggered everytime an ACK is received
				//fmt.Println("***** BEFORE receiving EndRoundHandler")
				<- node.EndRoundHandler
				//fmt.Println("received endRoundHandler. Peers: ", node.RemainingPeers.v)
				//fmt.Println("***** AFTER receiving EndRoundHandler")
				if node.RemainingPeers.Length() == 0 {
					node.FinalPredictionPropagationTerminate <- struct{}{}
					node.startNewRound()
					break
				}
			}
		} else {
			if node.isLeader() {
				//fmt.Println("External predictions: ", node.ExternalPredictions.v, ", state: ", node.CurrentStatus.StatusValue.CurrentState)
			}
		}
	}

	return nil
}

// checking whether all the external predictions have been received: necessary condition before performing the aggregation
func (node *Node) allExternalPredictionsObtained() bool {
	node.ExternalPredictions.mux.Lock()
	defer node.ExternalPredictions.mux.Unlock()
	for peer, _ := range node.Peers {
		if _, exists := node.ExternalPredictions.v[peer]; !exists {
		//if prediction := node.GetPrediction(peer); prediction == nil {
		//	fmt.Println("the peer ", peer, " did not send the prediction!")
			return false
		}
	}

	return true
}

// this function should terminate before the new round is started, otherwise it will
// keep sending messages in the following rounds too
func (node *Node) propagateFinalPredictions(prediction FinalPredictionMessage) {
	packet := &Packet { FinalPrediction: &prediction}

	ticker := time.NewTicker(1500*time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case _ = <- ticker.C:
			fmt.Println("checking length! ...")
			//fmt.Println("will propagate final predictions end? ", node.RemainingPeers.v)
			if node.RemainingPeers.Length() == 0 {
				//fmt.Println("ending propagate final predictions")
				return
			}

			node.sendToPeers(packet, node.RemainingPeers.v)
		case _ = <- node.FinalPredictionPropagationTerminate:
			return
		}

	}
}