package packet

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
	"strings"
)

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

func tcpFlags(flags []string) int {
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

func (tcp *TCPHeader) MarshalTCP(sourceAddr, destAddr string, seq, ack uint32, flags []string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, tcp.SourcePort)
	binary.Write(buf, binary.BigEndian, tcp.DestinationPort)
	binary.Write(buf, binary.BigEndian, seq)
	binary.Write(buf, binary.BigEndian, ack)

	flagsInt := tcpFlags(flags)
	mix := uint16(5)<<12 | // top 4 bits //data offset
		uint16(0)<<9 | // 3 bits
		uint16(flagsInt) // 3 bits
	binary.Write(buf, binary.BigEndian, mix)

	binary.Write(buf, binary.BigEndian, uint16(8192))
	binary.Write(buf, binary.BigEndian, uint16(0))
	binary.Write(buf, binary.BigEndian, uint16(0))

	bufBytes := buf.Bytes()

	source := to4byte(sourceAddr)
	dest := to4byte(destAddr)
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

func UnmarshalTCP(data []byte) *TCPHeader {
	var tcp TCPHeader

	tcp.SourcePort = binary.BigEndian.Uint16(data[0:2])
	tcp.DestinationPort = binary.BigEndian.Uint16(data[2:4])
	tcp.SequenceNumber = binary.BigEndian.Uint32(data[4:8])
	tcp.AcknowlegementNumber = binary.BigEndian.Uint32(data[8:12])
	tcp.DataOffset = data[12] >> 4
	tcp.Flags = data[13]
	tcp.Checksum = binary.BigEndian.Uint16(data[16:18])

	dataStart := int(tcp.DataOffset) * 4
	tcp.Data = data[dataStart:]

	return &tcp
}
