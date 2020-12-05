package utils

import (
	"bytes"
	"fmt"
	"net"

	"github.com/vmware/go-ipfix/pkg/entities"
	"github.com/vmware/go-ipfix/pkg/exporter"
	"github.com/vmware/go-ipfix/pkg/registry"
)

type FlowExporter struct {
	exporter        *exporter.ExportingProcess
	dataSet         entities.Set
	templateID      uint16
	FlowRecordsSent uint64
	RecordBuffer    bytes.Buffer
	// ElementBuffer   []*entities.InfoElementWithValue

}

func NewFlowExporter(collectorIp net.IP, collectorPort int) *FlowExporter {
	fg := FlowExporter{}

	udpAddr := net.UDPAddr{
		IP:   collectorIp,
		Port: collectorPort,
	}
	exporter, _ := exporter.InitExportingProcess(&udpAddr, 1, 0)
	fg.exporter = exporter

	templateElementNames := []string{"sourceIPv4Address", "destinationIPv4Address", "sourceTransportPort", "destinationTransportPort", "protocolIdentifier", "flowStartMilliseconds", "flowEndMilliseconds", "octetTotalCount", "packetTotalCount"}
	// 4 + 4 + 2 + 2 + 1 + 8 + 8 + 8 + 8
	// 8 + 4 + 1 + 32
	// 45 bytes per flow, 424880 flows =

	// Create template record with two fields
	fg.templateID = fg.exporter.NewTemplateID()
	templateSet := entities.NewSet(entities.Template, fg.templateID, false)
	elements := make([]*entities.InfoElementWithValue, 0)

	for _, elementName := range templateElementNames {
		element, err := registry.GetInfoElement(elementName, registry.IANAEnterpriseID)
		if err != nil {
			fmt.Printf("Did not find the element with name %v\n", elementName)
			return nil
		}
		ie := entities.NewInfoElementWithValue(element, nil)
		elements = append(elements, ie)
	}

	templateSet.AddRecord(elements, fg.templateID)

	bytesSent, err := fg.exporter.AddSetAndSendMsg(entities.Template, templateSet)
	if err != nil {
		fmt.Printf("Got error when sending record: %v\n", err)
		return nil
	}
	// Sleep for 2s for template refresh routine to get executed
	fmt.Printf("sent template: %v\n", bytesSent)

	dataSet := entities.NewSet(entities.Data, fg.templateID, false)
	fg.dataSet = dataSet

	fg.RecordBuffer = bytes.Buffer{}
	fg.RecordBuffer.Grow(100)

	return &fg
}

func (f *FlowExporter) GetCurrentMessageSize() int {
	return 20 + int(f.dataSet.GetBuffLen())
}

func (f *FlowExporter) SendDataSet() {

	_, err := f.exporter.AddSetAndSendMsg(entities.Data, f.dataSet)
	if err != nil {
		fmt.Printf("Got error when sending record: %v\n", err)
		return
	}
	f.FlowRecordsSent += uint64(len(f.dataSet.GetRecords()))

	f.dataSet = entities.NewSet(entities.Data, f.templateID, false)
}

func (f *FlowExporter) CloseExporter() {
	f.exporter.CloseConnToCollector()
}

func (f *FlowExporter) GenerateFlowMessage(flow Flow7Tuple) {
	elementBuffer := make([]*entities.InfoElementWithValue, 9)
	element, err := registry.GetInfoElement("sourceIPv4Address", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name sourceIPv4Address\n")
		return
	}
	ie := entities.NewInfoElementWithValue(element, flow.SrcAddr)
	elementBuffer[0] = ie

	element, err = registry.GetInfoElement("destinationIPv4Address", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name destinationIPv4Address\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.DstAddr)
	elementBuffer[1] = ie

	element, err = registry.GetInfoElement("sourceTransportPort", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name sourceTransportPort\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.SrcPort)
	elementBuffer[2] = ie

	element, err = registry.GetInfoElement("destinationTransportPort", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name destinationTransportPort\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.DstPort)
	elementBuffer[3] = ie

	element, err = registry.GetInfoElement("protocolIdentifier", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name protocolIdentifier\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.Protocol)
	elementBuffer[4] = ie

	element, err = registry.GetInfoElement("flowStartMilliseconds", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name flowStartMilliseconds\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.FlowStartMilliseconds)
	elementBuffer[5] = ie

	element, err = registry.GetInfoElement("flowEndMilliseconds", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name flowEndMilliseconds\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.FlowEndMilliseconds)
	elementBuffer[6] = ie

	element, err = registry.GetInfoElement("octetTotalCount", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name octetTotalCount\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.OctetCount)
	elementBuffer[7] = ie

	element, err = registry.GetInfoElement("packetTotalCount", registry.IANAEnterpriseID)
	if err != nil {
		fmt.Printf("Did not find the element with name packetTotalCount\n")
		return
	}
	ie = entities.NewInfoElementWithValue(element, flow.PacketCount)
	elementBuffer[8] = ie

	f.dataSet.AddRecord(elementBuffer, f.templateID)
}

func (f *FlowExporter) GenerateFlowMessageInStack(flow Flow7Tuple) {

	record := entities.NewDataRecord(f.templateID, f.RecordBuffer)
	record.PrepareRecord()
	record.EncodeIPv4Address(flow.SrcAddr)
	record.EncodeIPv4Address(flow.DstAddr)
	record.EncodeUint16(flow.SrcPort)
	record.EncodeUint16(flow.DstPort)
	record.EncodeUint8(flow.Protocol)
	record.EncodeUint64(flow.FlowStartMilliseconds)
	record.EncodeUint64(flow.FlowEndMilliseconds)
	record.EncodeUint64(flow.OctetCount)
	record.EncodeUint64(flow.PacketCount)

	f.dataSet.EncodeDataRecord(record)

	if f.GetCurrentMessageSize() >= 60000 {
		f.dataSet.ResetBuffer()
	}

	f.RecordBuffer.Reset()

}
