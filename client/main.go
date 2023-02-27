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

	packet.SendPacketSocket(ipAddress, uint16(port), uint16(srcPort))
}
