package packet

import (
	"bytes"
	"encoding/binary"
	"net"
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

func UnmarshalIP(data []byte) *IPHeader {
	var ipHeader IPHeader

	firstSection := binary.BigEndian.Uint16(data[0:2])
	ipHeader.Version = byte(firstSection >> 12) // top 4 bits

	thirdSection := binary.BigEndian.Uint16(data[8:10])
	_ = byte(thirdSection >> 8)                 // top 8 bits
	ipHeader.Protocol = byte(thirdSection >> 0) // second 8 bits

	ipHeader.SourceAddress = binary.BigEndian.Uint32(data[12:16])
	ipHeader.DestinationAddress = binary.BigEndian.Uint32(data[16:20])

	ipHeader.Data = data[20:]

	ipHeader.sAddr = ipByteToString(ipHeader.SourceAddress)
	ipHeader.dAddr = ipByteToString(ipHeader.DestinationAddress)

	return &ipHeader
}

func ipByteToString(ipLong uint32) string {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	ip := net.IP(ipByte)
	return ip.String()
}

func (ih *IPHeader) validateTcpPacket() bool {
	return ih.Protocol == 6
}

func (ih *IPHeader) MarshalIP() []byte {
	buf := new(bytes.Buffer)

	sourceAddr := to4byte(ih.sAddr)
	destAddr := to4byte(ih.dAddr)

	binary.Write(buf, binary.BigEndian, []byte{69, 0, 0, 40})   // Version, IHL, Type of Service | Total Length
	binary.Write(buf, binary.BigEndian, []byte{141, 245, 0, 0}) // Identification | Flags, Fragment Offset
	binary.Write(buf, binary.BigEndian, []byte{64, 6, 0, 0})    // TTL, Protocol | Header Checksum
	binary.Write(buf, binary.BigEndian, sourceAddr)
	binary.Write(buf, binary.BigEndian, destAddr)

	bufBytes := buf.Bytes()

	checksum := calculateChecksum(bufBytes)

	buf2 := new(bytes.Buffer)
	binary.Write(buf2, binary.BigEndian, checksum)

	ipHeader := bufBytes[:10]
	ipHeader = append(ipHeader, buf2.Bytes()...)
	ipHeader = append(ipHeader, bufBytes[12:]...)

	return ipHeader
}
