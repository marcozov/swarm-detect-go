package main

import (
	"sync"
	"net"
)

type SafeMapPeers struct {
	v   map[string]Peer
	mux sync.Mutex
}

type SafeMapSinglePredictions struct {
	//v 	map[string]SinglePredictionWithSender
	v 	map[string]SinglePrediction
	mux sync.Mutex
}

func (m *SafeMapPeers) Length() int {
	m.mux.Lock()
	defer m.mux.Unlock()

	return len(m.v)
}

func (m *SafeMapSinglePredictions) Length() int {
	m.mux.Lock()
	defer m.mux.Unlock()

	return len(m.v)
}



// helper functions for the SafeMapPeers type
func (m *SafeMapPeers) removePeer(peer net.UDPAddr) {
	m.mux.Lock()
	defer m.mux.Unlock()

	delete(m.v, peer.String())
}

func (m *SafeMapSinglePredictions) removePrediction(prediction string) {
	m.mux.Lock()
	defer m.mux.Unlock()

	delete(m.v, prediction)
}



func (m *SafeMapPeers) addPeer(peer Peer) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.v[peer.peerAddress.String()] = peer
}

func (m *SafeMapSinglePredictions) addPrediction(host string, prediction SinglePrediction) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.v[host] = prediction
}

func (m *SafeMapPeers) getPeer(peer net.UDPAddr) *Peer {
	m.mux.Lock()
	defer m.mux.Unlock()

	if val, ok := m.v[peer.String()]; ok {
		return &val
	}

	return nil
}

func (m *SafeMapSinglePredictions) getPrediction(host string) *SinglePrediction {
	m.mux.Lock()
	defer m.mux.Unlock()

	if val, ok := m.v[host]; ok {
		return &val
	}

	return nil
}