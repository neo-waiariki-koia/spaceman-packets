package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type PacketConstructor struct {
	SrcHost  string
	SrcPort  int
	DestHost string
	DestPort int
	Seq      int
	Ack      int
	Flags    []string
	Checksum int
	Data     []byte
}

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

func BuildTCPPacket(destHost string, destPort int, srcHost string, srcPort int, seq int, ack int, flags []string, checksum int, data []byte) []byte {
	packet := NewPacketConstructor(srcHost, srcPort, destHost, destPort, seq, ack, flags, checksum, data)

	//completeTCPPacket := append(buildEthernetHeader(), packet.buildIPHeader()...)
	completeTCPPacket := append(packet.buildIPHeader(), packet.buildTCPHeader()...)
	completeTCPPacket = append(completeTCPPacket, data...)

	return completeTCPPacket
}

func NewPacketConstructor(srcHost string, srcPort int, destHost string, destPort int, seq int, ack int, flags []string, checksum int, data []byte) *PacketConstructor {
	return &PacketConstructor{
		SrcHost:  srcHost,
		SrcPort:  srcPort,
		DestHost: destHost,
		DestPort: destPort,
		Seq:      seq,
		Ack:      ack,
		Flags:    flags,
		Checksum: checksum,
		Data:     data,
	}
}

func (pc *PacketConstructor) buildIPHeader() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, []byte{69, 0, 0, 40})   // Version, IHL, Type of Service | Total Length
	binary.Write(buf, binary.BigEndian, []byte{141, 245, 0, 0}) // Identification | Flags, Fragment Offset
	binary.Write(buf, binary.BigEndian, []byte{64, 6, 0, 0})    // TTL, Protocol | Header Checksum
	binary.Write(buf, binary.BigEndian, to4byte(pc.SrcHost))
	binary.Write(buf, binary.BigEndian, to4byte(pc.DestHost))

	bufBytes := buf.Bytes()

	checksum := calculateChecksum(bufBytes)

	buf2 := new(bytes.Buffer)
	binary.Write(buf2, binary.BigEndian, checksum)

	ipHeader := bufBytes[:10]
	ipHeader = append(ipHeader, buf2.Bytes()...)
	ipHeader = append(ipHeader, bufBytes[12:]...)

	return ipHeader
}

func (pc *PacketConstructor) buildTCPHeader() []byte {
	flagsInt := computeTCPFlags(pc.Flags)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(pc.SrcPort))
	binary.Write(buf, binary.BigEndian, uint16(pc.DestPort))
	binary.Write(buf, binary.BigEndian, uint32(pc.Seq))
	binary.Write(buf, binary.BigEndian, uint32(pc.Ack))

	mix := uint16(5)<<12 | // top 4 bits //data offset
		uint16(0)<<9 | // 3 bits
		uint16(flagsInt) // 3 bits
	binary.Write(buf, binary.BigEndian, mix)

	binary.Write(buf, binary.BigEndian, uint16(8192))
	binary.Write(buf, binary.BigEndian, uint16(0))
	binary.Write(buf, binary.BigEndian, uint16(0))

	bufBytes := buf.Bytes()

	source := to4byte(pc.SrcHost)
	dest := to4byte(pc.DestHost)

	pseudoHeader := []byte{
		source[0], source[1], source[2], source[3],
		dest[0], dest[1], dest[2], dest[3],
		0,                      // zero
		6,                      // protocol number (6 == TCP)
		0, byte(len(bufBytes)), // TCP length (16 bits), not inc pseudo header
	}

	sumThis := make([]byte, 0, len(pseudoHeader)+len(bufBytes))
	sumThis = append(sumThis, pseudoHeader...)
	sumThis = append(sumThis, bufBytes...)

	checksum := calculateChecksum(sumThis)

	buf2 := new(bytes.Buffer)
	binary.Write(buf2, binary.BigEndian, checksum)

	tcpHeader := bufBytes[:16]
	tcpHeader = append(tcpHeader, buf2.Bytes()...)
	tcpHeader = append(tcpHeader, bufBytes[18:]...)

	// Pad to min tcp header size, which is 20 bytes (5 32-bit words)
	pad := 20 - len(tcpHeader)
	for i := 0; i < pad; i++ {
		tcpHeader = append(tcpHeader, 0)
	}
	return tcpHeader
}

func to4byte(addr string) [4]byte {
	parts := strings.Split(addr, ".")
	b0, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalf("to4byte: %s (latency works with IPv4 addresses only, but not IPv6!)\n", err)
	}
	b1, _ := strconv.Atoi(parts[1])
	b2, _ := strconv.Atoi(parts[2])
	b3, _ := strconv.Atoi(parts[3])
	return [4]byte{byte(b0), byte(b1), byte(b2), byte(b3)}
}

func computeTCPFlags(flags []string) int {
	var fin, syn, rst, psh, ack, urg int
	for _, flag := range flags {
		switch flag {
		case "fin":
			fin = 1
		case "syn":
			syn = 1
		case "rst":
			rst = 1
		case "psh":
			psh = 1
		case "ack":
			ack = 1
		case "urg":
			urg = 1
		}
	}
	computedFlags := fin + (syn << 1) + (rst << 2) + (psh << 3) + (ack << 4) + (urg << 5)

	return computedFlags
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
		sum += uint32(data[dataLength-1])
	}

	// Add back any carry, and any carry from adding the carry
	sum = (sum >> 16) + (sum & 0xffff)
	sum = sum + (sum >> 16)

	// Bitwise complement
	return uint16(^sum)
}

func getMacAddr() (*net.Interface, error) {
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Fatal("get link by name:", err)
	}

	fmt.Println(iface.HardwareAddr)

	return iface, nil
}

func buildEthernetHeader() []byte {
	ipv4 := [2]byte{0x08, 0x00}

	src, err := getMacAddr()
	if err != nil {
		log.Fatalf("buildEthernetHeader: %s", err)
	}

	srcMac := src.HardwareAddr

	dstMac := srcMac

	ethHeader := []byte{
		dstMac[0], dstMac[1], dstMac[2], dstMac[3], dstMac[4], dstMac[5],
		srcMac[0], srcMac[1], srcMac[2], srcMac[3], srcMac[4], srcMac[5],
		ipv4[0], ipv4[1], // your custom ethertype
	}

	return ethHeader
}
