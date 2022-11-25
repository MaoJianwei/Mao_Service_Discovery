package InfluxDB

import (
	"MaoServerDiscovery/util"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

const (
	MODULE_NAME = "InfluxDB-Util"
)

var (
	config_influxdbUrl = ""
	config_influxdbToken = ""
	config_influxdbOrgBucket = ""
)

//a.influxdbUrl = influxdbUrl
//a.influxdbToken = influxdbToken
//a.influxdbOrgBucket = influxdbOrgBucket

func nat66UploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, v6In uint64, v6Out uint64) {
	// write point asynchronously
	(*writeAPI).WritePoint(
		influxdb2.NewPointWithMeasurement("NAT66_Gateway").
			AddTag("Geo", "Beijing-HQ").
			AddField("v6In", v6In).
			AddField("v6Out", v6Out).
			SetTime(time.Now()))
	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
	//(*writeAPI).Flush()
}

//
func EnvTempUploadInfluxdb(geo string, timestamp time.Time, envTemperature float32) {
	client, writeAPI := CreateClientAndWriteAPI()
	if writeAPI == nil {
		if time.Now().Second() % 10 == 0 {
			// suppress the number of logs
			util.MaoLogM(util.WARN, MODULE_NAME, "Fail to upload env temp, InfluxDB API URLs haven't configured yet.")
		}
		return
	}
	defer (*client).Close()

	// write point asynchronously
	(*writeAPI).WritePoint(
		influxdb2.NewPointWithMeasurement("Temperature").
			AddTag("Geo", geo).
			AddField("env", envTemperature).
			SetTime(timestamp))
	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
	//(*writeAPI).Flush()
}

func CreateClientAndWriteAPI() (*influxdb2.Client, *influxdb2Api.WriteAPI) {
	if config_influxdbUrl == "" {
		return nil, nil
	}

	influxdbClient := influxdb2.NewClient(config_influxdbUrl, config_influxdbToken)
	influxdbWriteAPI := influxdbClient.WriteAPI(config_influxdbOrgBucket, config_influxdbOrgBucket)
	return &influxdbClient, &influxdbWriteAPI
}

func ConfigInfluxdbUtils(influxdbUrl, influxdbToken, influxdbOrgBucket string) {
	config_influxdbUrl = influxdbUrl
	config_influxdbToken = influxdbToken
	config_influxdbOrgBucket = influxdbOrgBucket
}