package AuxDataProcessor

import (
	"MaoServerDiscovery/cmd/lib/InfluxDB"
	"MaoServerDiscovery/util"
	"encoding/json"
	"time"
)

const (
	p_EnvTemp_MODULE_NAME = "Env-Temperature-module"
)

type EnvTempProcessor struct {

}

type EnvTempData struct {
	EnvGeo string `json:"envGeo"`
	EnvTime string `json:"envTime"`
	EnvTemp float32 `json:"envTemp"`
}

func (e EnvTempProcessor) Process(auxData string) {

	auxDataMap := EnvTempData{}
	err := json.Unmarshal([]byte(auxData), &auxDataMap)
	if err != nil {
		util.MaoLogM(util.WARN, p_EnvTemp_MODULE_NAME, "Fail to json.Unmarshal aux data. %s", err.Error())
		return
	}
	if auxDataMap.EnvTime == "" {
		return // not contain Environment Temperature data
	}

	envTime, err := time.Parse(time.RFC3339Nano, auxDataMap.EnvTime)
	if err != nil {
		util.MaoLogM(util.WARN, p_EnvTemp_MODULE_NAME, "Fail to parse time string as RFC3339Nano format, %s, err: %s", auxDataMap.EnvTime, err.Error())
		return
	}
	util.MaoLogM(util.HOT_DEBUG, p_EnvTemp_MODULE_NAME, "Get temp %f, %s", auxDataMap.EnvTemp, time.Now().String())

	InfluxDB.EnvTempUploadInfluxdb(auxDataMap.EnvGeo, envTime, auxDataMap.EnvTemp)
}



















