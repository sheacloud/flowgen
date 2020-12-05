package main

import (
	"net"
	"testing"
	"time"

	"github.com/sheacloud/flowgen/utils"
	"github.com/vmware/go-ipfix/pkg/registry"
)

var exporter *utils.FlowExporter
var flow utils.Flow7Tuple

func init() {
	registry.LoadRegistry()

	exporter = utils.NewFlowExporter(net.ParseIP("127.0.0.1"), 9001)

	flow = utils.Flow7Tuple{
		SrcAddr:               net.ParseIP("192.168.0.1"),
		DstAddr:               net.ParseIP("192.168.1.1"),
		SrcPort:               50000,
		DstPort:               443,
		Protocol:              6,
		FlowStartMilliseconds: uint64(time.Now().UnixNano() / 1000000),
		FlowEndMilliseconds:   uint64(time.Now().UnixNano()/1000000) + 100,
		OctetCount:            4000,
		PacketCount:           5,
	}
}

func BenchmarkStandard(b *testing.B) {
	for n := 0; n < b.N; n++ {
		exporter.GenerateFlowMessage(flow)
	}
}

func BenchmarkInStack(b *testing.B) {
	for n := 0; n < b.N; n++ {
		exporter.GenerateFlowMessageInStack(flow)
	}
}
