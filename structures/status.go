package structures

import (
	"net"
	"time"
	"fmt"
)

// need to make sure that this is really called only once: what about locking the status or the round?
func (node *Node) StartNewRound() {
	node.CurrentStatus.mux.Lock()

	defer node.CurrentStatus.mux.Unlock()

	node.startNewRoundNoLOCK()
}

func (node *Node) startNewRoundNoLOCKwithRound(newRound uint64) {
	fmt.Println("status (round) before changing round: ", node.CurrentStatus.StatusValue.CurrentRound)
	//fmt.Println("START NEW ROUND: BEGIN (no lock) ***************************")

	// keep the status as this until the end of this round (and the beginning of the new one)
	node.ReceivedLocalPredictions = 0
	node.CurrentStatus.StatusValue.CurrentRound = newRound
	node.CurrentStatus.StatusValue.CurrentState = WaitingForLocalPredictions

	// re-initialize the set of peers that should send me the prediction in the current round
	node.RemainingPeers = node.InitPeersMap()
	node.ExternalPredictions = node.InitExternalPredictionsMap()

	//node.Timeout = false
	node.Timeout.mux.Lock()
	node.Timeout.Timeout = false
	node.Timeout.mux.Unlock()
	go func() {
		fmt.Println("starting timeout goroutine .. (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
		//time.Sleep(10*time.Second)
		//node.Timeout = true
		//fmt.Println("timeout true!")
		//node.PredictionsAggregatorHandler <- struct{}{}

		select {
		case <- node.TimeoutHandler:
			fmt.Println("forcing timeout goroutine to terminate! (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
			return
		case <- time.After(10 * time.Second):
			//node.Timeout = true
			node.Timeout.mux.Lock()
			//node.Timeout = true
			node.Timeout.Timeout = true
			node.Timeout.mux.Unlock()
			fmt.Println("timeout true! (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
			node.PredictionsAggregatorHandler <- struct{}{}
			fmt.Println("terminating timeout goroutine.. (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
			return
		}
	} ()

	fmt.Println("status (round) after changing round: ", node.CurrentStatus.StatusValue.CurrentRound)
	//fmt.Println("START NEW ROUND: END (no lock) ***************************")
}

func (node *Node) startNewRoundNoLOCK() {
	newRound := node.CurrentStatus.StatusValue.CurrentRound+1
	node.startNewRoundNoLOCKwithRound(newRound)
}

func (node *Node) timeout() bool {
	//return node.Timeout
	node.Timeout.mux.Lock()
	defer node.Timeout.mux.Unlock()

	return node.Timeout.Timeout
}

// access to CurrentStatus: is there any concurrent access to CurrentRound?
// CurrentRound is modified when a new round starts.
// New rounds start after all external predictions have been received and after local prediction has been performed (state = 2)
func (node *Node) HandleReceivedStatus(packet *Packet, senderAddress net.UDPAddr) {
	// is any lock needed in order to access the status?
	//fmt.Println("Handling status message..")
	// the external prediction must not come from the leader..
	if node.isLeader() && senderAddress.String() != node.Leader.PeerAddress.String() &&
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