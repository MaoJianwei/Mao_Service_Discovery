package Soap

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/InfluxDB"
	"MaoServerDiscovery/util"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

const (
	p_TPLINK_MODULE_NAME = "TPLINK-Gateway-module"

	FLAG_GATEWAY_BytesReceivedSpeed = 1 << 0
	FLAG_GATEWAY_BytesSentSpeed = 1 << 1
	FLAG_GATEWAY_PacketsReceivedSpeed = 1 << 2
	FLAG_GATEWAY_PacketsSentSpeed = 1 << 3
	FLAG_GATEWAY_Uptime = 1 << 4

	INT32_MAX_VALUE = 0x100000000
)

type TplinkGatewayModule struct {

	lastseen_BytesReceived_Accumulated_Baseline uint64
	lastseen_BytesReceived uint64
	lastseen_BytesReceived_timestamp time.Time

	lastseen_BytesSent_Accumulated_Baseline uint64
	lastseen_BytesSent uint64
	lastseen_BytesSent_timestamp time.Time

	lastseen_PacketsReceived_Accumulated_Baseline uint64
	lastseen_PacketsReceived uint64
	lastseen_PacketsReceived_timestamp time.Time

	lastseen_PacketsSent_Accumulated_Baseline uint64
	lastseen_PacketsSent uint64
	lastseen_PacketsSent_timestamp time.Time

	lastseen_Uptime_Accumulated_Baseline uint64
	lastseen_Uptime uint64
	lastseen_Uptime_timestamp time.Time

	BytesReceivedSpeed uint64
	BytesSentSpeed uint64
	PacketsReceivedSpeed uint64
	PacketsSentSpeed uint64
	Uptime uint64
}

func (t *TplinkGatewayModule) publishInfluxDB(writeAPI *influxdb2Api.WriteAPI, finishFlag uint) {
	// write point asynchronously

	if finishFlag & FLAG_GATEWAY_BytesReceivedSpeed != 0 {
		(*writeAPI).WritePoint(
			influxdb2.NewPointWithMeasurement(MaoApi.GATEWAY_MEASUREMENT).
				AddTag(MaoApi.GATEWAY_TAG_GEO, "Beijing-HQ").
				AddField(MaoApi.GATEWAY_FIELD_BytesReceivedSpeed, t.BytesReceivedSpeed).
				AddField(MaoApi.GATEWAY_FIELD_BytesReceived, t.lastseen_BytesReceived_Accumulated_Baseline + t.lastseen_BytesReceived).
				SetTime(t.lastseen_BytesReceived_timestamp))
	}
	if finishFlag & FLAG_GATEWAY_BytesSentSpeed != 0 {
		(*writeAPI).WritePoint(
			influxdb2.NewPointWithMeasurement(MaoApi.GATEWAY_MEASUREMENT).
				AddTag(MaoApi.GATEWAY_TAG_GEO, "Beijing-HQ").
				AddField(MaoApi.GATEWAY_FIELD_BytesSentSpeed, t.BytesSentSpeed).
				AddField(MaoApi.GATEWAY_FIELD_BytesSent, t.lastseen_BytesSent_Accumulated_Baseline + t.lastseen_BytesSent).
				SetTime(t.lastseen_BytesSent_timestamp))
	}
	if finishFlag & FLAG_GATEWAY_PacketsReceivedSpeed != 0 {
		(*writeAPI).WritePoint(
			influxdb2.NewPointWithMeasurement(MaoApi.GATEWAY_MEASUREMENT).
				AddTag(MaoApi.GATEWAY_TAG_GEO, "Beijing-HQ").
				AddField(MaoApi.GATEWAY_FIELD_PacketsReceivedSpeed, t.PacketsReceivedSpeed).
				AddField(MaoApi.GATEWAY_FIELD_PacketsReceived, t.lastseen_PacketsReceived_Accumulated_Baseline + t.lastseen_PacketsReceived).
				SetTime(t.lastseen_PacketsReceived_timestamp))
	}
	if finishFlag & FLAG_GATEWAY_PacketsSentSpeed != 0 {
		(*writeAPI).WritePoint(
			influxdb2.NewPointWithMeasurement(MaoApi.GATEWAY_MEASUREMENT).
				AddTag(MaoApi.GATEWAY_TAG_GEO, "Beijing-HQ").
				AddField(MaoApi.GATEWAY_FIELD_PacketsSentSpeed, t.PacketsSentSpeed).
				AddField(MaoApi.GATEWAY_FIELD_PacketsSent, t.lastseen_PacketsSent_Accumulated_Baseline + t.lastseen_PacketsSent).
				SetTime(t.lastseen_PacketsSent_timestamp))
	}
	if finishFlag & FLAG_GATEWAY_Uptime != 0 {
		(*writeAPI).WritePoint(
			influxdb2.NewPointWithMeasurement(MaoApi.GATEWAY_MEASUREMENT).
				AddTag(MaoApi.GATEWAY_TAG_GEO, "Beijing-HQ").
				AddField(MaoApi.GATEWAY_FIELD_Uptime, t.Uptime).
				SetTime(t.lastseen_Uptime_timestamp))
	}

	// Not flush writes, avoid blocking my thread, then the lib's thread will block itself.
	//(*writeAPI).Flush()
}

func (t *TplinkGatewayModule) pushLoop(triggerChannel *chan uint) {
	var client *influxdb2.Client
	var writeApi *influxdb2Api.WriteAPI

	for {
		// do...while, wait for being well configured.
		client, writeApi = InfluxDB.CreateClientAndWriteAPI()
		if writeApi != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	defer (*client).Close()

	for finishFlag := range *triggerChannel {
		t.publishInfluxDB(writeApi, finishFlag)
	}
}

func (t *TplinkGatewayModule) controlLoop(triggerChannel *chan uint) {
	for {
		time.Sleep(2 * time.Second)
		var finishFlag uint = 0

		newBytesReceived, err := GetTotalBytesReceived()
		newBytesReceived_timestamp := time.Now()
		if err == nil {
			// statistic may overflow (rollback to 0)
			if newBytesReceived >= t.lastseen_BytesReceived {
				t.BytesReceivedSpeed = uint64(float64(newBytesReceived - t.lastseen_BytesReceived) / (newBytesReceived_timestamp.Sub(t.lastseen_BytesReceived_timestamp).Seconds()))
			} else {
				t.BytesReceivedSpeed = uint64(float64(newBytesReceived + (INT32_MAX_VALUE- t.lastseen_BytesReceived)) / (newBytesReceived_timestamp.Sub(t.lastseen_BytesReceived_timestamp).Seconds()))
				t.lastseen_BytesReceived_Accumulated_Baseline += INT32_MAX_VALUE
			}
			//log.Printf("%d, %f, %f", t.BytesReceivedSpeed, float64(newBytesReceived - t.lastseen_BytesReceived), newBytesReceived_timestamp.Sub(t.lastseen_BytesReceived_timestamp).Seconds())
			t.lastseen_BytesReceived = newBytesReceived
			t.lastseen_BytesReceived_timestamp = newBytesReceived_timestamp
			finishFlag |= FLAG_GATEWAY_BytesReceivedSpeed
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalBytesReceived, %s", err.Error())
		}

		newBytesSent, err := GetTotalBytesSent()
		newBytesSent_timestamp := time.Now()
		if err == nil {
			// statistic may overflow (rollback to 0)
			if newBytesSent >= t.lastseen_BytesSent {
				t.BytesSentSpeed = uint64(float64(newBytesSent - t.lastseen_BytesSent) / (newBytesSent_timestamp.Sub(t.lastseen_BytesSent_timestamp).Seconds()))
			} else {
				t.BytesSentSpeed = uint64(float64(newBytesSent + (INT32_MAX_VALUE - t.lastseen_BytesSent)) / (newBytesSent_timestamp.Sub(t.lastseen_BytesSent_timestamp).Seconds()))
				t.lastseen_BytesSent_Accumulated_Baseline += INT32_MAX_VALUE
			}
			t.lastseen_BytesSent = newBytesSent
			t.lastseen_BytesSent_timestamp = newBytesSent_timestamp
			finishFlag |= FLAG_GATEWAY_BytesSentSpeed
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalBytesSent, %s", err.Error())
		}

		newPacketsReceived, err := GetTotalPacketsReceived()
		newPacketsReceived_timestamp := time.Now()
		if err == nil {
			// statistic may overflow (rollback to 0)
			if newPacketsReceived >= t.lastseen_PacketsReceived {
				t.PacketsReceivedSpeed = uint64(float64(newPacketsReceived - t.lastseen_PacketsReceived) / (newPacketsReceived_timestamp.Sub(t.lastseen_PacketsReceived_timestamp).Seconds()))
			} else {
				t.PacketsReceivedSpeed = uint64(float64(newPacketsReceived + (INT32_MAX_VALUE - t.lastseen_PacketsReceived)) / (newPacketsReceived_timestamp.Sub(t.lastseen_PacketsReceived_timestamp).Seconds()))
				t.lastseen_PacketsReceived_Accumulated_Baseline += INT32_MAX_VALUE
			}
			t.lastseen_PacketsReceived = newPacketsReceived
			t.lastseen_PacketsReceived_timestamp = newPacketsReceived_timestamp
			finishFlag |= FLAG_GATEWAY_PacketsReceivedSpeed
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalPacketsReceived, %s", err.Error())
		}

		newPacketsSent, err := GetTotalPacketsSent()
		newPacketsSent_timestamp := time.Now()
		if err == nil {
			// statistic may overflow (rollback to 0)
			if newPacketsSent >= t.lastseen_PacketsSent {
				t.PacketsSentSpeed = uint64(float64(newPacketsSent - t.lastseen_PacketsSent) / (newPacketsSent_timestamp.Sub(t.lastseen_PacketsSent_timestamp).Seconds()))
			} else {
				t.PacketsSentSpeed = uint64(float64(newPacketsSent + (INT32_MAX_VALUE - t.lastseen_PacketsSent)) / (newPacketsSent_timestamp.Sub(t.lastseen_PacketsSent_timestamp).Seconds()))
				t.lastseen_PacketsSent_Accumulated_Baseline += INT32_MAX_VALUE
			}
			t.lastseen_PacketsSent = newPacketsSent
			t.lastseen_PacketsSent_timestamp = newPacketsSent_timestamp
			finishFlag |= FLAG_GATEWAY_PacketsSentSpeed
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalPacketsSent, %s", err.Error())
		}

		newUptime, err := GetUptime()
		newUptime_timestamp := time.Now()
		if err == nil {
			if newUptime < t.lastseen_Uptime {
				t.lastseen_Uptime_Accumulated_Baseline += INT32_MAX_VALUE
			}
			t.Uptime = t.lastseen_Uptime_Accumulated_Baseline + newUptime
			t.lastseen_Uptime = newUptime
			t.lastseen_Uptime_timestamp = newUptime_timestamp
			finishFlag |= FLAG_GATEWAY_Uptime
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get Uptime, %s", err.Error())
		}

		*triggerChannel <- finishFlag

		util.MaoLogM(util.DEBUG, p_TPLINK_MODULE_NAME, "BytesSentSpeed: %d, BytesReceivedSpeed: %d, PacketsSentSpeed: %d, PacketsReceivedSpeed: %d, Uptime: %d",
			t.BytesSentSpeed, t.BytesReceivedSpeed, t.PacketsSentSpeed, t.PacketsReceivedSpeed, t.Uptime)
	}
}

func (t *TplinkGatewayModule) InitTplinkGatewayModule() bool {
	triggerChannel := make(chan uint, 100)
	go t.controlLoop(&triggerChannel)
	go t.pushLoop(&triggerChannel)
	return true
}