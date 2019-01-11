package structures

import (
	"net"
	"time"
	"fmt"
)

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

	node.Timeout.mux.Lock()
	node.Timeout.Timeout = false
	node.Timeout.mux.Unlock()
	go func() {
		fmt.Println("starting timeout goroutine .. (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
		//time.Sleep(10*time.Second)
		//node.TimeoutWrapper = true
		//fmt.Println("timeout true!")
		//node.PredictionsAggregatorHandler <- struct{}{}

		select {
		case <- node.TimeoutHandler:
			fmt.Println("forcing timeout goroutine to terminate! (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
			return
		case <- time.After(10 * time.Second):
			//node.TimeoutWrapper = true
			node.Timeout.mux.Lock()
			//node.TimeoutWrapper = true
			node.Timeout.Timeout = true
			node.Timeout.mux.Unlock()
			fmt.Println("timeout true! (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
			node.PredictionsAggregatorHandler <- struct{}{}
			fmt.Println("terminating timeout goroutine.. (round ", node.CurrentStatus.StatusValue.CurrentRound, ")")
			return
		}
	} ()

	fmt.Println("status (round) after changing round: ", node.CurrentStatus.StatusValue.CurrentRound)
}

func (node *Node) timeout() bool {
	//return node.TimeoutWrapper
	node.Timeout.mux.Lock()
	defer node.Timeout.mux.Unlock()

	return node.Timeout.Timeout
}

// access to CurrentStatus: is there any concurrent access to CurrentRound?
// CurrentRound is modified when a new round starts.
// New rounds start after all external predictions have been received and after local prediction has been performed (state = 2)
func (node *Node) HandleReceivedStatus(packet *Packet, senderAddress net.UDPAddr) {
	fmt.Println("received status: ", packet.Status, " from ", senderAddress.String())

	node.CurrentStatus.mux.Lock()
	defer node.CurrentStatus.mux.Unlock()

	if ackReceived, _ := node.ReceivedAcknowledgements[node.BaseStationAddress.String()]; ackReceived {
		return
	}

	if packet.Status.CurrentRound > node.CurrentStatus.StatusValue.CurrentRound {
		//node.StartRoundHandler <- packet.Status.CurrentRound // deadlock may occur ... or the StartRound routing will simply wait for this whole message to be processed
		// what if I want the round to be changed here? So that future code will be affected as well. I may call the StartRound function directly.
		node.ReceivedLocal.mux.Lock()
		node.StartRound(packet.Status.CurrentRound, false, false)
		node.ReceivedLocal.mux.Unlock()
	} else if packet.Status.CurrentRound < node.CurrentStatus.StatusValue.CurrentRound {
		probeMessage := &Packet {
			Probe: &ProbeMessage{
				RoundID: node.CurrentStatus.StatusValue.CurrentRound,
			},
		}

		node.sendToPeer(probeMessage, node.Peers[senderAddress.String()])
	}

	if node.isLeader() && packet.Status.CurrentRound >= node.CurrentStatus.StatusValue.CurrentRound {
		if _, exists := node.RemainingPeers.v[senderAddress.String()]; exists || senderAddress.String() == node.Address.String() {
			node.RemoveRemainingPeer(senderAddress)

			fmt.Println("remaining needed positive votes: ", node.RemainingNeededPositiveVotes)
			if packet.Status.CurrentPrediction.Value[0] >= node.ConfidenceThresholds[node.DetectionClass] {
				node.RemainingNeededPositiveVotes = node.RemainingNeededPositiveVotes-1

				if node.RemainingNeededPositiveVotes == 0 {
					fmt.Println("Object present! Creating prediction..")
					node.FinalPredictionHandler <- Present
				}
			}

			//if node.RemainingNeededPositiveVotes == 0 {
			//	fmt.Println("Object present! Creating prediction..")
			//	node.FinalPredictionHandler <- Present
			//} else if node.RemainingNeededPositiveVotes > 0 {
			if node.RemainingNeededPositiveVotes > 0 {
				node.ReceivedLocal.mux.Lock()
				if len(node.RemainingPeers.v) == 0 && node.ReceivedLocal.Value {
					fmt.Println("Object absent! Creating prediction..")
					node.FinalPredictionHandler <- Absent
				}
				node.ReceivedLocal.mux.Unlock()
			}
		}
	}
}

func (node *Node) TimeoutTrigger() {
	for {
		timer := time.NewTimer(5*time.Second)
		fmt.Println("new timer has started!")

		for {
			select {
			case <- timer.C: // if timeout occurs, change round (this will trigger StartRound)
				if node.isLeader() {
					fmt.Println("timeout occurred!")
					node.FinalPredictionHandler <- Timeout
				} else {
					node.CurrentStatus.mux.Lock()
					node.StartRoundHandler <- node.CurrentStatus.StatusValue.CurrentRound+1
					node.CurrentStatus.mux.Unlock()
				}
			case <- node.TimeoutHandler: // if new round has already started, stop the timer and start again (StartRound is triggering this)
				fmt.Println("no timeout here!")
				timer.Stop()

			}

			break
		}

	}
}

func (status *StatusConcurrent) getRoundIDConcurrent() uint64 {
	status.mux.Lock()
	defer status.mux.Unlock()

	return status.StatusValue.CurrentRound
}

// handles changes of round triggered by internal processing (leader) or external (start round message received from previous leader)
func (node *Node) HandleStartRound() {
	for {
		roundID := <- node.StartRoundHandler
		node.StartRound(roundID, true, true)
	}
}

func (node *Node) StartRound(newRound uint64, needStatusLock, needReceivedLocalLock bool) {
	fmt.Println("status (round) before changing round: ", node.CurrentStatus.StatusValue.CurrentRound)

	node.RemainingPeers = node.InitPeersMap()
	node.RemainingNeededPositiveVotes = (int8(len(node.Peers)) + 1) / 3 + 1

	for k, _ := range node.ReceivedAcknowledgements {
		node.ReceivedAcknowledgements[k] = false
	}

	// accessing node.CurrentStatus
	if needStatusLock {
		node.CurrentStatus.mux.Lock()
	}

	if node.CurrentStatus.StatusValue.CurrentRound < newRound {
		node.CurrentStatus.StatusValue.CurrentRound = newRound
		node.CurrentStatus.StatusValue.CurrentState = WaitingForLocalPredictions
		node.ReceivedLocalPredictions=0
	}

	if needStatusLock {
		node.CurrentStatus.mux.Unlock()
	}

	// accessing node.ReceivedLocal
	if needReceivedLocalLock {
		node.ReceivedLocal.mux.Lock()
	}

	node.ReceivedLocal.Value = false

	if needReceivedLocalLock {
		node.ReceivedLocal.mux.Unlock()
	}

	// reset the timeout
	node.TimeoutHandler <- struct{}{}

	fmt.Println("status (round) after changing round: ", node.CurrentStatus.StatusValue.CurrentRound)
}
