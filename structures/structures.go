package structures

import (
"net"
"fmt"
"strings"
	"time"
	"sync"
)

type State int32

const (
	WaitingForLocalPredictions  = iota+1
	LocalPredictionsTerminated
)

const DetectionClasses int = 91

type Node struct {
	Address                             *net.UDPAddr    // peerAddress on which peers send messages
	Connection                          *net.UDPConn    // to receive data from peers and to send data

	BaseStationAddress                  *net.UDPAddr    // address of the base station

	BaseStationLocalListenerAddress 	*net.UDPAddr    // address of the local connection that handles the communication with the base station
	BaseStationConnection           	*net.UDPConn    // connection that handles the communication with the base station
	nodeID                          	int8            // name of the gossiper
	Peers                           	map[string]Peer // set of known peers
	RemainingPeers                  	*SafeMapPeers   // set of peers from which the host needs the prediction (for the current round)
	Leader                              Peer
	CurrentStatus                       *StatusConcurrent
	LocalDecision                       LocalOpinionVector // accumulator of the local predictions
	ReceivedLocalPredictions            int
	ExternalPredictions                 *SafeMapSinglePredictions
	PredictionsAggregatorHandler        chan struct{}
	EndRoundHandler                     chan struct{}
	FinalPredictionPropagationTerminate chan struct{}
	TimeoutHandler						chan struct{}
	DetectionClass 						int // the object that we are trying to detect
	ConfidenceThresholds 				map[int]float64
	Timeout 							Timeout
}

type Timeout struct {
	Timeout bool
	mux sync.Mutex
}

type SinglePrediction struct {
	Value []float64
}

type Peer struct {
	PeerAddress    *net.UDPAddr
}

func NewNode(address, baseStationAddress string, id int8, peers string, detectionClass int) *Node {
	udpAddress, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		panic (fmt.Sprintf("Address not valid: %s", err))
	}

	udpConnection, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		panic (fmt.Sprintf("Error in opening UDP listener: %s", err))
	}

	udpBaseStationAddress, err := net.ResolveUDPAddr("udp4", baseStationAddress)
	if err != nil {
		panic (fmt.Sprintf("Address not valid: %s", err))
	}


	udpBaseStationLocalListenerAddress, err := net.ResolveUDPAddr("udp4", udpBaseStationAddress.IP.String() + string(int8(udpBaseStationAddress.Port) + id))
	udpBaseStationConnection, err := net.ListenUDP("udp4", udpBaseStationLocalListenerAddress)
	if err != nil {
		panic (fmt.Sprintf("Error in opening UDP listener: %s", err))
	}

	err = udpBaseStationConnection.SetReadDeadline(time.Now().Add(2*time.Second))
	if err != nil {
		panic(fmt.Sprintf("Error in setting the deadline: %s", err))
	}


	myPeers := make(map[string]Peer)
	node := &Node{
		Address:               udpAddress,
		Connection:            udpConnection,

		BaseStationAddress:    udpBaseStationAddress,

		BaseStationLocalListenerAddress: udpBaseStationLocalListenerAddress,
		BaseStationConnection:           udpBaseStationConnection,
		nodeID:                          id,
		Peers:                           myPeers,
		CurrentStatus:
		&StatusConcurrent{
			StatusValue: Status{
				CurrentRound:      0,
				CurrentState:      WaitingForLocalPredictions,
				CurrentPrediction: SinglePrediction{ Value: []float64{0, 0, 0}},
			},
		},
		LocalDecision: LocalOpinionVector{
			alpha: 0.5,
			scores: [DetectionClasses]float64{},
			boundingBoxCoefficients: [DetectionClasses]float64{},
		},

		ReceivedLocalPredictions: 0,
		DetectionClass: detectionClass,
		ConfidenceThresholds : map[int]float64{
			//"person": 0.6,
			//"bottle": 0.12,
			1: 0.6,
			44: 0.12,
			72: 0.4,
		},

		//Timeout: false,
		Timeout: Timeout{
			Timeout: false,
		},
	}


	for _, peer := range strings.Split(peers, ",") {
		myPeers := node.Peers

		peerAddress, err := net.ResolveUDPAddr("udp4", peer)
		if err != nil {
			panic(fmt.Sprintf("Error in parsing the UDP peerAddress: %s", err))
		}

		peerWrapper := Peer {
			PeerAddress: peerAddress,
		}

		myPeers[peerAddress.String()] = peerWrapper
		node.Peers = myPeers
	}

	node.RemainingPeers = node.InitPeersMap()
	node.ExternalPredictions = node.InitExternalPredictionsMap()

	return node
}

func (node *Node) isLeader() bool {
	//return node.Address.String() == node.Leader.PeerAddress.String()
	//fmt.Println("leader? nodeID: ", uint64(node.nodeID), ", modulo: ", (node.CurrentStatus.StatusValue.CurrentRound % uint64(len(node.Peers) + 1)), ", currentRound: ", node.CurrentStatus.StatusValue.CurrentRound, "len+1: ", uint64(len(node.Peers) + 1))
	return (uint64(node.nodeID) % uint64(len(node.Peers) + 1)) == (node.CurrentStatus.StatusValue.CurrentRound % uint64(len(node.Peers) + 1))
}

func (status *Status) String() string {
	return fmt.Sprintf("Current round: %d, state of the round: %s, current prediction: %s", status.CurrentRound, status.CurrentState, status.CurrentPrediction)
}