package main

import (
	"net"
	"fmt"
	"strings"
)

type State int32

const (
	WaitingForLocalPredictions  = iota+1
	LocalPredictionsTerminated
)

const DetectionClasses int = 91

type Node struct {
	Address    *net.UDPAddr      // peerAddress on which peers send messages
	Connection *net.UDPConn      // to receive data from peers and to send data
	Name       string            // name of the gossiper
	Peers      map[string]Peer   // set of known peers
	RemainingPeers *SafeMapPeers // set of peers from which the host needs the prediction (for the current round)
	Leader Peer
	CurrentStatus *StatusConcurrent
	LocalDecision LocalOpinionVector // accumulator of the local predictions
	ReceivedLocalPredictions int
	ExternalPredictions *SafeMapSinglePredictions
	ExternalPredictionsHandler chan *PacketWithSender
	PredictionsAggregatorHandler chan struct{}
	EndRoundHandler chan struct{}
	FinalPredictionPropagationTerminate chan struct{}
	DetectionClass int // the object that we are trying to detect
}

type SinglePrediction struct {
	Value []float64
}

type Peer struct {
	peerAddress    *net.UDPAddr
}

func NewNode(address, name, peers string, detectionClass int) *Node {
	udpAddress, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		panic (fmt.Sprintf("Address not valid: %s", err))
	}

	udpConnection, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		panic (fmt.Sprintf("Error in opening UDP listener: %s", err))
	}

	myPeers := make(map[string]Peer)
	node := &Node{
		Address:    udpAddress,
		Connection: udpConnection,
		Name:       name,
		Peers:      myPeers,
		CurrentStatus:
			&StatusConcurrent{
				StatusValue: Status{
					CurrentRound:      1,
					CurrentState:      WaitingForLocalPredictions,
					CurrentPrediction: SinglePrediction{ Value: []float64{0, 0}},
				},
			},
		LocalDecision: LocalOpinionVector{
			alpha: 0.5,
			scores: [DetectionClasses]float64{},
			boundingBoxCoefficients: [DetectionClasses]float64{},
		},

		ReceivedLocalPredictions: 0,
		DetectionClass: detectionClass,
	}


	for _, peer := range strings.Split(peers, ",") {
		myPeers := node.Peers

		peerAddress, err := net.ResolveUDPAddr("udp4", peer)
		if err != nil {
			panic(fmt.Sprintf("Error in parsing the UDP peerAddress: %s", err))
		}

		peerWrapper := Peer {
			peerAddress: peerAddress,
		}

		myPeers[peerAddress.String()] = peerWrapper
		node.Peers = myPeers
	}

	node.RemainingPeers = node.InitPeersMap()
	node.ExternalPredictions = node.InitExternalPredictionsMap()

	return node
}

func (node *Node) isLeader() bool {
	return node.Address.String() == node.Leader.peerAddress.String()
}

func (status *Status) String() string {
	return fmt.Sprintf("Current round: %d, state of the round: %s, current prediction: %s", status.CurrentRound, status.CurrentState, status.CurrentPrediction)
}