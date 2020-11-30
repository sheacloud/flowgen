package utils

import (
	"math/rand"
	"net"
)

type Flow7Tuple struct {
	SrcAddr, DstAddr                           net.IP
	SrcPort, DstPort                           uint16
	Protocol                                   uint8
	FlowStartMilliseconds, FlowEndMilliseconds uint64
	OctetCount, PacketCount                    uint64
}

func GenerateRandomIP(network net.IP, mask net.IPMask) net.IP {
	randomIP := net.IP{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		octetRange := int(255 - mask[i])
		randomIP[i] = network[i] + byte(rand.Intn(octetRange+1))
	}

	return randomIP
}
