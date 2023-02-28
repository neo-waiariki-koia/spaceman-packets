package packet

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
)

func (ip *IPPacket) createAckPacket(respSeq int) []byte {
	ack := respSeq + 1
	seq := ip.TCPHeader.SequenceNumber + 1
	flags := []string{"ack"}
	data := make([]byte, 0)
	packet := ip.buildTCPPacket(uint32(seq), uint32(ack), flags, data)
	return packet
}

func (ip *IPPacket) createDataPacket(respSeq int, data []byte) []byte {
	ack := respSeq
	seq := ip.TCPHeader.SequenceNumber
	flags := []string{"psh", "ack"}
	packet := ip.buildTCPPacket(uint32(seq), uint32(ack), flags, data)
	return packet
}

func SendPacketSocket(remoteHost string, destPort, srcPort uint16, data []byte) {
	laddr := "0.0.0.0"
	ipPacket := NewIPPacket(remoteHost, destPort, laddr, srcPort)

	fd, err := initialiseSocket(laddr, int(srcPort))
	defer syscall.Close(fd)
	if err != nil {
		log.Fatal(err)
	}

	synPacket := ipPacket.createSynPacket()
	fmt.Println("SENDING Syn")
	err = sendPacket(fd, remoteHost, int(destPort), synPacket)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("RECEIVING Syn")
	respSeq, _, err := receivePacket(fd, remoteHost, int(destPort))
	if err != nil {
		log.Fatal(err)
	}

	ackPacket := ipPacket.createAckPacket(respSeq)
	fmt.Println("SENDING Ack")
	err = sendPacket(fd, remoteHost, int(destPort), ackPacket)
	if err != nil {
		log.Fatal(err)
	}

	pshPacket := ipPacket.createDataPacket(respSeq, data)
	fmt.Println("SENDING PUSH")
	err = sendPacket(fd, remoteHost, int(destPort), pshPacket)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("RECEIVING PUSH")
	_, _, err = receivePacket(fd, remoteHost, int(destPort))
	if err != nil {
		log.Fatal(err)
	}
}

func initialiseSocket(sourceAddress string, sourcePort int) (int, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		return 0, fmt.Errorf("initialiseSocket -> syscall.Socket: %s", err)
	}

	localAddr := syscall.SockaddrInet4{
		Port: sourcePort,
		Addr: to4byte(sourceAddress),
	}
	err = syscall.Bind(fd, &localAddr)
	if err != nil {
		return 0, fmt.Errorf("initialiseSocket -> syscall.Bind: %s", err)
	}
	return fd, nil
}

func sendPacket(fd int, destinationAddress string, destinationPort int, packet []byte) error {
	addr := syscall.SockaddrInet4{
		Port: destinationPort,
		Addr: to4byte(destinationAddress),
	}

	fmt.Printf("% X\n", packet)

	err := syscall.Sendto(fd, packet, 0, &addr)
	if err != nil {
		return fmt.Errorf("sendPacket -> syscall.Sendto: %s", err)
	}

	return nil
}

func receivePacket(fd int, destinationAddress string, destinationPort int) (int, int, error) {
	var tcpHeader *TCPHeader

	for {
		buf := make([]byte, 1024)
		numRead, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			return 0, 0, fmt.Errorf("receivePacket -> syscall.Recvfrom: %s", err)
		}

		fmt.Printf("% X\n", buf[:numRead])

		ipData := buf[:]
		ipHeader := UnmarshalIP(ipData)
		if validPacket := ipHeader.validateTcpPacket(); !validPacket {
			return 0, 0, fmt.Errorf("receivePacket -> `Not a valid TCP packet")
		}

		tcpHeader = UnmarshalTCP(ipHeader.Data)

		sourceAddress := ipByteToString(ipHeader.SourceAddress)
		fmt.Printf("Source Address: %s\n", sourceAddress)
		fmt.Printf("Source Port: %v\n", tcpHeader.SourcePort)

		fmt.Printf("Destination Address: %s\n", ipByteToString(ipHeader.DestinationAddress))
		fmt.Printf("Destination Port: %v\n", tcpHeader.DestinationPort)
		fmt.Printf("Data: %v\n", tcpHeader.Data)

		fmt.Printf("%s vs %s\n", sourceAddress, destinationAddress)
		fmt.Printf("%d vs %d\n", int(tcpHeader.SourcePort), destinationPort)
		if validated := validateSource(sourceAddress, destinationAddress, int(tcpHeader.SourcePort), destinationPort); validated {
			break
		}
	}

	return int(tcpHeader.SequenceNumber), int(tcpHeader.AcknowlegementNumber), nil
}

func ipByteToString(ipLong uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	ip := net.IP(ipByte)
	return ip.String()
}

func validateSource(sourceAddress, destinationAddress string, sourcePort, destinationPort int) bool {
	if sourceAddress == destinationAddress {
		if sourcePort == destinationPort {
			return true
		}
	}
	return false
}
