package main

import (
	"net"
)

func (node *Node) InitPeersMap() *SafeMapPeers {
	peers := &SafeMapPeers{
		v: make(map[string]Peer),
	}

	for key, value := range node.Peers {
		if key != node.Address.String() {
			peers.v[key] = value
		}
	}

	return peers
}

// TODO: check whether it is ok not to put concurrency here
func (node *Node) GetPeer(peer net.UDPAddr) *Peer {
	if val, ok := node.Peers[peer.String()]; ok {
		return &val
	}

	return nil
}


func (node *Node) RemoveRemainingPeer(peer net.UDPAddr) {
	node.RemainingPeers.removePeer(peer)
}