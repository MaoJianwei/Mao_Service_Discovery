package AuxDataProcessor

import (
	"MaoServerDiscovery/util"
	"encoding/json"
	"log"
)

const (
	p_MODULE_NAME = "GRPC-Detect-module"
)

type EnvTempProcessor struct {

}

type EnvTempData struct {
	EnvTemp float32 `json:"envTemp"`
}

func (e EnvTempProcessor) Process(auxData string) {
	auxDataMap := EnvTempData{}
	err := json.Unmarshal([]byte(auxData), &auxDataMap)
	if err != nil {
		util.MaoLogM(util.HOT_DEBUG, p_MODULE_NAME, "Fail to json.Unmarshal aux data. %s", err.Error())
		return
	}


	log.Println(auxDataMap.EnvTemp)
}



















