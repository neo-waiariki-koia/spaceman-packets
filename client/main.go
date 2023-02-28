package main

import (
	"client/pkg/packet"
)

func main() {
	nonsocket()
}

func nonsocket() {
	srcPort := 25566

	ipAddress := "127.0.0.1"
	port := 8080

	data := "GET HTTP/1.0\r\n" +
		"Host: 127.0.0.1" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"\r\n"

	for {
		packet.SendPacketSocket(ipAddress, uint16(port), uint16(srcPort), []byte(data))
	}
}
