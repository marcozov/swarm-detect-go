package main

import (
	"net"
	"os"
	"fmt"
	"flag"
	"strings"
	//"time"
)

func removeSocket(socketPath string) {
	err := os.Remove(socketPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error in removing unix socket: %s\n", err))
	}
}

// For now, the hosts in the network are assumed to be known: static configuration
func main() {
	fmt.Println("start: ", WaitingForLocalPredictions)
	fmt.Println("finish: ", LocalPredictionsTerminated)
	address := flag.String("address", "127.0.0.1:5000", "ip:port for the node")
	name := flag.String("name", "", "name of the node")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	leaderDummy := flag.String("leader", "", "ip:port for the leader")

	detectionClass := flag.String("class", "person", "define the object to detect")
	classesMapping := map[string]int{
		"person": 1,
	}

	fmt.Println("class: ", classesMapping[*detectionClass])
	flag.Parse()

	completeSocketPath := "/tmp/go" + strings.Split(*address, ":")[1] + ".sock"

	// the UDP channel could be probably be opened at the beginning
	// One is going to be ok! It can be used for sending and receiving data
	fmt.Println(*address)
	fmt.Println(*name)
	fmt.Println(*peers)

	node := NewNode(*address, *name, *peers, classesMapping[*detectionClass])

	// dummy leader init: assuming that it is fixed for now
	if leader, exists := node.Peers[*leaderDummy]; exists {
		node.Leader = leader
	} else if node.Address.String() == *leaderDummy {
		node.Leader = Peer{peerAddress: node.Address}
	}

	startMessagesChannel := make(chan *Packet)
	//statusMessagesChannel := make(chan *Packet)
	localPredictionsChannel := make(chan []byte)
	//externalPredictionsChannel := make(chan *PacketWithSender)
	endRoundMessagesChannel := make(chan *Packet)
	predictionsAggregatorHandler := make(chan struct{})
	endRoundHandler := make(chan struct{})
	finalPredictionPropagationTerminate := make(chan struct{})

	node.StartHandler = startMessagesChannel
	//node.StatusHandler = statusMessagesChannel
	//node.ExternalPredictionsHandler = externalPredictionsChannel
	node.EndRoundMessageHandler = endRoundMessagesChannel
	node.PredictionsAggregatorHandler = predictionsAggregatorHandler
	node.EndRoundHandler = endRoundHandler
	node.FinalPredictionPropagationTerminate = finalPredictionPropagationTerminate

	//go node.opinionVectorDEBUG()

	// periodic probes to the followers
	go node.periodicPeersProbe()
	// put together the external predictions and the local one, once everything is available
	go node.AggregateAllPredictions()

	// propagate my status in order to let the leader know my opinion
	//go node.propagateStatusMessage()

	// intercepting traffic from python process
	go node.HandleLocalPrediction(localPredictionsChannel)
	// getting a status message and saving the received external prediction
	//go node.updateExternalPredictions()
	go node.handleIncomingMessages()

	if _, err := os.Stat(completeSocketPath); os.IsExist(err) {
		removeSocket(completeSocketPath)
	}

	l, err := net.ListenUnix("unix", &net.UnixAddr{completeSocketPath, "unix"})
	if err != nil {
		panic(fmt.Sprintf("Error in opening UNIX socket listener: %s\n", err))
	}
	defer removeSocket(completeSocketPath)

	// retrieving data from the Python sockets: should this be run in parallel as a goroutine?
	for {
		conn, err := l.AcceptUnix()
		if err != nil {
			panic(fmt.Sprintf("Error in accepting UNIX socket connection: %s\n", err))
		}

		for {
			// any huge amount of data read at this point is coming from a socket..
			// no bandwidth is used
			var buf [4096]byte
			n, err := conn.Read(buf[:])
			if err != nil {
				fmt.Printf("Error in reading data: %s\n", err)
				conn.Close()
				//goto connectionCreation
				break
			}

			// data should now be interpreted from JSON and put in proper structure
			// 1) how often data will arrive?
			// 2) in what format? A single score/class/box or multiple values?
			// 2.1) if a single score/class/box is received, then it would be pretty hard to
			//		define some quality metrics in the system --> but lighter protocol
			// 2.2) otherwise, some kind of entropy can be determined
			// in any case, data should be propagated to other hosts in this phase..

			localPredictionsChannel <- buf[:n]
		}

		conn.Close()
	}

}