package packet

import (
	"bytes"
	"encoding/binary"
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

func BuildTCPPacket(destHost string, destPort int, srcHost string, srcPort int, seq int, ack int, flags []string, checksum int, data []byte) []byte {
	packet := NewPacketConstructor(destHost, destPort, srcHost, srcPort, seq, ack, flags, checksum, data)

	completeTCPPacket := append(packet.buildIPHeader(), packet.buildTCPHeader()...)
	completeTCPPacket = append(completeTCPPacket, data...)

	return completeTCPPacket
}
