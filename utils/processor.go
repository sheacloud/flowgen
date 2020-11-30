package utils

import (
	"fmt"
	"sync"
)

type FlowProcessor struct {
	FlowExporters []*FlowExporter
	Channel       chan Flow7Tuple
	Quit          chan bool
	EndWG         sync.WaitGroup
	BufferSize    int
}

func NewFlowProcessor(generators []*FlowExporter, bufferSize int) *FlowProcessor {
	channel := make(chan Flow7Tuple, bufferSize)

	fp := FlowProcessor{
		FlowExporters: generators,
		Channel:       channel,
		Quit:          make(chan bool),
		BufferSize:    bufferSize,
	}

	return &fp
}

func (fp *FlowProcessor) PutFlow(flow Flow7Tuple) {
	fp.Channel <- flow
}

func (fp *FlowProcessor) Start() {
	fp.EndWG.Add(len(fp.FlowExporters))

	for i := 0; i < len(fp.FlowExporters); i++ {
		go func(flowgen *FlowExporter) {
			fmt.Println("Starting flow processor")
			for flow := range fp.Channel {
				if flowgen.GetCurrentMessageSize() >= 8955 {
					flowgen.SendDataSet()
				}
				flowgen.GenerateFlowMessage(flow)
			}
			fmt.Println("flow processor shutting down")
			flowgen.SendDataSet()
			flowgen.CloseExporter()
			fp.EndWG.Done()
		}(fp.FlowExporters[i])
	}

}

func (fp *FlowProcessor) Stop() {
	close(fp.Channel)
	fp.EndWG.Wait()
}
