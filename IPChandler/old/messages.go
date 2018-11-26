package main

import (
	"sync"
	"net"
)

type BoundingBox struct {
	Coefficient float32
}

// there is probably no need to send information about other classes
type PredictionMessage struct {
	RoundID uint64
	DetectionClass uint64
	DetectionScore float64
	DetectionBox BoundingBox
	Entropy float32
}

type ProbeMessage struct {
	RoundID uint64
}

type StatusConcurrent struct {
	StatusValue Status
	mux sync.Mutex
}

type Status struct {
	CurrentRound uint64
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