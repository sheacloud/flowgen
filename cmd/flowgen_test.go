package main

import (
	"net"
	"testing"
	"time"

	"github.com/sheacloud/flowgen/utils"
	"github.com/vmware/go-ipfix/pkg/registry"
)

func init() {
	registry.LoadRegistry()
}

func BenchmarkSim(b *testing.B) {
	for n := 0; n < b.N; n++ {
		generator := utils.FlowGenerator{}
		generator.Initialize(net.ParseIP("127.0.0.1"), 9001)

		processor := utils.NewFlowProcessor(&generator, 10000)
		processor.Start()

		simulator := utils.NewFlowSimulator(processor, 3, 2000)

		simulator.Start()

		time.Sleep(10 * time.Second)

		simulator.Stop()
		processor.Stop()

		generator.CloseExporter()
	}
}
