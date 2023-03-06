package main

import (
	"client/pkg/packet"
	"fmt"
)

func main() {
	destHost := "127.0.0.1"
	destPort := 8080

	srcHost := destHost
	srcPort := 25566

	data := "GET / HTTP/1.1\r\n" +
		"Host: 127.0.0.1:8080" +
		"User-Agent: go-client\r\n" +
		"Accept: */*\r\n" +
		"\r\n"

	response := packet.SendTCPData(destHost, destPort, srcHost, srcPort, []byte(data))
	fmt.Println(string(response))
}
