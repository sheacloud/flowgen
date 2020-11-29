package main

import (
	"fmt"
	"net"
	"time"

	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
)

func init() {
	registry.LoadRegistry()
}

func main() {
	generator := FlowGenerator{}
	generator.Initialize(net.ParseIP("127.0.0.1"), 9001)

	flowStartTime := uint64(time.Now().UnixNano() / 1000000)
	flowEndTime := uint64(time.Now().UnixNano() / 1000000)
	srcAddr := net.ParseIP("192.168.10.122")
	dstAddr := net.ParseIP("100.90.80.70")

	fmt.Println("generating flows")
	for i := 0; i < 2200; i++ {
		generator.GenerateFlowMessage(srcAddr, dstAddr, 51250, 443, 6, flowStartTime, flowEndTime)
	}

	generator.SendDataSet()

	generator.CloseExporter()

	// Create data set with 1 data record
}

type FlowGenerator struct {
	exporter   *exporter.ExportingProcess
	dataSet    entities.Set
	templateID uint16
}

func (f *FlowGenerator) Initialize(collectorIp net.IP, collectorPort int) {
	udpAddr := net.UDPAddr{
		IP:   collectorIp,
		Port: collectorPort,
	}
	exporter, _ := exporter.InitExportingProcess(&udpAddr, 1, 0)
	f.exporter = exporter

	templateElementNames := []string{"sourceIPv4Address", "destinationIPv4Address", "sourceTransportPort", "destinationTransportPort", "protocolIdentifier", "flowStartMilliseconds", "flowEndMilliseconds"}
	// Create template record with two fields
	f.templateID = f.exporter.NewTemplateID()
	templateSet := entities.NewSet(entities.Template, f.templateID, false)
	elements := make([]*entities.InfoElementWithValue, 0)

	for _, elementName := range templateElementNames {
		element, err := registry.GetInfoElement(elementName, registry.IANAEnterpriseID)
		if err != nil {
			fmt.Printf("Did not find the element with name %v\n", elementName)
			return
		}
		ie := entities.NewInfoElementWithValue(element, nil)
		elements = append(elements, ie)
	}

	templateSet.AddRecord(elements, f.templateID)

	bytesSent, err := f.exporter.AddSetAndSendMsg(entities.Template, templateSet)
	if err != nil {
		fmt.Printf("Got error when sending record: %v\n", err)
		return
	}
	// Sleep for 2s for template refresh routine to get executed
	fmt.Printf("sent tempalte: %v\n", bytesSent)
	time.Sleep(2 * time.Second)

	dataSet := entities.NewSet(entities.Data, f.templateID, false)
	f.dataSet = dataSet
}

func (f *FlowGenerator) SendDataSet() {
	fmt.Println(f.dataSet.GetBuffLen())

	bytesSent, err := f.exporter.AddSetAndSendMsg(entities.Data, f.dataSet)
	if err != nil {
		fmt.Printf("Got error when sending record: %v\n", err)
		return
	}
	// 28 is the size of the IPFIX message including all headers (20 bytes)
	fmt.Println(bytesSent)

	f.dataSet = entities.NewSet(entities.Data, f.templateID, false)
}

func (f *FlowGenerator) CloseExporter() {
	f.exporter.CloseConnToCollector()
}

func (f *FlowGenerator) GenerateFlowMessage(srcAddr, dstAddr net.IP, srcPort, dstPort uint16, protocol uint8, flowStartMilliseconds, flowEndMilliseconds uint64) {
	elements := make([]*entities.InfoElementWithValue, 0)
	element, err := registry.GetInfoElement("sourceIPv4Address", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name sourceIPv4Address\n")
		return
	}
	ie := entities.NewInfoElementWithValue(element, srcAddr)
	elements = append(elements, ie)

	element, err = registry.GetInfoElement("destinationIPv4Address", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name destinationIPv4Address\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, dstAddr)
	elements = append(elements, ie)

	element, err = registry.GetInfoElement("sourceTransportPort", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name sourceTransportPort\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, srcPort)
	elements = append(elements, ie)

	element, err = registry.GetInfoElement("destinationTransportPort", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name destinationTransportPort\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, dstPort)
	elements = append(elements, ie)

	element, err = registry.GetInfoElement("protocolIdentifier", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name protocolIdentifier\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, protocol)
	elements = append(elements, ie)

	element, err = registry.GetInfoElement("flowStartMilliseconds", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name flowStartMilliseconds\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flowStartMilliseconds)
	elements = append(elements, ie)

	element, err = registry.GetInfoElement("flowEndMilliseconds", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name flowEndMilliseconds\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flowEndMilliseconds)
	elements = append(elements, ie)

	f.dataSet.AddRecord(elements, f.templateID)
}
