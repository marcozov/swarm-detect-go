package main

import (
	"net"
	"fmt"
	"strings"
	"sync"
	"time"
)

type State int32

const (
	WaitingForLocalPredictions  = iota+1
	LocalPredictionsTerminated
	FinalPredictionTerminated
)

const DetectionClasses int = 91

type Node struct {
	Address    *net.UDPAddr      // peerAddress on which peers send messages
	Connection *net.UDPConn      // to receive data from peers and to send data
	Name       string            // name of the gossiper
	Peers      map[string]Peer   // set of known peers
	RemainingPeers *SafeMapPeers // set of peers from which the host needs the prediction (for the current round)
	Leader Peer
	//CurrentStatus *Status
	CurrentStatus *StatusConcurrent
	StartHandler chan *Packet
	StatusHandler chan *Packet
	EndRoundMessageHandler chan *Packet
	LocalDecision LocalOpinionVector // accumulator of the local predictions
	ReceivedLocalPredictions int
	//ExternalPredictions map[string]SinglePredictionWithSender
	ExternalPredictions *SafeMapSinglePredictions
	//ExternalPredictionsHandler chan SinglePredictionWithSender
	ExternalPredictionsHandler chan *Packet
	PredictionsAggregatorHandler chan struct{}
	EndRoundHandler chan struct{}
	DetectionClass int // the object that we are trying to detect
}

type SinglePrediction struct {
	Value []float64
}

type SinglePredictionWithSender struct {
	Prediction SinglePrediction
	Sender string
}

type LocalOpinionVector struct {
	alpha float64
	scores [DetectionClasses]float64
	boundingBoxCoefficients [DetectionClasses]float64
	mux sync.Mutex
}

type Peer struct {
	peerAddress    *net.UDPAddr
}

func (node *Node) updateOpinionValue(class uint64, coefficient, score float64) {
	node.LocalDecision.mux.Lock()
	defer node.LocalDecision.mux.Unlock()

	node.LocalDecision.scores[class] = node.LocalDecision.scores[class]*(1-coefficient) + score*coefficient
}

//func (node *Node) AddAcknowledgedPeer(peer Peer) {
//	node.AcknowledgedPeers.addPeer(peer)
//}
//
//func (node *Node) GetAcknowledgedPeer(peerAddress net.UDPAddr) *Peer {
//	return node.AcknowledgedPeers.getPeer(peerAddress)
//}

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
		//ExternalPredictions: make(map[string]SinglePredictionWithSender),
		DetectionClass: detectionClass,
	}


	for _, peer := range strings.Split(peers, ",") {
		//node.addPeer(peer)
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

//func (node *Node) addPeer(peer string) {
//	myPeers := node.Peers
//
//	peerAddress, err := net.ResolveUDPAddr("udp4", peer)
//	if err != nil {
//		panic(fmt.Sprintf("Error in parsing the UDP peerAddress: %s", err))
//	}
//
//	peerWrapper := Peer {
//		peerAddress: peerAddress,
//	}
//
//	myPeers[peerAddress.String()] = peerWrapper
//	node.Peers = myPeers
//}

func (node *Node) statusDebug() {
	for {
		fmt.Println("STATUS DEBUG: ", node.CurrentStatus)
		time.Sleep(5 * time.Second)
	}
}

func (status *Status) String() string {
	return fmt.Sprintf("Current round: %d, state of the round: %s, current prediction: %s", status.CurrentRound, status.CurrentState, status.CurrentPrediction)
}

func entropy(scores []float32) float32 {
	return 0
}