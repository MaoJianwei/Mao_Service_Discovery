package AuxDataProcessor

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/util"
	"time"
)

const (
	MODULE_NAME = "Aux-Data-Processor-module"
)

type AuxDataProcessorModule struct {
	processors []*MaoApi.AuxDataProcessor
}

func (a *AuxDataProcessorModule) AddProcessor(p *MaoApi.AuxDataProcessor) {
	a.processors = append(a.processors, p)
}

func (a *AuxDataProcessorModule) controlLoop() {
	for {
		time.Sleep(1 * time.Second)

		grpcKaModule := MaoCommon.ServiceRegistryGetGrpcKaModule()
		if grpcKaModule == nil {
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get GrpcKaModule")
			continue
		}

		serviceNodes := grpcKaModule.GetServiceInfo()
		for _, service := range serviceNodes {
			for _, processor := range a.processors {
				go (*processor).Process(service.OtherData)
			}
		}
	}
}

func (a *AuxDataProcessorModule) InitAuxDataProcessor() {
	a.processors = make([]*MaoApi.AuxDataProcessor, 0)
	go a.controlLoop()
}




















