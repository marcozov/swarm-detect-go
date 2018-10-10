package main

import (
	"net"
	"os"
	"fmt"
	"flag"
	"strings"
	"time"
)

/*
func echoServer(c net.Conn) {
	for {
		fmt.Println("reading...")
		buf := make([]byte, 4096)
		nr, err := c.Read(buf)
		fmt.Println("READ! **********")
		if err != nil {
			fmt.Sprintf("Error: %s\n", err)
			return
		}

		data := buf[0:nr]
		println("Server got:", string(data))
		_, err = c.Write(data)
		if err != nil {
			fmt.Sprintf("Error: %s\n", err)
			log.Fatal("Writing client error: ", err)
		}

		fmt.Println("after IPChandler got")
	}
}

func main() {
	log.Println("Starting echo IPChandler")
	ln, err := net.Listen("unix", "/tmp/go.sock")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		os.Exit(0)
	}(ln, sigc)

	for {
		fmt.Println("main method")
		fd, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
			log.Fatal("Accept error: ", err)
		}

		go echoServer(fd)
	}
}
*/

//var socketPath = "/tmp/go.sock"

func removeSocket(socketPath string) {
	err := os.Remove(socketPath)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error in removing unix socket: %s\n", err))
	}
}

// For now, the hosts in the network are assumed to be known: static configuration
func main() {
	address := flag.String("address", "127.0.0.1:5000", "ip:port for the node")
	name := flag.String("name", "", "name of the node")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")

	flag.Parse()

	completeSocketPath := "/tmp/go" + strings.Split(*address, ":")[1] + ".sock"

	// the UDP channel could be probably be opened at the beginning
	// One is going to be ok! It can be used for sending and receiving data
	fmt.Println(*address)
	fmt.Println(*name)
	fmt.Println(*peers)

	node := NewNode(*address, *name, *peers)

	startMessagesChannel := make(chan *Packet)
	statusMessagesChannel := make(chan *Packet)

	node.StartHandler = startMessagesChannel
	node.StatusHandler = statusMessagesChannel
	node.RemainingPeers = node.InitPeersMap()

	go func() {
		time.Sleep(10*time.Second)
		node.CurrentStatus.CurrentState = Finish
	} ()

	go node.propagateStartMessage(startMessagesChannel)
	go node.propagateStatusMessage(statusMessagesChannel)

	//go node.statusDebug()
	go node.triggerPropagators()
	go node.handleIncomingMessages()

	if _, err := os.Stat(completeSocketPath); os.IsExist(err) {
		removeSocket(completeSocketPath)
	}

	l, err := net.ListenUnix("unix",  &net.UnixAddr{completeSocketPath, "unix"})
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
			var buf [4096]byte
			n, err := conn.Read(buf[:])
			if err != nil {
				fmt.Println("Error in reading data: %s\n", err)
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


			fmt.Printf("IPC handler got: %s\n(%d bytes)", string(buf[:n]), n);
		}

		conn.Close()
	}


}

/*
func main() {
	//connectionCreation:
	conn, err := net.ListenUnixgram("unixgram",  &net.UnixAddr{"/tmp/unixdomain", "unixgram"})
	if err != nil {
		panic(err)
	}
	defer os.Remove("/tmp/go.sock")


	//connectionCreation:
	//
	//conn, err := l.AcceptUnix()
	//if err != nil {
	//	panic(err)
	//}

	for {
		var buf [4096]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("Error in reading data:", err)
			//conn.Close()
			//goto connectionCreation
			//panic(err)
		}
		fmt.Printf("%s\n", string(buf[:n]));
	}

	conn.Close()
}*/