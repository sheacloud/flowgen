package utils

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"

	"gonum.org/v1/gonum/stat/distuv"
)

type FlowSimulator struct {
	FlowProcessor              *FlowProcessor
	Ticker                     *time.Ticker
	Quit                       chan bool
	TickIntervalMilliseconds   int
	FlowsPerTick               int
	SampleFlow                 Flow7Tuple
	FlowRecordsCreated         uint64
	LatencyDistribution        distuv.Rander // probability distribution for how many milliseconds set the latency between request and response flows
	FlowDurationDistribution   distuv.Rander // probability distribution for how many milliseconds long the flows should be
	FlowTimeJitterDistribution distuv.Rander // probability distribution for how many milliseconds to jitter the flow by
}

func NewFlowSimulator(flowProcessor *FlowProcessor, tickIntervalMilliseconds int, flowsPerTick int) *FlowSimulator {
	return &FlowSimulator{
		FlowProcessor:            flowProcessor,
		Quit:                     make(chan bool),
		TickIntervalMilliseconds: tickIntervalMilliseconds,
		FlowsPerTick:             flowsPerTick,
		FlowRecordsCreated:       0,
		LatencyDistribution: distuv.Normal{
			Mu:    100,
			Sigma: 50,
		},
		FlowDurationDistribution: distuv.Normal{
			Mu:    3000,
			Sigma: 5000,
		},
		FlowTimeJitterDistribution: distuv.Normal{
			Mu:    0,
			Sigma: float64(tickIntervalMilliseconds) / 3,
		},
	}
}

func (fs *FlowSimulator) CreateFlow() (Flow7Tuple, Flow7Tuple) {
	flowEndTime := uint64(time.Now().UnixNano()/1000000) + uint64(fs.FlowTimeJitterDistribution.Rand())
	flowDuration := uint64(math.Max(1, fs.FlowDurationDistribution.Rand())) // can't have a flow start after it ended, limit flows to at least 1 millisecond in duration
	flowStartTime := flowEndTime - flowDuration

	srcAddr := GenerateRandomIP(net.IP{0, 0, 0, 0}, net.IPv4Mask(0, 0, 0, 0))
	dstAddr := GenerateRandomIP(net.IP{0, 0, 0, 0}, net.IPv4Mask(0, 0, 0, 0))

	var srcPort uint16 = uint16(rand.Intn(32768)) + 32768
	var dstPort uint16 = uint16(rand.Intn(2048)) + 1
	var protocol byte = 6

	responseLatency := uint64(math.Max(1, fs.LatencyDistribution.Rand())) // can't have negative latency on a network, limit it to at least 1 millisecond

	srcFlow := Flow7Tuple{
		SrcAddr:               srcAddr,
		DstAddr:               dstAddr,
		SrcPort:               srcPort,
		DstPort:               dstPort,
		Protocol:              protocol,
		FlowStartMilliseconds: flowStartTime,
		FlowEndMilliseconds:   flowEndTime,
		OctetCount:            rand.Uint64(),
		PacketCount:           rand.Uint64(),
	}

	dstFlow := Flow7Tuple{
		SrcAddr:               dstAddr,
		DstAddr:               srcAddr,
		SrcPort:               dstPort,
		DstPort:               srcPort,
		Protocol:              protocol,
		FlowStartMilliseconds: flowStartTime + responseLatency,
		FlowEndMilliseconds:   flowEndTime + responseLatency,
		OctetCount:            rand.Uint64(),
		PacketCount:           rand.Uint64(),
	}

	fs.FlowRecordsCreated++

	return srcFlow, dstFlow
}

func (fs *FlowSimulator) Start() {
	fs.Ticker = time.NewTicker(time.Duration(fs.TickIntervalMilliseconds) * time.Millisecond)

	go func() {
		var srcFlow, dstFlow Flow7Tuple
		for {
			select {
			case <-fs.Quit:
				return
			case _ = <-fs.Ticker.C:
				if fs.FlowProcessor.BufferSize > 0 {
					if (fs.FlowProcessor.BufferSize - len(fs.FlowProcessor.Channel)) < fs.FlowsPerTick {
						fmt.Println("skipping tick as channel cannot receive")
						continue
					}
				}
				for i := 0; i < fs.FlowsPerTick; i++ {
					srcFlow, dstFlow = fs.CreateFlow()
					fs.FlowProcessor.Channel <- srcFlow
					fs.FlowProcessor.Channel <- dstFlow
				}
			}
		}
	}()
}

func (fs *FlowSimulator) Stop() {
	fs.Ticker.Stop()
	fs.Quit <- true
}
