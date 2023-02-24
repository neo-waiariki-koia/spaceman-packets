package packet

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func (ip *IPPacket) createPacket() []byte {
	ack := 0
	seq := 0
	flags := []string{"syn"}
	data := make([]byte, 0)
	packet := ip.buildTCPPacket(uint32(seq), uint32(ack), flags, data)
	return packet
}

func (ip *IPPacket) buildTCPPacket(seq, ack uint32, flags []string, data []byte) []byte {
	tcpHeader := ip.TCPHeader.MarshalTCP(ip.sAddr, ip.dAddr, seq, ack, flags)

	return tcpHeader
}

func SendPacketNoSocket(remoteHost string, destPort, srcPort uint16, data []byte) {
	localAddr := interfaceAddress("en0")
	laddr := strings.Split(localAddr.String(), "/")[0] // Clean addresses like 192.168.1.30/24

	ipPacket := NewIPPacket(remoteHost, destPort, laddr, srcPort, data)

	packet := ipPacket.createPacket()

	var wg sync.WaitGroup
	wg.Add(1)
	var receiveTime time.Time

	addrs, err := net.LookupHost(remoteHost)
	if err != nil {
		log.Fatalf("Error resolving %s. %s\n", remoteHost, err)
	}
	remoteAddr := addrs[0]

	go func() {
		receiveTime = receiveSynAck(laddr, remoteAddr)
		wg.Done()
	}()

	time.Sleep(1 * time.Millisecond)

	sendTime := sendSyn(remoteAddr, destPort, packet)

	wg.Wait()

	fmt.Println(receiveTime.Sub(sendTime))
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
