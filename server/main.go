package main

import (
	"fmt"
	"log"
	"time"

	"server/pkg/socket"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

func main() {
	laddr := "127.0.0.1"
	port := 8080
	socket, err := socket.NewNetSocket(laddr, port)
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()

	fmt.Printf("Serving on %s:%d\n", laddr, port)

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)

		}
		go func() {
			buffer := make([]byte, 1024)
			_, err := conn.Read(buffer)
			if err != nil {
				log.Fatal(err)
			}

			// write data to response
			time := time.Now().Format(time.ANSIC)
			responseStr := fmt.Sprintf("socket-server received message: %v. Received time: %v", string(buffer[:]), time)
			fmt.Println(responseStr)
			conn.Write([]byte(responseStr))

			// close conn
			conn.Close()
		}()
	}
}
