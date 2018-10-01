package main

import (
	//"log"
	"net"
	"os"
	//"os/signal"
	//"syscall"
	"fmt"
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

		fmt.Println("after server got")
	}
}

func main() {
	log.Println("Starting echo server")
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

func main() {
	l, err := net.ListenUnix("unix",  &net.UnixAddr{"/tmp/go.sock", "unix"})
	if err != nil {
		panic(err)
	}
	defer os.Remove("/tmp/go.sock")


	connectionCreation:

	conn, err := l.AcceptUnix()
	if err != nil {
		panic(err)
	}

	for {
		//conn, err := l.AcceptUnix()
		//if err != nil {
		//	panic(err)
		//}
		var buf [4096]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("Error in reading data:", err)
			conn.Close()
			goto connectionCreation
			//panic(err)

		}
		fmt.Printf("server got: %s\n(%d bytes)", string(buf[:n]), n);
		//conn.Close()
	}

	conn.Close()
}