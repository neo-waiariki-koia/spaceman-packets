package main

import (
	"fmt"

	"client/pkg/packet"
)

func main() {
	nonsocket()
}

func nonsocket() {
	ipAddress := "192.168.1.100"
	port := 8080
	httpReqHeaders := []byte(fmt.Sprintf(`GET / HTTP/1.1\r\nHost: %s:%d\r\nContent-Type: application/json\r\nContent-Length: 18\r\n\r\n`, ipAddress, port))
	httpPayload := []byte(`{"hello": "world"}`)
	data := append(httpReqHeaders, httpPayload...)

	packet.SendPacketNoSocket(ipAddress, uint16(port), data)
}
