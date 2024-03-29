package branch

import (
	"MaoServerDiscovery/cmd/api/data"
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	util "MaoServerDiscovery/util"
	"encoding/json"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
	"strings"

	"context"
	"google.golang.org/grpc"
	"net"
	"os/exec"
	"strconv"
	"time"
)

const (
	c_MODULE_NAME = "General-Client"
)

var (
	envTemp float64
	gpsLast *data.GpsData
)

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

/*
   return v6In,v6Out,error
*/
func getNat66GatewayData() (uint64, uint64, error) {
	ccc := exec.Command("/bin/bash", "-c", "ip6tables -nvxL FORWARD | grep MaoIPv6In | awk '{printf $2}'")
	ininin, err := ccc.CombinedOutput()
	if err != nil {
		util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to get MaoIPv6In, %s", err.Error())
		return 0, 0, err
	}
	ccc = exec.Command("/bin/bash", "-c", "ip6tables -nvxL FORWARD | grep MaoIPv6Out | awk '{printf $2}'")
	outoutout, err := ccc.CombinedOutput()
	if err != nil {
		util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to get MaoIPv6Out, %s", err.Error())
		return 0, 0, err
	}
	v6In, err := strconv.ParseUint(string(ininin), 10, 64)
	if err != nil {
		util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to parse MaoIPv6In, %s", err.Error())
		return 0, 0, err
	}
	v6Out, err := strconv.ParseUint(string(outoutout), 10, 64)
	if err != nil {
		util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to parse MaoIPv6Out, %s", err.Error())
		return 0, 0, err
	}
	util.MaoLogM(util.DEBUG, c_MODULE_NAME, "v6In: %d , v6Out: %d", v6In, v6Out)
	return v6In, v6Out, nil
}



func envTempUploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, envTemperature float64) {
	// write point asynchronously
	(*writeAPI).WritePoint(
		influxdb2.NewPointWithMeasurement("Temperature").
			AddTag("Geo", "Beijing-HQ").
			AddField("env", envTemperature).
			SetTime(time.Now()))
	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
	//(*writeAPI).Flush()
}

func updateEnvironmentTemperature() {
	for {
		ccc := exec.Command("/bin/bash", "-c", "cat /sys/bus/w1/devices/28-00141093caff/w1_slave")
		w1Data, err := ccc.CombinedOutput()
		if err == nil {

			w1DataSplit := strings.Split(string(w1Data), "\n")
			if len(w1DataSplit) == 3 {

				tempText := strings.Split(w1DataSplit[1], "=")
				if len(tempText) == 2 {

					temp, err := strconv.ParseFloat(tempText[1], 64)
					if err == nil {
						envTemp = temp / 1000
						util.MaoLogM(util.DEBUG, c_MODULE_NAME, "Get envTemp: %f, %f", temp, envTemp)
					} else {
						util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse temperature text, %s", err.Error())
					}
				} else {
					util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse 1-line protocol data slice, the number of elements is not 2, in fact %d", len(tempText))
				}
			} else {
				util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse 1-line protocol data, the lines of the result text is not 3, in fact %d.", len(w1DataSplit))
			}
		} else {
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to get 1-line protocol data, %s", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func gpsDataUploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, gpsData *data.GpsData) {
	// write point asynchronously
	(*writeAPI).WritePoint(
		influxdb2.NewPointWithMeasurement("GPS").
			AddTag("Geo", "Beijing-HQ").
			AddField("GPS_Timestamp", gpsData.Timestamp).
			AddField("GPS_Latitude", gpsData.Latitude).
			AddField("GPS_Longitude", gpsData.Longitude).
			AddField("GPS_Altitude", gpsData.Altitude).
			AddField("GPS_Satellite", gpsData.Satellite).
			AddField("GPS_Hdop", gpsData.Hdop).
			AddField("GPS_Vdop", gpsData.Vdop).
			SetTime(time.Now()))
	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
	//(*writeAPI).Flush()
}

func updateGpsInfo(gpsApiData string) (*data.GpsData) {

	var err1 error
	var err2 error
	var err3 error
	var err4 error

	newGps := &data.GpsData{}

	items := strings.Split(gpsApiData, ";")
	for _, item := range items {

		kv := strings.Split(item, "=")
		switch kv[0] {
		case "GPS":
			datas := strings.Split(kv[1], ",")
			if datas[0] == "lost" || datas[0] == "init" {
				return nil
			}

			newGps.Latitude, err1 = strconv.ParseFloat(datas[0], 64)
			newGps.Longitude, err2 = strconv.ParseFloat(datas[1], 64)
			newGps.Altitude, err3 = strconv.ParseFloat(datas[2], 64)
			newGps.Satellite, err4 = strconv.ParseUint(datas[3], 10, 64)

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse GPS data, %s; %s; %s; %s",
					err1.Error(), err2.Error(), err3.Error(), err4.Error())
				return nil
			}

		case "GpsTime":
			newGps.Timestamp = kv[1]

		case "GPS_Precision":
			datas := strings.Split(kv[1], ",")
			newGps.Hdop, err1 = strconv.ParseFloat(datas[0], 64)
			newGps.Vdop, err2 = strconv.ParseFloat(datas[1], 64)

			if err1 != nil || err2 != nil {
				util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse GPS_Precision data, %s; %s",
					err1.Error(), err2.Error())
				return nil
			}
		}
	}
	return newGps
}

func readAndUpdateGpsInfo() {
	for {
		ccc := exec.Command("/bin/bash", "-c", "cat /home/pi/MaoTemp/monitorAPI.html")
		gpsData, err := ccc.CombinedOutput()
		if err == nil {
			gpsLast = updateGpsInfo(string(gpsData))
		} else {
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to get GPS API data, %s", err.Error())
		}

		time.Sleep(1 * time.Second)
	}
}

func supportRttMeasure(rttStreamClient pb.MaoServerDiscovery_RttMeasureClient, silent bool) {
	util.MaoLogM(util.INFO, c_MODULE_NAME, "Enable RTT measure feature.")
	for {
		// After sending echo response, we have some time (measure interval) to get latest hostname
		hostname, err := util.GetHostname()
		if err != nil {
			hostname = "Mao-Unknown"
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to get hostname for RTT measure, %s", err.Error())
		}

		rttEchoRequest, err := rttStreamClient.Recv()
		if err != nil {
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to receive RTT echo request, %s", err.Error())
			return
		}
		err = rttStreamClient.Send(&pb.RttEchoResponse{Ack: rttEchoRequest.GetSeq(), Hostname: hostname})
		if err != nil {
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to send RTT echo request, %s", err.Error())
			return
		}
		if silent == false {
			util.MaoLogM(util.INFO, c_MODULE_NAME, "RTT Measure: sent echo response with ack %d", rttEchoRequest.GetSeq())
		}
	}
}

func RunGeneralClient(report_server_addr *net.IP, report_server_port uint32, report_interval uint32, silent bool,
	influxdbUrl string, influxdbOrgBucket string, influxdbToken string,
	nat66Gateway bool, nat66Persistent bool,
	gpsMonitor bool, gpsPersistent bool,
	envTempMonitor bool, envTempPersistent bool,
	minLogLevel util.MaoLogLevel) {

	util.InitMaoLog(minLogLevel)

	var influxdbClient influxdb2.Client
	var influxdbWriteAPI influxdb2Api.WriteAPI
	if nat66Persistent || gpsPersistent || envTempPersistent {
		util.MaoLogM(util.INFO, c_MODULE_NAME, "Initiate influxdb client ...")
		influxdbClient = influxdb2.NewClient(influxdbUrl, influxdbToken)
		defer influxdbClient.Close()
		influxdbWriteAPI = influxdbClient.WriteAPI(influxdbOrgBucket, influxdbOrgBucket)
	}
	if envTempMonitor {
		go updateEnvironmentTemperature()
	}
	if gpsMonitor {
		go readAndUpdateGpsInfo()
	}

	util.MaoLogM(util.INFO, c_MODULE_NAME, "Connect to center ...")
	for {
		serverAddr := util.GetAddrPort(report_server_addr, report_server_port)
		util.MaoLogM(util.INFO, c_MODULE_NAME, "Connect to %s ...", serverAddr)

		ctx, cancelCtx := context.WithTimeout(context.Background(), 3 * time.Second)
		connect, err := grpc.DialContext(ctx, serverAddr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Retry, %s ...", err.Error())
			continue
		}
		cancelCtx()
		util.MaoLogM(util.INFO, c_MODULE_NAME, "Connected.")

		client := pb.NewMaoServerDiscoveryClient(connect)


		clientCommonContext, cancelCommonContext := context.WithCancel(context.Background())
		rttStreamClient, err := client.RttMeasure(clientCommonContext)
		if err != nil {
			util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to get rttStreamClient, %s", err.Error())
			cancelCommonContext()
			continue
		}
		go supportRttMeasure(rttStreamClient, silent)


		reportStreamClient, err := client.Report(clientCommonContext)
		if err != nil {
			util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to get reportStreamClient, %s", err.Error())
			cancelCommonContext()
			continue
		}
		util.MaoLogM(util.INFO, c_MODULE_NAME, "Got reportStreamClient.")

		count := 1
		for {
			dataOk := true
			hostname, err := util.GetHostname()
			if err != nil {
				hostname = "Mao-Unknown"
				dataOk = false
			}

			ips, err := util.GetUnicastIp()
			if err != nil {
				ips = []string{"Mao-Fail", err.Error()}
				dataOk = false
			}

			util.MaoLogM(util.DEBUG, c_MODULE_NAME, "%d: To send", count)
			report := &pb.ServerReport{
				Ok:          dataOk,
				Hostname:    hostname,
				Ips:         ips,
				NowDatetime: time.Now().String(),
				AuxData: "",
			}

			auxDataMap := make(map[string]interface{})

			if nat66Gateway {
				v6In, v6Out, err := getNat66GatewayData()
				if err == nil {
					auxDataMap["v6In"] = v6In
					auxDataMap["v6Out"] = v6Out
					if nat66Persistent {
						nat66UploadInfluxdb(&influxdbWriteAPI, v6In, v6Out)
					}
				}
			}
			if envTempMonitor {
				env := envTemp
				auxDataMap["envTemp"] = env
				auxDataMap["envGeo"] = "Beijing-HQ"
				auxDataMap["envTime"] = time.Now().Format(time.RFC3339Nano) // RFC3339Nano, most precise format
				if envTempPersistent {
					envTempUploadInfluxdb(&influxdbWriteAPI, env)
				}
			}
			if gpsMonitor {
				gpsNow := gpsLast
				if gpsNow != nil {
					auxDataMap["GPS_Timestamp"] = gpsNow.Timestamp
					auxDataMap["GPS_Latitude"] = gpsNow.Latitude
					auxDataMap["GPS_Longitude"] = gpsNow.Longitude
					auxDataMap["GPS_Altitude"] = gpsNow.Altitude
					auxDataMap["GPS_Satellite"] = gpsNow.Satellite
					auxDataMap["GPS_Hdop"] = gpsNow.Hdop
					auxDataMap["GPS_Vdop"] = gpsNow.Vdop

					if gpsPersistent {
						gpsDataUploadInfluxdb(&influxdbWriteAPI, gpsNow)
					}
				}
			}

			auxDataByte, err := json.Marshal(auxDataMap)
			if err != nil {
				util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to marshal auxDataMap to json format, %s", err.Error())
			} else {
				report.AuxData = string(auxDataByte)
			}

			err = reportStreamClient.Send(report)
			if err != nil {
				util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to report, %s", err.Error())
				break
			}
			if silent == false {
				util.MaoLogM(util.INFO, c_MODULE_NAME, "ServerReport - %v", report)
			}
			util.MaoLogM(util.DEBUG, c_MODULE_NAME, "%d: Sent", count)

			count++
			time.Sleep(time.Duration(report_interval) * time.Millisecond)
		}
		cancelCommonContext()
		time.Sleep(1 * time.Second)
	}
}
