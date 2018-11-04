package main

import (
	"time"
	"net"
)

// if a probe is received, then I am already in the same round as the leader (since the leader
// sends probes only after it goes to the next round, which can occur only after all ACKs of the previous are received)
// this should be done only by the follower nodes
func (node *Node) HandleReceivedProbe(packet *Packet, senderAddress net.UDPAddr) {
	if node.CurrentStatus.StatusValue.CurrentState == LocalPredictionsTerminated {
		node.CurrentStatus.mux.Lock()
		//newPrediction := []float64{node.LocalDecision.scores[node.DetectionClass], node.LocalDecision.boundingBoxCoefficients[node.DetectionClass]}

		statusToSend := Status{
			CurrentRound: node.CurrentStatus.StatusValue.CurrentRound,
			CurrentState: node.CurrentStatus.StatusValue.CurrentState,
			CurrentPrediction: SinglePrediction {
				Value: []float64{node.LocalDecision.scores[node.DetectionClass], node.LocalDecision.boundingBoxCoefficients[node.DetectionClass]},
			},
		}
		statusPacket := &Packet{
			//Status: &node.CurrentStatus.StatusValue,
			Status: &statusToSend,
		}

		node.CurrentStatus.mux.Unlock()

		node.StatusHandler <- statusPacket
	}
}

// periodically send probes to the peers
// this should be done only by the leader node
func (node *Node) periodicPeersProbe() {
	ticker := time.NewTicker(500*time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case _ = <- ticker.C:
			if node.isLeader() && !node.allExternalPredictionsObtained() {
				node.probeFollowers()
			}
		}
	}
}

// probe the peers that still need to send me their prediction
func (node *Node) probeFollowers() {
	packet := &Packet {
		Probe: &ProbeMessage {
			RoundID: node.CurrentStatus.StatusValue.CurrentRound,
		},
	}

	peersToProbe := make(map[string]Peer)
	for k, v := range node.Peers {
		if prediction := node.ExternalPredictions.getPrediction(k); prediction == nil {
			peersToProbe[k] = v
		}
	}

	node.sendToPeers(packet, peersToProbe)
}