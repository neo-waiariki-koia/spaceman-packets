package packet

import (
	"fmt"
	"log"
	"syscall"
)

type PacketExchange struct {
	SrcHost  string
	SrcPort  int
	DestHost string
	DestPort int
	Checksum int
	Data     []byte
}

func NewPacketExchange(srcHost string, srcPort int, destHost string, destPort int, checksum int, data []byte) *PacketExchange {
	return &PacketExchange{
		SrcHost:  srcHost,
		SrcPort:  srcPort,
		DestHost: destHost,
		DestPort: destPort,
		Checksum: checksum,
		Data:     data,
	}
}

func (pe *PacketExchange) tcpHandshake() (int, int) {
	ack := 0
	sequence := 0
	flags := []string{"syn"}
	emptyData := make([]byte, 0)

	packet := BuildTCPPacket(pe.DestHost, pe.DestPort, pe.SrcHost, pe.SrcPort, sequence, ack, flags, pe.Checksum, emptyData)
	err := pe.sendPacket(packet)
	if err != nil {
		log.Fatal(err)
	}
	responseSequence, _, err := pe.receiveResponse()
	if err != nil {
		log.Fatal(err)
	}

	ack = responseSequence + 1
	sequence += 1
	flags = []string{"ack"}
	packet = BuildTCPPacket(pe.DestHost, pe.DestPort, pe.SrcHost, pe.SrcPort, sequence, ack, flags, pe.Checksum, emptyData)
	err = pe.sendPacket(packet)
	if err != nil {
		log.Fatal(err)
	}

	return sequence, responseSequence
}

func (pe *PacketExchange) tcpPushData(sequence, ack int) []byte {
	flags := []string{"psh", "ack"}
	packet := BuildTCPPacket(pe.DestHost, pe.DestPort, pe.SrcHost, pe.SrcPort, sequence, ack, flags, pe.Checksum, pe.Data)
	err := pe.sendPacket(packet)
	if err != nil {
		log.Fatal(err)
	}
	_, response, err := pe.receiveResponse()
	if err != nil {
		log.Fatal(err)
	}

	return response
}

func (pe *PacketExchange) sendPacket(packet []byte) error {
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		return fmt.Errorf("sendPacket -> syscall.Socket: %s", err)
	}
	err = syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 1)
	if err != nil {
		return fmt.Errorf("sendPacket -> syscall.SetsockoptInt: %s", err)
	}

	addr := &syscall.SockaddrInet4{
		Port: pe.DestPort,
		Addr: to4byte(pe.DestHost),
	}
	err = syscall.Sendto(socket, packet, 0, addr)
	if err != nil {
		return fmt.Errorf("sendPacket -> syscall.Sendto: %s", err)
	}

	err = syscall.Close(socket)
	if err != nil {
		return fmt.Errorf("sendPacket -> syscall.Close: %s", err)
	}

	return nil
}

func (pe *PacketExchange) receiveResponse() (int, []byte, error) {
	socket, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Socket: %s", err)
	}

	err = syscall.SetsockoptInt(socket, syscall.IPPROTO_IP, syscall.SO_RCVBUF, 56789)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.SetsockoptInt: %s", err)
	}

	addr := &syscall.SockaddrInet4{
		Port: pe.SrcPort,
		Addr: to4byte(pe.SrcHost),
	}
	err = syscall.Bind(socket, addr)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Bind: %s", err)
	}

	var tcpHeader *TCPHeader
	for {
		buf := make([]byte, 1024)
		numRead, _, err := syscall.Recvfrom(socket, buf, 0)
		if err != nil {
			return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Recvfrom: %s", err)
		}

		fmt.Printf("% X\n", buf[numRead:])

		ipData := buf[:]
		ipHeader := UnmarshalIP(ipData)
		if validPacket := ipHeader.validateTcpPacket(); !validPacket {
			return 0, []byte{}, fmt.Errorf("receiveResponse -> `Not a valid TCP packet")
		}

		tcpHeader = UnmarshalTCP(ipHeader.Data)

		sourceAddress := ipByteToString(ipHeader.SourceAddress)
		fmt.Printf("Source Address: %s\n", ipHeader.sAddr)
		fmt.Printf("Source Port: %v\n", tcpHeader.SourcePort)

		fmt.Printf("Destination Address: %s\n", ipHeader.dAddr)
		fmt.Printf("Destination Port: %v\n", tcpHeader.DestinationPort)
		fmt.Printf("Data: %v\n", tcpHeader.Data)

		if validated := pe.validateSource(sourceAddress, int(tcpHeader.DestinationPort)); validated {
			break
		}
	}

	return int(tcpHeader.SequenceNumber), tcpHeader.Data, nil
}

func (pe *PacketExchange) validateSource(sourceAddress string, destinationPort int) bool {
	if sourceAddress == pe.DestHost {
		if pe.SrcPort == destinationPort {
			return true
		}
	}
	return false
}

func SendTCPData(destHost string, destPort int, srcHost string, srcPort int, data []byte) []byte {
	checkSum := 0
	send := NewPacketExchange(srcHost, srcPort, destHost, destPort, checkSum, data)

	seq, ack := send.tcpHandshake()
	response := send.tcpPushData(seq, ack)

	return response
}
