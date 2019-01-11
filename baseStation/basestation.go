package main

import (
	"fmt"
	"net"
	"github.com/marcozov/swarm-detect-go/structures"
	"github.com/dedis/protobuf"
	"flag"
	"os"
	"time"
)


func main() {
	address := flag.String("address", "192.168.0.100:3000", "ip:port for the node")

	flag.Parse()

	udpAddress, err := net.ResolveUDPAddr("udp4", *address)
	if err != nil {
		panic (fmt.Sprintf("Address not valid: %s", err))
	}

	udpConnection, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		panic (fmt.Sprintf("Error in opening UDP listener: %s", err))
	}

	fmt.Println("listening on ", udpAddress.String())


	f, err := os.OpenFile("experiments.txt", os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	f.WriteString("RoundID,Predicted Value,Real Value,Taken Time\n")

	for {
		// may need to be expanded to support bigger messages..
		udpBuffer := make([]byte, 64)

		start := time.Now()
		n, senderAddress, err := udpConnection.ReadFromUDP(udpBuffer)
		totalTime := time.Since(start)

		if err != nil {
			panic(fmt.Sprintf("error in reading UDP data: %s.\nudpBuffer: %v\nsenderAddress: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), n))
		}

		udpBuffer = udpBuffer[:n]

		receivedFinalPrediction := &structures.Packet{}
		err = protobuf.Decode(udpBuffer, receivedFinalPrediction)

		if err != nil {
			panic(fmt.Sprintf("error in decoding UDP data: %s\nudpBuffer: %v\nsenderAddress: %s\nreceivedFinalPrediction: %s\nn bytes: %d", err, udpBuffer, senderAddress.String(), receivedFinalPrediction, n))
		}

		//go node.processMessage(receivedFinalPrediction, senderAddress, counter)

		if receivedFinalPrediction.FinalPrediction == nil {
			panic(fmt.Sprintf("The base station can receive only final prediction messages: %s", receivedFinalPrediction))
		}

		fmt.Println("This is the result ot the consensus (from", senderAddress.String(), "! ", receivedFinalPrediction.FinalPrediction.ID, receivedFinalPrediction.FinalPrediction.Prediction.Value[0])

		ack := &structures.Packet{
			Ack: &structures.AcknowledgementMessage{ID: receivedFinalPrediction.FinalPrediction.ID,},
		}

		//udpConnection.WriteToUDP()

		f.WriteString(fmt.Sprintf("%d,%f,%d\n", receivedFinalPrediction.FinalPrediction.ID, receivedFinalPrediction.FinalPrediction.Prediction.Value[0], totalTime.Nanoseconds() / 1000000))

		fmt.Println("sending ack back to ", senderAddress.String())
		structures.SendToPeer(ack, senderAddress, udpConnection)
	}
}
