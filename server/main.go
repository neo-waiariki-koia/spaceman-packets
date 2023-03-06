package main

import (
	"fmt"
	"io"
	"log"

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

			log.Print("Writing response")
			io.WriteString(conn, "HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/html; charset=utf-8\r\n"+
				"Content-Length: 20\r\n"+
				"\r\n"+
				"<h1>hello world</h1>\r\n")
			if err != nil {
				log.Fatal(err)
			}
			// close conn
			err = conn.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}
