package main

import (
	"fmt"

	"client/pkg/packet"
)

func main() {
	nonsocket()
}

func nonsocket() {
	srcPort := 25566

	ipAddress := "127.0.0.1"
	port := 8080
	httpReqHeaders := []byte(fmt.Sprintf(`POST / HTTP/1.1\r\nHost: %s:%d\r\nContent-Type: application/json\r\nContent-Length: 18\r\n\r\n`, ipAddress, port))
	httpPayload := []byte(`{"hello": "world"}`)
	data := append(httpReqHeaders, httpPayload...)

	packet.SendPacketNoSocket(ipAddress, uint16(port), uint16(srcPort), data)
}
