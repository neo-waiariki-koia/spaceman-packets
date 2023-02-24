package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

func main() {
	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	// close listener
	fmt.Printf("Serving on %s\n", listen.Addr().String())
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)

		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	// incoming request
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// write data to response
	time := time.Now().Format(time.ANSIC)
	responseStr := fmt.Sprintf("simple-server received message: %v. Received time: %v", string(buffer[:]), time)
	fmt.Println(responseStr)
	conn.Write([]byte(responseStr))

	// close conn
	conn.Close()
}
