//go:build linux && amd64
// +build linux,amd64

package packet

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
)

type IPHeader struct {
	Version            uint8
	Protocol           uint8
	SourceAddress      uint32
	sAddr              string
	DestinationAddress uint32
	dAddr              string
	Data               []byte
}

type TCPHeader struct {
	SourcePort           uint16
	DestinationPort      uint16
	SequenceNumber       uint32
	AcknowlegementNumber uint32
	DataOffset           uint8
	Flags                uint8
	Window               uint16
	Checksum             uint16 // Kernel will set this if it's 0
	Urgent               uint16
	Data                 []byte
}

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
	fmt.Println("Starting handshake")
	ack := 0
	sequence := 0
	flags := []string{"syn"}
	emptyData := make([]byte, 0)

	fmt.Println("Building syn packet")
	packet := BuildTCPPacket(pe.DestHost, pe.DestPort, pe.SrcHost, pe.SrcPort, sequence, ack, flags, pe.Checksum, emptyData)
	err := pe.sendPacket(packet)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Receiving response")
	responseSequence, _, err := pe.receiveResponse()
	if err != nil {
		log.Fatal(err)
	}

	ack = responseSequence + 1
	sequence += 1
	flags = []string{"ack"}
	fmt.Println("Building ack packet")
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
	err = syscall.SetsockoptInt(socket, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		return fmt.Errorf("sendPacket -> syscall.SetsockoptInt: %s", err)
	}

	addr := &syscall.SockaddrInet4{
		Port: pe.DestPort,
		Addr: to4byte(pe.DestHost),
	}

	fmt.Printf("Sending to %v:%v\n", addr.Addr, addr.Port)
	fmt.Printf("% X\n", packet)
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
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, syscall.ETH_P_ALL)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Socket: %s", err)
	}

	err = syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 56789)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.SetsockoptInt: %s", err)
	}

	// addr := &syscall.SockaddrInet4{
	// 	Port: pe.SrcPort,
	// 	Addr: to4byte(pe.SrcHost),
	// }
	// err = syscall.Bind(socket, addr)
	// if err != nil {
	// 	return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Bind: %s", err)
	// }

	var tcpHeader *TCPHeader
	for {
		buf := make([]byte, 1024)
		numRead, _, err := syscall.Recvfrom(socket, buf, 0)
		if err != nil {
			return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Recvfrom: %s", err)
		}

		fmt.Printf("% X\n", buf[numRead:])

		ipData := buf[:]
		ipHeader := unmarshalIPHeader(ipData)
		if validPacket := ipHeader.validateTcpPacket(); !validPacket {
			return 0, []byte{}, fmt.Errorf("receiveResponse -> `Not a valid TCP packet")
		}

		tcpHeader = unmarshalTCPHeader(ipHeader.Data)

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

func unmarshalIPHeader(data []byte) *IPHeader {
	var ipHeader IPHeader

	firstSection := netToHostShort(binary.BigEndian.Uint16(data[0:2]))
	ipHeader.Version = byte(firstSection >> 12) // top 4 bits

	thirdSection := netToHostShort(binary.BigEndian.Uint16(data[8:10]))
	_ = byte(thirdSection >> 8)                 // top 8 bits
	ipHeader.Protocol = byte(thirdSection >> 0) // second 8 bits

	ipHeader.SourceAddress = netToHostLong(binary.BigEndian.Uint32(data[12:16]))
	ipHeader.DestinationAddress = netToHostLong(binary.BigEndian.Uint32(data[16:20]))

	ipHeader.Data = data[20:]

	ipHeader.sAddr = ipByteToString(ipHeader.SourceAddress)
	ipHeader.dAddr = ipByteToString(ipHeader.DestinationAddress)

	return &ipHeader
}

func (ih *IPHeader) validateTcpPacket() bool {
	return ih.Protocol == 6
}

func ipByteToString(ipLong uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	ip := net.IP(ipByte)
	return ip.String()
}

func unmarshalTCPHeader(data []byte) *TCPHeader {
	var tcp TCPHeader

	tcp.SourcePort = netToHostShort(binary.BigEndian.Uint16(data[0:2]))
	tcp.DestinationPort = netToHostShort(binary.BigEndian.Uint16(data[2:4]))
	tcp.SequenceNumber = netToHostLong(binary.BigEndian.Uint32(data[4:8]))
	tcp.AcknowlegementNumber = netToHostLong(binary.BigEndian.Uint32(data[8:12]))
	tcp.DataOffset = data[12] >> 4
	tcp.Flags = data[13]
	tcp.Checksum = netToHostShort(binary.BigEndian.Uint16(data[16:18]))

	dataStart := int(tcp.DataOffset) * 4
	tcp.Data = data[dataStart:]

	return &tcp
}

func SendTCPData(destHost string, destPort int, srcHost string, srcPort int, data []byte) []byte {
	checkSum := 0
	send := NewPacketExchange(srcHost, srcPort, destHost, destPort, checkSum, data)

	seq, ack := send.tcpHandshake()
	response := send.tcpPushData(seq, ack)

	return response
}

// NetToHostShort converts a 16-bit integer from network to host byte order, aka "ntohs"
func netToHostShort(i uint16) uint16 {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, i)
	return binary.LittleEndian.Uint16(data)
}

// NetToHostLong converts a 32-bit integer from network to host byte order, aka "ntohl"
func netToHostLong(i uint32) uint32 {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, i)
	return binary.LittleEndian.Uint32(data)
}
