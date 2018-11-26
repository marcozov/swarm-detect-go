package main

import (
	"net"
)

// need to make sure that this is really called only once: what about locking the status or the round?
func (node *Node) startNewRound() {
	node.CurrentStatus.mux.Lock()
	defer node.CurrentStatus.mux.Unlock()

	//fmt.Println("START NEW ROUND: BEGIN ***************************")

	newRound := node.CurrentStatus.StatusValue.CurrentRound+1

	// keep the status as this until the end of this round (and the beginning of the new one)
	node.ReceivedLocalPredictions = 0
	node.CurrentStatus.StatusValue.CurrentRound = newRound
	node.CurrentStatus.StatusValue.CurrentState = WaitingForLocalPredictions

	// re-initialize the set of peers that should send me the prediction in the current round
	node.RemainingPeers = node.InitPeersMap()
	node.ExternalPredictions = node.InitExternalPredictionsMap()

	//fmt.Println("START NEW ROUND: END ***************************")
}

func (node *Node) startNewRoundNoLOCK() {
	//fmt.Println("START NEW ROUND: BEGIN (no lock) ***************************")

	newRound := node.CurrentStatus.StatusValue.CurrentRound+1

	// keep the status as this until the end of this round (and the beginning of the new one)
	node.ReceivedLocalPredictions = 0
	node.CurrentStatus.StatusValue.CurrentRound = newRound
	node.CurrentStatus.StatusValue.CurrentState = WaitingForLocalPredictions

	// re-initialize the set of peers that should send me the prediction in the current round
	node.RemainingPeers = node.InitPeersMap()
	node.ExternalPredictions = node.InitExternalPredictionsMap()

	//fmt.Println("START NEW ROUND: END (no lock) ***************************")
}

// access to CurrentStatus: is there any concurrent access to CurrentRound?
// CurrentRound is modified when a new round starts.
// New rounds start after all external predictions have been received and after local prediction has been performed (state = 2)
func (node *Node) HandleReceivedStatus(packet *Packet, senderAddress net.UDPAddr) {
	// is any lock needed in order to access the status?
	//fmt.Println("Handling status message..")
	// the external prediction must not come from the leader..
	if node.isLeader() && senderAddress.String() != node.Leader.peerAddress.String() &&
		node.CurrentStatus.getRoundIDConcurrent() == packet.Status.CurrentRound &&
		node.GetPeer(senderAddress) != nil {
			if existingPrediction := node.ExternalPredictions.getPrediction(senderAddress.String()); existingPrediction == nil {
				//fmt.Println("new prediction: ", packet.Status.CurrentPrediction)
				node.ExternalPredictions.addPrediction(senderAddress.String(), packet.Status.CurrentPrediction)
				// send signal to check that all external predictions have been received
				node.PredictionsAggregatorHandler <- struct{}{}
			}
	}
}

func (status *StatusConcurrent) getRoundIDConcurrent() uint64 {
	status.mux.Lock()
	defer status.mux.Unlock()

	return status.StatusValue.CurrentRound
}