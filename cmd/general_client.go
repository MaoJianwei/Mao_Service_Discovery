package branch

import (
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
)

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
					util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse 1-line protocol data slice, %s", err.Error())
				}
			} else {
				util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to parse 1-line protocol data, %s", err.Error())
			}
		} else {
			util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to get 1-line protocol data, %s", err.Error())
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func RunGeneralClient(report_server_addr *net.IP, report_server_port uint32, report_interval uint32, silent bool,
	nat66Gateway bool, nat66Persistent bool, influxdbUrl string, influxdbOrgBucket string, influxdbToken string,
	envTempMonitor bool, envTempPersistent bool, minLogLevel util.MaoLogLevel) {

	util.InitMaoLog(minLogLevel)

	var influxdbClient influxdb2.Client
	var influxdbWriteAPI influxdb2Api.WriteAPI
	if nat66Persistent {
		util.MaoLogM(util.INFO, c_MODULE_NAME, "Initiate influxdb client ...")
		influxdbClient = influxdb2.NewClient(influxdbUrl, influxdbToken)
		defer influxdbClient.Close()
		influxdbWriteAPI = influxdbClient.WriteAPI(influxdbOrgBucket, influxdbOrgBucket)
	}
	if envTempMonitor {
		go updateEnvironmentTemperature()
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
		streamClient, err := client.Report(context.Background())
		if err != nil {
			util.MaoLogM(util.ERROR, c_MODULE_NAME, "Fail to get streamClient, %s", err.Error())
			continue
		}
		util.MaoLogM(util.INFO, c_MODULE_NAME, "Got StreamClient.")

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
				if envTempPersistent {
					envTempUploadInfluxdb(&influxdbWriteAPI, env)
				}
			}

			auxDataByte, err := json.Marshal(auxDataMap)
			if err != nil {
				util.MaoLogM(util.WARN, c_MODULE_NAME, "Fail to marshal auxDataMap to json format, %s", err.Error())
			} else {
				report.AuxData = string(auxDataByte)
			}

			err = streamClient.Send(report)
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
		time.Sleep(1 * time.Second)
	}
}
