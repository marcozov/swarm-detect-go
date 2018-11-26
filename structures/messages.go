package structures

import (
	"sync"
	"net"
)

type ProbeMessage struct {
	RoundID uint64
}

type StatusConcurrent struct {
	StatusValue Status
	mux sync.Mutex
}

type Status struct {
	CurrentRound uint64 // also indicates who the leader is
	CurrentState State
	CurrentPrediction SinglePrediction
}

type AcknowledgementMessage struct {
	ID uint64
}

type FinalPredictionMessage struct {
	ID uint64
	Prediction *SinglePrediction
}

type Packet struct {
	FinalPrediction *FinalPredictionMessage
	Status          *Status
	Probe 			*ProbeMessage
	Ack 			*AcknowledgementMessage
}

type PacketWithSender struct {
	Packet 			*Packet
	SenderAddress	*net.UDPAddr
}