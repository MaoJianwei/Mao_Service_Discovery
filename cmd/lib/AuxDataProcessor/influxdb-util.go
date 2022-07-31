package AuxDataProcessor
//
//import (
//	"MaoServerDiscovery/util"
//	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
//	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
//	"time"
//)
//
////a.influxdbUrl = influxdbUrl
////a.influxdbToken = influxdbToken
////a.influxdbOrgBucket = influxdbOrgBucket
//
//func nat66UploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, v6In uint64, v6Out uint64) {
//	// write point asynchronously
//	(*writeAPI).WritePoint(
//		influxdb2.NewPointWithMeasurement("NAT66_Gateway").
//			AddTag("Geo", "Beijing-HQ").
//			AddField("v6In", v6In).
//			AddField("v6Out", v6Out).
//			SetTime(time.Now()))
//	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
//	//(*writeAPI).Flush()
//}
//
//func envTempUploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, envTemperature float64) {
//	// write point asynchronously
//	(*writeAPI).WritePoint(
//		influxdb2.NewPointWithMeasurement("Temperature").
//			AddTag("Geo", "Beijing-HQ").
//			AddField("env", envTemperature).
//			SetTime(time.Now()))
//	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
//	//(*writeAPI).Flush()
//}
//
//func foo() {
//	var influxdbClient influxdb2.Client
//	var influxdbWriteAPI influxdb2Api.WriteAPI
//
//	influxdbClient = influxdb2.NewClient(influxdbUrl, influxdbToken)
//	defer influxdbClient.Close()
//	influxdbWriteAPI = influxdbClient.WriteAPI(influxdbOrgBucket, influxdbOrgBucket)
//
//
//}