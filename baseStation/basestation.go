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

/*

type Reply struct {
	Field int64
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		//netData, err := bufio.NewReader(c).ReadString('\n')
		tcpBuffer := make([]byte, 64)
		n, err := c.Read(tcpBuffer)
		if err != nil {
			panic(fmt.Sprintf("error in reading TCP data: %s.\ntcpBuffer: %v\nlocal address: %s\nn read bytes: %s", err, tcpBuffer, c.LocalAddr(), n))
		}

		tcpBuffer = tcpBuffer[:n]
		//fmt.Println("data from client: ", tcpBuffer)

		packet := &structures.Packet{}
		err = protobuf.Decode(tcpBuffer, packet)

		if err != nil {
			panic(fmt.Sprintf("error in decoding TCP data: %s\ntcpBuffer: %v\nlocalAddress: %s\npacket: %s\nn bytes: %d", err, tcpBuffer, c.LocalAddr(), packet, n))
		}
		fmt.Println("packet: ", packet)


		time.Sleep(1*time.Second)

		//result := &Reply{Field: 1232}
		//packetBytes, err := protobuf.Encode(result)
		//fmt.Println("bytes to send back: ", packetBytes)
		//c.Write(packetBytes)//result := strconv.Itoa(12321) + "\n"
		////c.Write([]byte(string(result)))
	}
	c.Close()
}

func main() {

	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
*/

func main() {
	address := flag.String("address", "192.168.0.100:3000", "ip:port for the node")
	realPrediction := flag.Int("realPrediction", 1, "usage: specify the real prediction for getting the TP, FP and FN")

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

		fmt.Println("This is the resulf ot the consensus (from", senderAddress.String(), "! ", receivedFinalPrediction.FinalPrediction.ID, receivedFinalPrediction.FinalPrediction.Prediction.Value[0])

		ack := &structures.Packet{
			Ack: &structures.AcknowledgementMessage{ID: receivedFinalPrediction.FinalPrediction.ID,},
		}

		//udpConnection.WriteToUDP()

		f.WriteString(fmt.Sprintf("%d,%f,%d,%d\n", receivedFinalPrediction.FinalPrediction.ID, receivedFinalPrediction.FinalPrediction.Prediction.Value[0], *realPrediction, totalTime.Nanoseconds() / 1000000))

		fmt.Println("sending ack back to ", senderAddress.String())
		structures.SendToPeer(ack, senderAddress, udpConnection)
	}
}