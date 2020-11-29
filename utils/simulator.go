package utils

import (
	"fmt"
	"net"
	"time"
)

type FlowSimulator struct {
	FlowProcessor            *FlowProcessor
	Ticker                   *time.Ticker
	Quit                     chan bool
	TickIntervalMilliseconds int64
	FlowsPerTick             int
}

func NewFlowSimulator(flowProcessor *FlowProcessor, tickIntervalMilliseconds int64, flowsPerTick int) *FlowSimulator {
	return &FlowSimulator{
		FlowProcessor:            flowProcessor,
		Quit:                     make(chan bool),
		TickIntervalMilliseconds: tickIntervalMilliseconds,
		FlowsPerTick:             flowsPerTick,
	}
}

func (fs *FlowSimulator) CreateFlow() Flow7Tuple {
	flowStartTime := uint64(time.Now().UnixNano() / 1000000)
	flowEndTime := uint64(time.Now().UnixNano() / 1000000)
	srcAddr := net.ParseIP("192.168.10.122")
	dstAddr := net.ParseIP("100.90.80.70")

	flow := Flow7Tuple{
		SrcAddr:               srcAddr,
		DstAddr:               dstAddr,
		SrcPort:               51250,
		DstPort:               443,
		Protocol:              6,
		FlowStartMilliseconds: flowStartTime,
		FlowEndMilliseconds:   flowEndTime,
	}

	return flow
}

func (fs *FlowSimulator) Start() {
	fs.Ticker = time.NewTicker(time.Duration(fs.TickIntervalMilliseconds) * time.Millisecond)

	go func() {
		for {
			select {
			case <-fs.Quit:
				return
			case _ = <-fs.Ticker.C:
				if float32(len(fs.FlowProcessor.Channel)) >= 0.75*float32(fs.FlowProcessor.BufferSize) {
					fmt.Println("skipping tick as channel is >75% full")
					continue
				}
				for i := 0; i < fs.FlowsPerTick; i++ {
					fs.FlowProcessor.PutFlow(fs.CreateFlow())
				}
			}
		}
	}()
}

func (fs *FlowSimulator) Stop() {
	fs.Ticker.Stop()
	fs.Quit <- true
}
