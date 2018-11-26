package main

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
	if node.CurrentStatus.StatusValue.CurrentState == LocalPredictionsTerminated {
		localScore, localBBcoefficient, localEntropy := node.LocalDecision.getOpinion(node.DetectionClass)
		statusToSend := Status{
			CurrentRound: node.CurrentStatus.StatusValue.CurrentRound,
			CurrentState: node.CurrentStatus.StatusValue.CurrentState,
			CurrentPrediction: SinglePrediction {
				Value: []float64{localScore, localBBcoefficient, localEntropy},
			},
		}
		node.CurrentStatus.mux.Unlock()

		statusPacket := &Packet{
			Status: &statusToSend,
		}

		fmt.Println("probe received by ", senderAddress.String(), ". STATUS to propagate: ", statusPacket)
		node.sendToPeer(statusPacket, node.Peers[senderAddress.String()])
	} else {
		node.CurrentStatus.mux.Unlock()
	}
}

// periodically send probes to the peers
// this should be done only by the leader node
func (node *Node) periodicPeersProbe() {
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
		if _, exists := node.ExternalPredictions.v[k]; !exists {
			peersToProbe[k] = v
		}
	}
	node.sendToPeers(packet, peersToProbe)
}