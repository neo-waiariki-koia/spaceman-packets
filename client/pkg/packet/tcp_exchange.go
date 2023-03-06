package packet

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"
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

func SendTCPData(destHost string, destPort int, srcHost string, srcPort int, data []byte) []byte {
	checkSum := 0
	send := NewPacketExchange(srcHost, srcPort, destHost, destPort, checkSum, data)

	for i := 0; i < 10; i++ {
		send.tcpHandshake()
		time.Sleep(6 * time.Second)
	}

	return []byte{}
}

func (pe *PacketExchange) tcpHandshake() (int, int) {
	flags := []string{"syn"}
	emptyData := make([]byte, 0)

	packet := BuildTCPPacket(pe.DestHost, pe.DestPort, pe.SrcHost, pe.SrcPort, 0, 0, flags, 0, emptyData)

	if err := pe.sendPacket(packet); err != nil {
		log.Printf("error sending packet: %v", err)
	}

	if err := pe.receive(); err != nil {
		log.Printf("error receiving: %v", err)
	}

	return 0, 0
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

func (pe *PacketExchange) receive() error {
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, syscall.ETH_P_ALL)
	if err != nil {
		return fmt.Errorf("receive -> syscall.Socket: %s", err)
	}

	// if err := syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 1073741824); err != nil {
	// 	return fmt.Errorf("receive -> syscall.SetsockoptInt: %s", err)
	// }

	netIFace, err := net.InterfaceByName("eth0")
	if err != nil {
		return fmt.Errorf("receive -> net.InterfaceByName: %s", err)
	}

	var haddr [8]byte
	copy(haddr[0:7], netIFace.HardwareAddr[0:7])

	addr := syscall.SockaddrLinklayer{
		Protocol: syscall.ETH_P_IP,
		Ifindex:  netIFace.Index,
		Halen:    uint8(len(netIFace.HardwareAddr)),
		Addr:     haddr,
	}

	// probably need to bind to device here instead
	if err := syscall.Bind(socket, &addr); err != nil {
		return fmt.Errorf("receive -> syscall.Bind: %s", err)
	}

	// if err := syscall.SetLsfPromisc("eth0", true); err != nil {
	// 	return fmt.Errorf("receive -> ")
	// }

	return nil
}

// func (pe *PacketExchange) receiveResponse() (int, []byte, error) {
// 	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, syscall.ETH_P_ALL)
// 	if err != nil {
// 		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Socket: %s", err)
// 	}

// 	err = syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 56789)
// 	if err != nil {
// 		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.SetsockoptInt: %s", err)
// 	}

// 	addr := &syscall.SockaddrInet4{
// 		Port: pe.SrcPort,
// 		Addr: to4byte(pe.SrcHost),
// 	}
// 	err = syscall.Bind(socket, addr)
// 	if err != nil {
// 		return 0, []byte{}, fmt.Errorf("receiveResponse -> syscall.Bind: %s", err)
// 	}

// 	var tcpHeader *TCPHeader
// 	seqDataChan := make(chan map[string]struct {
// 		SeqNum int
// 		Data   []byte
// 	}, 1)

// 	go func() {
// 		for {
// 			buf := make([]byte, 1024)
// 			numRead, _, err := syscall.Recvfrom(socket, buf, 0)
// 			if err != nil {
// 				log.Fatalf("receiveResponse -> syscall.Recvfrom: %s", err)
// 			}

// 			fmt.Printf("% X\n", buf[numRead:])

// 			ipData := buf[:]
// 			ipHeader := unmarshalIPHeader(ipData)
// 			if validPacket := ipHeader.validateTcpPacket(); !validPacket {
// 				log.Fatal("receiveResponse -> `Not a valid TCP packet")
// 			}

// 			tcpHeader = unmarshalTCPHeader(ipHeader.Data)

// 			sourceAddress := ipByteToString(ipHeader.SourceAddress)
// 			fmt.Printf("Source Address: %s\n", ipHeader.sAddr)
// 			fmt.Printf("Source Port: %v\n", tcpHeader.SourcePort)

// 			fmt.Printf("Destination Address: %s\n", ipHeader.dAddr)
// 			fmt.Printf("Destination Port: %v\n", tcpHeader.DestinationPort)
// 			fmt.Printf("Data: %v\n", tcpHeader.Data)

// 			if validated := pe.validateSource(sourceAddress, int(tcpHeader.DestinationPort)); validated {
// 				seqDataChan <- map[string]struct {
// 					SeqNum int
// 					Data   []byte
// 				}{
// 					"Result": {
// 						SeqNum: int(tcpHeader.SequenceNumber),
// 						Data:   tcpHeader.Data,
// 					},
// 				}
// 			}
// 		}
// 	}()

// 	var seqNum int
// 	var data []byte
// 	ticker := time.NewTicker(30 * time.Second)
// 	for {
// 		select {
// 		case <-ticker.C:
// 			return 0, []byte{}, fmt.Errorf("Timed out waiting for function")

// 		case resp := <-seqDataChan:
// 			result := resp["Result"]
// 			seqNum = result.SeqNum
// 			data = result.Data

// 			return seqNum, data, nil
// 		}
// 	}
// }

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

// // NetToHostShort converts a 16-bit integer from network to host byte order, aka "ntohs"
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

func htons(i uint16) uint16 {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, i)
	return binary.LittleEndian.Uint16(data)
}

func htonl(i uint32) uint16 {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint32(data, i)
	return binary.LittleEndian.Uint16(data)
}
