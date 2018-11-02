package main

import "sync"

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

type StatusConcurrent struct {
	StatusValue Status
	mux sync.Mutex
}

type Status struct {
	CurrentRound uint64
	CurrentState State
	CurrentPrediction SinglePrediction
}

type Packet struct {
	Prediction *PredictionMessage
	Start      *StartMessage
	End        *EndRoundMessage
	Status     *Status
}