package main

type BoundingBox struct {
	// TODO: define the structure
}

// more than ust a single prediction... right?
type PredictionMessage struct {
	RoundID uint64
	DetectionClass uint64
	DetectionScore float32
	DetectionBox BoundingBox
}

type StartMessage struct {
	RoundID uint64
}

type EndMessage struct {
	RoundID uint64
}

type StatusMessage struct {
	RoundID uint64
}

type Status struct {
	CurrentRound uint64
	CurrentState State
}

type Packet struct {
	Message *PredictionMessage
	Start *StartMessage
	End *EndMessage
	Status *Status
}