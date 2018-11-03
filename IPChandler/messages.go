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

type StartMessage struct {
	RoundID uint64
}

type EndRoundMessage struct {
	RoundID uint64
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
	//FinalPrediction *PredictionMessage
	//FinalPrediction *SinglePrediction
	FinalPrediction *FinalPredictionMessage
	Start           *StartMessage
	End             *EndRoundMessage
	Status          *Status
	Probe 			*ProbeMessage
	Ack 			*AcknowledgementMessage
	SenderAddress	*net.UDPAddr
}