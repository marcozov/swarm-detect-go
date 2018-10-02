package main

type BoundingBox struct {
	// TODO: define the structure
}

type SimpleMessage struct {
	RoundID int
	DetectionClass uint64
	DetectionScore float32
	DetectionBox BoundingBox
}

type Packet struct {
	Message *SimpleMessage
}