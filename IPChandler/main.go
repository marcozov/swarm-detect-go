package main

import (
	"net"
	"os"
	"fmt"
	"flag"
	"strings"
	"github.com/marcozov/swarm-detect-go/structures"
	_ "net/http/pprof"
	"net/http"
	"log"
	"strconv"
)

func removeSocket(socketPath string) {
	err := os.Remove(socketPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error in removing unix socket: %s\n", err))
	}
}

// For now, the hosts in the network are assumed to be known: static configuration
func main() {



	fmt.Println("start: ", structures.WaitingForLocalPredictions)
	fmt.Println("finish: ", structures.LocalPredictionsTerminated)
	address := flag.String("address", "127.0.0.1:5000", "ip:port for the node")
	nodeID := flag.Int("nodeID", -1, "ID of the node")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	baseStation := flag.String("BS", "", "ip:port for the base station")
	baseStationListenerIP := flag.String("BSListener", "", "ip for the base station listener")

	detectionClass := flag.String("class", "person", "define the object to detect")
	classesMapping := map[string]int{
		"person": 1,
		"bottle": 44,
		"tv": 72,
		"laptop": 73,
		"mouse": 74,
		"keyboard": 76,
		"cellphone": 77,
		"book": 84,
	}

	flag.Parse()


	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%s", strconv.Itoa(6060 + *nodeID)), nil))
	}()

	completeSocketPath := "/tmp/go" + strings.Split(*address, ":")[1] + ".sock"

	// the UDP channel could be probably be opened at the beginning
	// One is going to be ok! It can be used for sending and receiving data
	fmt.Println(*address)
	fmt.Println(*nodeID)
	fmt.Println(*peers)

	node := structures.NewNode(*address, *baseStation, *baseStationListenerIP, int8(*nodeID), *peers, classesMapping[*detectionClass])

	localPredictionsChannel := make(chan []byte)


	timeoutHandler := make(chan struct{})
	packetHandler := make(chan structures.PacketChannelMessage)
	finalPredictionHandler := make(chan int8)
	startRoundHandler := make(chan uint64)

	node.TimeoutHandler = timeoutHandler
	node.PacketHandler = packetHandler
	node.FinalPredictionHandler = finalPredictionHandler
	node.StartRoundHandler = startRoundHandler


	// periodic probes to the followers
	go node.ProcessMessage(packetHandler)

	go node.PeriodicPeersProbe()

	// put together the external predictions and the local one, once everything is available

	// intercepting traffic from python process
	go node.HandleLocalPrediction(localPredictionsChannel)
	// getting a status message and saving the received external prediction
	go node.HandleIncomingMessages()

	go node.HandleFinalPredictions()
	go node.HandleStartRound()
	go node.TimeoutTrigger()

	if _, err := os.Stat(completeSocketPath); !os.IsNotExist(err) {
		removeSocket(completeSocketPath)
	}

	//node.StartNewRound()
	node.StartRound(1, true, true)
	fmt.Println("status: ", node.CurrentStatus.StatusValue)

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

			localPredictionsChannel <- buf[:n]
		}

		conn.Close()
	}

}
