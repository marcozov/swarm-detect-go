package structures

import (
	"time"
	"net"
	"fmt"
)

// if a probe is received, then I am already in the same round as the leader (since the leader
// sends probes only after it goes to the next round, which can occur only after all ACKs of the previous are received)
// this should be done only by the follower nodes
func (node *Node) HandleReceivedProbe(packet *Packet, senderAddress net.UDPAddr) {
	node.CurrentStatus.mux.Lock()
	fmt.Println("handling probe: ", packet.Probe)

	// if Probe.RoundID > CurrentRound, start a new round
	if packet.Probe.RoundID > node.CurrentStatus.StatusValue.CurrentRound {
		// should be done before completing the rest of the function?
		node.StartRound(packet.Probe.RoundID, false, false)
	}

	if node.CurrentStatus.StatusValue.CurrentState == LocalPredictionsTerminated {
		localScore, localBBcoefficient := node.LocalDecision.getOpinion(node.DetectionClass)
		statusToSend := Status{
			CurrentRound: node.CurrentStatus.StatusValue.CurrentRound,
			CurrentState: node.CurrentStatus.StatusValue.CurrentState,
			CurrentPrediction: SinglePrediction {
				Value: []float64{localScore, localBBcoefficient},
			},
		}

		statusPacket := &Packet{
			Status: &statusToSend,
		}

		fmt.Println("probe received by ", senderAddress.String(), ". STATUS to propagate: ", statusPacket)
		node.sendToPeer(statusPacket, node.Peers[senderAddress.String()])
	}

	node.CurrentStatus.mux.Unlock()
}

// periodically send probes to the peers
// this should be done only by the leader node
func (node *Node) PeriodicPeersProbe() {
	ticker := time.NewTicker(1500*time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case _ = <- ticker.C:
			if node.isLeader() {
				node.probeFollowers()
			}
		}
	}
}

// probe the peers that still need to send me their prediction
func (node *Node) probeFollowers() {
	node.ExternalPredictions.mux.Lock()
	defer node.ExternalPredictions.mux.Unlock()
	roundID := node.CurrentStatus.getRoundIDConcurrent()
	packet := &Packet {
		Probe: &ProbeMessage {
			RoundID: roundID,
		},
	}

	peersToProbe := make(map[string]Peer)
	for k, v := range node.Peers {
		if _, exists := node.RemainingPeers.v[k]; exists {
			peersToProbe[k] = v
		}
	}

	fmt.Println("peers to probe: ", peersToProbe)
	node.sendToPeers(packet, peersToProbe)
}
