package main

import (
	"fmt"
	"net"
	"time"

	"github.com/sheacloud/flowgen/utils"
	"github.com/vmware/go-ipfix/pkg/registry"
)

func init() {
	registry.LoadRegistry()
}

func main() {
	targetFlowsPerSecond := 1000
	ticksPerSecond := 10
	numFlowExporters := 8
	numFlowSimulators := 8
	bufferSize := 0

	tickInterval := 1000 / ticksPerSecond
	flowsPerTick := (targetFlowsPerSecond / ticksPerSecond)
	if flowsPerTick < numFlowSimulators {
		fmt.Println("Cannot have more flow simulators than flows per tick")
		return
	}
	flowsPerTickPerSimulator := flowsPerTick / numFlowSimulators

	fmt.Printf("Spinning up flow simulation with:\n\t%v ticks per second\n\t%v flows per tick, per simulator\n\t%v simulators\nFor a total of %v flows/second\n\n", ticksPerSecond, flowsPerTickPerSimulator, numFlowSimulators, (ticksPerSecond * flowsPerTickPerSimulator * numFlowSimulators))

	generators := []*utils.FlowExporter{}

	for i := 0; i < numFlowExporters; i++ {
		generator := utils.NewFlowExporter(net.ParseIP("127.0.0.1"), 9001)
		generators = append(generators, generator)
	}

	processor := utils.NewFlowProcessor(generators, bufferSize)
	processor.Start()

	simulators := []*utils.FlowSimulator{}

	for i := 0; i < numFlowSimulators; i++ {
		simulator := utils.NewFlowSimulator(processor, tickInterval, flowsPerTickPerSimulator)
		simulators = append(simulators, simulator)
	}

	for _, simulator := range simulators {
		simulator.Start()
	}

	time.Sleep(60 * time.Second)

	fmt.Println("shutting down simulators")
	for _, simulator := range simulators {
		simulator.Stop()
	}

	processor.Stop()

	var totalFlowsSent uint64
	for _, generator := range generators {
		totalFlowsSent += generator.FlowRecordsSent
		// fmt.Printf("generator%v sent %v flow records\n", i, generator.FlowRecordsSent)
	}
	fmt.Printf("total flows sent: %v\n", totalFlowsSent)

	var totalFlowsCreated uint64
	for _, generator := range generators {
		totalFlowsCreated += generator.FlowRecordsSent
		// fmt.Printf("generator%v sent %v flow records\n", i, generator.FlowRecordsSent)
	}
	fmt.Printf("total flows created: %v\n", totalFlowsCreated)
}
