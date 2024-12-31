package branch

import (
	"MaoServerDiscovery/cmd/api/data"
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	"MaoServerDiscovery/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	c2_MODULE_NAME = "General-Client-V2"
	INVALID_ENV_TEMP = -10000
)

type GeneralClientV2 struct {

	// for influxdb persistent
	envTempDataChannel chan *data.Temperature
	gpsDataChannel chan *data.GpsData
	nat66DataChannel chan *data.Nat66

	// for gRPC upload report
	envTempLast *data.Temperature
	gpsLast *data.GpsData
	nat66Last *data.Nat66
}



func (c *GeneralClientV2) parseGpsInfo(gpsApiData string) (*data.GpsData, error) {

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
				return nil, errors.New("gps location is not ready")
			}

			newGps.Latitude, err1 = strconv.ParseFloat(datas[0], 64)
			newGps.Longitude, err2 = strconv.ParseFloat(datas[1], 64)
			newGps.Altitude, err3 = strconv.ParseFloat(datas[2], 64)
			newGps.Satellite, err4 = strconv.ParseUint(datas[3], 10, 64)

			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to parse GPS data, %s; %s; %s; %s",
					err1.Error(), err2.Error(), err3.Error(), err4.Error())
				return nil, errors.New(fmt.Sprintf("Fail to parse GPS data, %s; %s; %s; %s",
					err1.Error(), err2.Error(), err3.Error(), err4.Error()))
			}

		case "GpsTime":
			newGps.Timestamp = kv[1]

		case "GPS_Precision":
			datas := strings.Split(kv[1], ",")
			newGps.Hdop, err1 = strconv.ParseFloat(datas[0], 64)
			newGps.Vdop, err2 = strconv.ParseFloat(datas[1], 64)

			if err1 != nil || err2 != nil {
				util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to parse GPS_Precision data, %s; %s",
					err1.Error(), err2.Error())
				return nil, errors.New(fmt.Sprintf("Fail to parse GPS_Precision data, %s; %s",
					err1.Error(), err2.Error()))
			}
		}
	}
	return newGps, nil
}

func (c *GeneralClientV2) getGpsInfo() (*data.GpsData, error) {
	ccc := exec.Command("/bin/bash", "-c", "cat /home/pi/MaoTemp/monitorAPI.html")
	gpsDataStr, err := ccc.CombinedOutput()
	if err != nil {
		util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to get GPS API data, %s", err.Error())
		return nil, err
	}
	return c.parseGpsInfo(string(gpsDataStr))
}

func (c *GeneralClientV2) gpsProcessor(gpsPersistent bool) {
	var epoch uint32 = 1
	for {
		time.Sleep(1 * time.Second)

		gpsData, err := c.getGpsInfo()
		if err != nil {
			continue
		}

		gpsData.Epoch = epoch
		epoch++

		if gpsPersistent {
			c.gpsDataChannel <- gpsData
		}
		c.gpsLast = gpsData
	}
}



func (c *GeneralClientV2) getEnvironmentTemperature() (float64, error) {
	ccc := exec.Command("/bin/bash", "-c", "cat /sys/bus/w1/devices/28-00141093caff/w1_slave")
	w1Data, err := ccc.CombinedOutput()
	if err == nil {

		w1DataSplit := strings.Split(string(w1Data), "\n")
		if len(w1DataSplit) == 3 {

			tempText := strings.Split(w1DataSplit[1], "=")
			if len(tempText) == 2 {

				temp, err := strconv.ParseFloat(tempText[1], 64)
				if err == nil {
					util.MaoLogM(util.DEBUG, c2_MODULE_NAME, "Get envTemp: %f, %f", temp, temp / 1000)
					return temp / 1000, err
				} else {
					util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to parse temperature text, %s", err.Error())
				}
			} else {
				util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to parse 1-line protocol data slice, the number of elements is not 2, in fact %d", len(tempText))
				err = errors.New(fmt.Sprintf("Fail to parse 1-line protocol data slice, the number of elements is not 2, in fact %d", len(tempText)))
			}
		} else {
			util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to parse 1-line protocol data, the lines of the result text is not 3, in fact %d.", len(w1DataSplit))
			err = errors.New(fmt.Sprintf("Fail to parse 1-line protocol data, the lines of the result text is not 3, in fact %d", len(w1DataSplit)))
		}
	} else {
		util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to get 1-line protocol data, %s", err.Error())
	}
	return INVALID_ENV_TEMP, err
}

func (c *GeneralClientV2) envTempProcessor(envTempPersistent bool) {
	var epoch uint32 = 1
	for {
		time.Sleep(500 * time.Millisecond)

		envTempData, err := c.getEnvironmentTemperature()
		if err != nil {
			continue
		}

		envTemperature := &data.Temperature{
			Epoch:       epoch,
			Temperature: envTempData,
		}
		epoch++

		if envTempPersistent {
			c.envTempDataChannel <- envTemperature
		}
		c.envTempLast = envTemperature
	}
}



/*
   return v6In,v6Out,error
*/
func (c *GeneralClientV2) getNat66GatewayData() (uint64, uint64, error) {
	ccc := exec.Command("/bin/bash", "-c", "ip6tables -nvxL FORWARD | grep MaoIPv6In | awk '{printf $2}'")
	ininin, err := ccc.CombinedOutput()
	if err != nil {
		util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to get MaoIPv6In, %s", err.Error())
		return 0, 0, err
	}
	ccc = exec.Command("/bin/bash", "-c", "ip6tables -nvxL FORWARD | grep MaoIPv6Out | awk '{printf $2}'")
	outoutout, err := ccc.CombinedOutput()
	if err != nil {
		util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to get MaoIPv6Out, %s", err.Error())
		return 0, 0, err
	}
	v6In, err := strconv.ParseUint(string(ininin), 10, 64)
	if err != nil {
		util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to parse MaoIPv6In, %s", err.Error())
		return 0, 0, err
	}
	v6Out, err := strconv.ParseUint(string(outoutout), 10, 64)
	if err != nil {
		util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to parse MaoIPv6Out, %s", err.Error())
		return 0, 0, err
	}
	util.MaoLogM(util.DEBUG, c2_MODULE_NAME, "v6In: %d , v6Out: %d", v6In, v6Out)
	return v6In, v6Out, nil
}

func (c *GeneralClientV2) nat66Processor(nat66Persistent bool) {
	var epoch uint32 = 1
	for {
		time.Sleep(1 * time.Second)

		v6In, v6Out, err := c.getNat66GatewayData()
		if err != nil {
			continue
		}

		nat66Data := &data.Nat66{
			Epoch:   epoch,
			IPv6In:  v6In,
			IPv6Out: v6Out,
		}
		epoch++

		if nat66Persistent {
			c.nat66DataChannel <- nat66Data
		}
		c.nat66Last = nat66Data
	}
}



func (c *GeneralClientV2) envTempUploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, envTemperature float64) {
	// write point asynchronously
	(*writeAPI).WritePoint(
		influxdb2.NewPointWithMeasurement("Temperature").
			AddTag("Geo", "Beijing-HQ").
			AddField("env", envTemperature).
			SetTime(time.Now()))
	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
	//(*writeAPI).Flush()
}

func (c *GeneralClientV2) gpsDataUploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, gpsData *data.GpsData) {
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

func (c *GeneralClientV2) nat66UploadInfluxdb(writeAPI *influxdb2Api.WriteAPI, v6In uint64, v6Out uint64) {
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

func (c *GeneralClientV2) influxdbPersistentProcessor(influxdbUrl string, influxdbOrgBucket string, influxdbToken string,
	nat66Persistent bool, gpsPersistent bool, envTempPersistent bool) {

	if !nat66Persistent && !gpsPersistent && !envTempPersistent {
		return
	}

	util.MaoLogM(util.INFO, c2_MODULE_NAME, "Initiate influxdb client ...")
	var influxdbClient influxdb2.Client
	var influxdbWriteAPI influxdb2Api.WriteAPI
	influxdbClient = influxdb2.NewClient(influxdbUrl, influxdbToken)
	influxdbWriteAPI = influxdbClient.WriteAPI(influxdbOrgBucket, influxdbOrgBucket)
	defer influxdbClient.Close()

	for {
		select {
		case envTempData := <- c.envTempDataChannel:
			c.envTempUploadInfluxdb(&influxdbWriteAPI, envTempData.Temperature)
		case gpsData := <- c.gpsDataChannel:
			c.gpsDataUploadInfluxdb(&influxdbWriteAPI, gpsData)
		case nat66Data := <- c.nat66DataChannel:
			c.nat66UploadInfluxdb(&influxdbWriteAPI, nat66Data.IPv6In, nat66Data.IPv6Out)

		// todo: case shutdown signal.
		}
	}
}

func (c *GeneralClientV2) grpcRttMeasureProcessor(rttStreamClient pb.MaoServerDiscovery_RttMeasureClient, silent bool) {
	util.MaoLogM(util.INFO, c2_MODULE_NAME, "Enable RTT measure feature.")
	for {
		// After sending echo response, we have some time (measure interval) to get latest hostname
		hostname, err := util.GetHostname()
		if err != nil {
			hostname = "Mao-Unknown"
			util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to get hostname for RTT measure, %s", err.Error())
		}

		rttEchoRequest, err := rttStreamClient.Recv()
		if err != nil {
			util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to receive RTT echo request, %s", err.Error())
			return
		}
		err = rttStreamClient.Send(&pb.RttEchoResponse{Ack: rttEchoRequest.GetSeq(), Hostname: hostname})
		if err != nil {
			util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to send RTT echo request, %s", err.Error())
			return
		}
		if silent == false {
			util.MaoLogM(util.INFO, c2_MODULE_NAME, "RTT Measure: sent echo response with ack %d", rttEchoRequest.GetSeq())
		}
	}
}

func (c *GeneralClientV2) gRpcProcessor(
	reportServerAddr *net.IP, reportServerPort uint32, reportInterval uint32, silent bool,
	nat66Gateway bool, gpsMonitor bool, envTempMonitor bool ) {
	for {
		time.Sleep(1 * time.Second)

		serverAddr := util.GetAddrPort(reportServerAddr, reportServerPort)
		util.MaoLogM(util.INFO, c2_MODULE_NAME, "Connect to %s ...", serverAddr)

		ctx, cancelCtx := context.WithTimeout(context.Background(), 3 * time.Second)
		connect, err := grpc.DialContext(ctx, serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			util.MaoLogM(util.WARN, c2_MODULE_NAME, "Retry, %s ...", err.Error())
			continue
		}
		cancelCtx()
		util.MaoLogM(util.INFO, c2_MODULE_NAME, "Connected.")

		client := pb.NewMaoServerDiscoveryClient(connect)


		clientCommonContext, cancelCommonContext := context.WithCancel(context.Background())
		rttStreamClient, err := client.RttMeasure(clientCommonContext)
		if err != nil {
			util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to get rttStreamClient, %s", err.Error())
			cancelCommonContext()
			continue
		}
		go c.grpcRttMeasureProcessor(rttStreamClient, silent)


		reportStreamClient, err := client.Report(clientCommonContext)
		if err != nil {
			util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to get reportStreamClient, %s", err.Error())
			cancelCommonContext()
			continue
		}
		util.MaoLogM(util.INFO, c2_MODULE_NAME, "Got reportStreamClient.")

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

			util.MaoLogM(util.DEBUG, c2_MODULE_NAME, "%d: To send", count)
			report := &pb.ServerReport{
				Ok:          dataOk,
				Hostname:    hostname,
				Ips:         ips,
				NowDatetime: time.Now().String(),
				AuxData: "",
			}

			auxDataMap := make(map[string]interface{})

			if nat66Gateway {
				nat66Now := c.nat66Last
				//c.nat66Last = nil
				if nat66Now != nil {
					auxDataMap["NAT66_Epoch"] = nat66Now.Epoch
					auxDataMap["NAT66_v6In"] = nat66Now.IPv6In
					auxDataMap["NAT66_v6Out"] = nat66Now.IPv6Out
				}
			}
			if envTempMonitor {
				envTempNow := c.envTempLast
				//c.envTempLast = INVALID_ENV_TEMP
				if envTempNow != nil && envTempNow.Temperature > INVALID_ENV_TEMP + 100 {
					auxDataMap["Env_Temp_Epoch"] = envTempNow.Epoch
					auxDataMap["Env_Temp"] = envTempNow.Temperature
					auxDataMap["Env_Geo"] = "Beijing-HQ"
					auxDataMap["Env_Time"] = time.Now().Format(time.RFC3339Nano) // RFC3339Nano, most precise format
				}
			}
			if gpsMonitor {
				gpsNow := c.gpsLast
				//c.gpsLast = nil
				if gpsNow != nil {
					auxDataMap["GPS_Epoch"] = gpsNow.Epoch
					auxDataMap["GPS_Timestamp"] = gpsNow.Timestamp
					auxDataMap["GPS_Latitude"] = gpsNow.Latitude
					auxDataMap["GPS_Longitude"] = gpsNow.Longitude
					auxDataMap["GPS_Altitude"] = gpsNow.Altitude
					auxDataMap["GPS_Satellite"] = gpsNow.Satellite
					auxDataMap["GPS_Hdop"] = gpsNow.Hdop
					auxDataMap["GPS_Vdop"] = gpsNow.Vdop
				}
			}

			auxDataByte, err := json.Marshal(auxDataMap)
			if err != nil {
				util.MaoLogM(util.WARN, c2_MODULE_NAME, "Fail to marshal auxDataMap to json format, %s", err.Error())
			} else {
				report.AuxData = string(auxDataByte)
			}

			err = reportStreamClient.Send(report)
			if err != nil {
				util.MaoLogM(util.ERROR, c2_MODULE_NAME, "Fail to report, %s", err.Error())
				break
			}
			if silent == false {
				util.MaoLogM(util.INFO, c2_MODULE_NAME, "ServerReport - %v", report)
			}
			util.MaoLogM(util.DEBUG, c2_MODULE_NAME, "%d: Sent", count)

			count++
			time.Sleep(time.Duration(reportInterval) * time.Millisecond)
		}
		cancelCommonContext()
	}
}

func (c *GeneralClientV2) Run(reportServerAddr *net.IP, reportServerPort uint32, reportInterval uint32, silent bool,
	influxdbUrl string, influxdbOrgBucket string, influxdbToken string,
	nat66Gateway bool, nat66Persistent bool,
	gpsMonitor bool, gpsPersistent bool,
	envTempMonitor bool, envTempPersistent bool,
	minLogLevel util.MaoLogLevel) {

	util.InitMaoLog(minLogLevel)

	c.envTempLast = nil
	c.gpsLast = nil
	c.nat66Last = nil

	c.envTempDataChannel = make(chan *data.Temperature, 1024)
	c.gpsDataChannel = make(chan *data.GpsData, 1024)
	c.nat66DataChannel = make(chan *data.Nat66, 1024)

	go c.influxdbPersistentProcessor(influxdbUrl, influxdbOrgBucket, influxdbToken,
		nat66Persistent, gpsPersistent, envTempPersistent)

	if gpsMonitor {
		go c.gpsProcessor(gpsPersistent)
	}
	if nat66Gateway {
		go c.nat66Processor(nat66Persistent)
	}
	if envTempMonitor {
		go c.envTempProcessor(envTempPersistent)
	}

	c.gRpcProcessor(reportServerAddr, reportServerPort, reportInterval, silent,
		nat66Gateway, gpsMonitor, envTempMonitor)
}
