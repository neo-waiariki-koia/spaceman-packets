package packet

import (
	"fmt"
	"log"
	"net"
	"time"
)

func (ip *IPPacket) createSynPacket() []byte {
	ack := 0
	seq := 0
	flags := []string{"syn"}
	data := make([]byte, 0)
	packet := ip.buildTCPPacket(uint32(seq), uint32(ack), flags, data)
	return packet
}

func (ip *IPPacket) buildTCPPacket(seq, ack uint32, flags []string, data []byte) []byte {
	// ipHeader := ip.IPHeader.MarshalIP()
	tcpHeader := ip.TCPHeader.MarshalTCP(ip.sAddr, ip.dAddr, seq, ack, flags)

	// packet := append(ipHeader, tcpHeader...)
	// return append(packet, data...)
	return tcpHeader
}

func sendSyn(raddr string, port uint16, packet []byte) time.Time {
	fmt.Printf("Sending to %s:%d\n", raddr, port)
	conn, err := net.Dial("ip4:tcp", raddr)
	if err != nil {
		log.Fatalf("Dial: %s\n", err)
	}

	sendTime := time.Now()

	numWrote, err := conn.Write(packet)
	if err != nil {
		log.Fatalf("Write: %s\n", err)
	}
	if numWrote != len(packet) {
		log.Fatalf("Short write. Wrote %d/%d bytes\n", numWrote, len(packet))
	}

	err = conn.Close()
	if err != nil {
		log.Fatalf("Close: %s\n", err)
	}

	return sendTime
}

func receiveSynAck(localAddress, remoteAddress string) time.Time {
	netaddr, err := net.ResolveIPAddr("ip4", localAddress)
	if err != nil {
		log.Fatalf("net.ResolveIPAddr: %s. %s\n", localAddress, netaddr)
	}

	conn, err := net.ListenIP("ip4:tcp", netaddr)
	if err != nil {
		log.Fatalf("ListenIP: %s\n", err)
	}

	fmt.Println("Ready to receive")
	var receiveTime time.Time
	for {
		buf := make([]byte, 1024)
		_, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatalf("ReadFrom: %s\n", err)
		}

		if raddr.String() != remoteAddress {
			fmt.Println("Remote addresses are not equal")
			// this is not the packet we are looking for
			continue
		}

		ipData := buf[14:]
		ipHeader := UnmarshalIP(ipData)
		fmt.Printf("%v\n", ipHeader)
		if validPacket := ipHeader.validateTcpPacket(); !validPacket {
			log.Fatal("Not a valid TCP Packet")
		}

		tcpHeader := UnmarshalTCP(ipHeader.Data)
		fmt.Printf("%v\n", tcpHeader)
		break
	}

	return receiveTime
}
