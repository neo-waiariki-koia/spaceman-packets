package packet

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

func SendPacketSocket(remoteHost string, destPort, srcPort uint16) {
	laddr := "127.0.0.1"
	ipPacket := NewIPPacket(remoteHost, destPort, laddr, srcPort, []byte{})

	packet := ipPacket.createPacket()

	fmt.Println("SENDING")
	sendTime := socketSendSyn(remoteHost, destPort, packet)
	fmt.Println("RECEIVING")
	receiveTime := socketReceiveSynAck(laddr, int(srcPort))

	fmt.Println(receiveTime.Sub(sendTime))
}

func socketSendSyn(raddr string, port uint16, packet []byte) time.Time {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		log.Fatalf("Socket: %s\n", err)
	}

	addr := syscall.SockaddrInet4{
		Addr: to4byte(raddr),
		Port: int(port),
	}

	fmt.Printf("Sending to %s:%d\n", raddr, port)

	err = syscall.Sendto(fd, packet, 0, &addr)
	if err != nil {
		log.Fatal("Sendto:", err)
	}

	return time.Now()
}

func socketReceiveSynAck(localAddress string, localPort int) time.Time {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		log.Fatal("Socket: ", err)
	}

	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SetsockoptInt error :%s\ns", err.Error())
		log.Fatal(err)
	}

	addr := syscall.SockaddrInet4{
		Port: int(localPort),
		Addr: to4byte(localAddress),
	}

	err = syscall.Bind(fd, &addr)
	if err != nil {
		log.Fatal("Bind: ", err)
	}

	for {
		buf := make([]byte, 1024)
		numRead, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			log.Fatal("Recvfrom: ", err)
		}
		fmt.Printf("% X\n", buf[:numRead])

		ipData := buf[14:]
		ipHeader := UnmarshalIP(ipData)
		if validPacket := ipHeader.validateTcpPacket(); !validPacket {
			log.Fatal("Not a valid TCP Packet")
		}

		tcpHeader := UnmarshalTCP(ipHeader.Data)
		fmt.Printf("%v\n", tcpHeader)
	}

	return time.Now()
}
