package utils

import "net"

type Flow7Tuple struct {
	SrcAddr, DstAddr                           net.IP
	SrcPort, DstPort                           uint16
	Protocol                                   uint8
	FlowStartMilliseconds, FlowEndMilliseconds uint64
}
