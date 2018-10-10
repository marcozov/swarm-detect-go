package main

type BoundingBox struct {
	// TODO: define the structure
}

// more than ust a single prediction... right?
type SimpleMessage struct {
	RoundID int
	DetectionClass uint64
	DetectionScore float32
	DetectionBox BoundingBox
}

type StartMessage struct {
	RoundID int
}

type EndMessage struct {
	RoundID int
}

type StatusMessage struct {
	RoundID int
}

type Status struct {
	CurrentRound int
	CurrentState State
}

type Packet struct {
	Message *SimpleMessage
	Start *StartMessage
	End *EndMessage
	Status *Status
}