package utils

import (
	"fmt"
	"sync"
)

type FlowProcessor struct {
	FlowGenerator *FlowGenerator
	Channel       chan Flow7Tuple
	Quit          chan bool
	EndWG         sync.WaitGroup
	BufferSize    int
}

func NewFlowProcessor(generator *FlowGenerator, bufferSize int) *FlowProcessor {
	channel := make(chan Flow7Tuple, bufferSize)

	fp := FlowProcessor{
		FlowGenerator: generator,
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
	fp.EndWG.Add(1)
	go func() {
		for flow := range fp.Channel {
			if fp.FlowGenerator.GetCurrentMessageSize() >= 65000 {
				fp.FlowGenerator.SendDataSet()
			}
			fp.FlowGenerator.GenerateFlowMessage(flow)
		}
		fmt.Println("flow processor shutting down")
		fp.FlowGenerator.SendDataSet()
		fp.EndWG.Done()
	}()
}

func (fp *FlowProcessor) Stop() {
	close(fp.Channel)
	fp.EndWG.Wait()
}
