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
	Start = iota+1
	Finish
)

type Node struct {
	Address    *net.UDPAddr // peerAddress on which peers send messages
	Connection *net.UDPConn // to receive data from peers and to send data
	Name       string       // name of the gossiper
	Peers      map[string]Peer
	RemainingPeers *SafeMap
	//AcknowledgedPeers *SafeMap
	CurrentStatus *Status
	StartHandler chan *Packet
	StatusHandler chan *Packet
}

type Peer struct {
	peerAddress    *net.UDPAddr
}

type SafeMap struct {
	v   map[string]Peer
	mux sync.Mutex
}

func (node *Node) InitPeersMap() *SafeMap {
	peers := &SafeMap{
		v: make(map[string]Peer),
	}

	for key, value := range node.Peers {
		peers.v[key] = value
	}

	return peers
}

func (node *Node) GetPeer(peer net.UDPAddr) *Peer {
	if val, ok := node.Peers[peer.String()]; ok {
		return &val
	}

	return nil
}

//func (node *Node) AddAcknowledgedPeer(peer Peer) {
//	node.AcknowledgedPeers.addPeer(peer)
//}
//
//func (node *Node) GetAcknowledgedPeer(peerAddress net.UDPAddr) *Peer {
//	return node.AcknowledgedPeers.getPeer(peerAddress)
//}

func (node *Node) RemoveRemainingPeer(peer net.UDPAddr) {
	node.RemainingPeers.removePeer(peer)
}

func (m *SafeMap) Length() int {
	m.mux.Lock()
	defer m.mux.Unlock()

	return len(m.v)
}

func (m *SafeMap) addPeer(peer Peer) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.v[peer.peerAddress.String()] = peer
}

func (m *SafeMap) removePeer(peer net.UDPAddr) {
	m.mux.Lock()
	defer m.mux.Unlock()

	delete(m.v, peer.String())
}

func (m *SafeMap) getPeer(peer net.UDPAddr) *Peer {
	m.mux.Lock()
	defer m.mux.Unlock()

	if val, ok := m.v[peer.String()]; ok {
		return &val
	}

	return nil
}

func NewNode(address, name, peers string) *Node {
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
		CurrentStatus: &Status{
			CurrentRound: 1,
			CurrentState: Start,
		},
	}

	for _, peer := range strings.Split(peers, ",") {
		node.addPeer(peer)
	}

	return node
}

func (node *Node) addPeer(peer string) {
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

func (node *Node) statusDebug() {
	for {
		fmt.Println("STATUS DEBUG: ", node.CurrentStatus)
		time.Sleep(5 * time.Second)
	}
}

func (status *Status) String() string {
	return fmt.Sprintf("Current round: %d, state of the round: %s", status.CurrentRound, status.CurrentState)
}