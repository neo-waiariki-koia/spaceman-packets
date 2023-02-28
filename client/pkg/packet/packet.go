package packet

import (
	"log"
	"net"
)

type IPPacket struct {
	IPHeader
	TCPHeader
}

func NewIPPacket(destAddr string, destPort uint16, sourceAddr string, sourcePort uint16) *IPPacket {
	ipHeader := IPHeader{
		sAddr: sourceAddr,
		dAddr: destAddr,
	}

	tcpHeader := TCPHeader{
		SourcePort:      sourcePort,
		DestinationPort: destPort,
	}

	return &IPPacket{
		IPHeader:  ipHeader,
		TCPHeader: tcpHeader,
	}
}

func calculateChecksum(data []byte) uint16 {
	dataLength := len(data)
	var nextWord uint16
	var sum uint32

	for i := 0; i+1 < dataLength; i += 2 {
		nextWord = uint16(data[i])<<8 | uint16(data[i+1])
		sum += uint32(nextWord)
	}

	if dataLength%2 != 0 {
		//fmt.Println("Odd byte")
		sum += uint32(data[dataLength-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)

	// Bitwise complement
	return uint16(^sum)
}

func interfaceAddress(ifaceName string) net.Addr {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("net.InterfaceByName for %s. %s", ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Fatalf("iface.Addrs: %s", err)
	}

	return addrs[1]
}
