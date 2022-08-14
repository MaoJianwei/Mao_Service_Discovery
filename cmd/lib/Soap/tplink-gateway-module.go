package Soap

import (
	"MaoServerDiscovery/util"
	"time"
)

const (
	p_TPLINK_MODULE_NAME = "TPLINK-Gateway-module"
)

type TplinkGatewayModule struct {

	lastseen_BytesReceived uint64
	lastseen_BytesReceived_timestamp time.Time

	lastseen_BytesSent uint64
	lastseen_BytesSent_timestamp time.Time

	lastseen_PacketsReceived uint64
	lastseen_PacketsReceived_timestamp time.Time

	lastseen_PacketsSent uint64
	lastseen_PacketsSent_timestamp time.Time

	BytesReceivedSpeed uint64
	BytesSentSpeed uint64
	PacketsReceivedSpeed uint64
	PacketsSentSpeed uint64
	Uptime uint64
}

func (t *TplinkGatewayModule) controlLoop() {
	for {
		time.Sleep(2 * time.Second)
		newBytesReceived, err := GetTotalBytesReceived()
		newBytesReceived_timestamp := time.Now()
		if err == nil {
			if newBytesReceived >= t.lastseen_BytesReceived {
				t.BytesReceivedSpeed = uint64(float64(newBytesReceived - t.lastseen_BytesReceived) / (newBytesReceived_timestamp.Sub(t.lastseen_BytesReceived_timestamp).Seconds()))
			} else {
				t.BytesReceivedSpeed = uint64(float64(newBytesReceived) / (newBytesReceived_timestamp.Sub(t.lastseen_BytesReceived_timestamp).Seconds())) // statistic may overflow (rollback to 0)
			}
			//log.Printf("%d, %f, %f", t.BytesReceivedSpeed, float64(newBytesReceived - t.lastseen_BytesReceived), newBytesReceived_timestamp.Sub(t.lastseen_BytesReceived_timestamp).Seconds())
			t.lastseen_BytesReceived = newBytesReceived // statistic may overflow (rollback to 0)
			t.lastseen_BytesReceived_timestamp = newBytesReceived_timestamp
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalBytesReceived, %s", err.Error())
		}

		newBytesSent, err := GetTotalBytesSent()
		newBytesSent_timestamp := time.Now()
		if err == nil {
			if newBytesSent >= t.lastseen_BytesSent {
				t.BytesSentSpeed = uint64(float64(newBytesSent - t.lastseen_BytesSent) / (newBytesSent_timestamp.Sub(t.lastseen_BytesSent_timestamp).Seconds()))
			} else {
				t.BytesSentSpeed = uint64(float64(newBytesSent) / (newBytesSent_timestamp.Sub(t.lastseen_BytesSent_timestamp).Seconds())) // statistic may overflow (rollback to 0)
			}
			t.lastseen_BytesSent = newBytesSent // statistic may overflow (rollback to 0)
			t.lastseen_BytesSent_timestamp = newBytesSent_timestamp

		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalBytesSent, %s", err.Error())
		}

		newPacketsReceived, err := GetTotalPacketsReceived()
		newPacketsReceived_timestamp := time.Now()
		if err == nil {
			if newPacketsReceived >= t.lastseen_PacketsReceived {
				t.PacketsReceivedSpeed = uint64(float64(newPacketsReceived - t.lastseen_PacketsReceived) / (newPacketsReceived_timestamp.Sub(t.lastseen_PacketsReceived_timestamp).Seconds()))
			} else {
				t.PacketsReceivedSpeed = uint64(float64(newPacketsReceived) / (newPacketsReceived_timestamp.Sub(t.lastseen_PacketsReceived_timestamp).Seconds())) // statistic may overflow (rollback to 0)
			}
			t.lastseen_PacketsReceived = newPacketsReceived // statistic may overflow (rollback to 0)
			t.lastseen_PacketsReceived_timestamp = newPacketsReceived_timestamp
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalPacketsReceived, %s", err.Error())
		}

		newPacketsSent, err := GetTotalPacketsSent()
		newPacketsSent_timestamp := time.Now()
		if err == nil {
			if newPacketsSent >= t.lastseen_PacketsSent {
				t.PacketsSentSpeed = uint64(float64(newPacketsSent - t.lastseen_PacketsSent) / (newPacketsSent_timestamp.Sub(t.lastseen_PacketsSent_timestamp).Seconds()))
			} else {
				t.PacketsSentSpeed = uint64(float64(newPacketsSent) / (newPacketsSent_timestamp.Sub(t.lastseen_PacketsSent_timestamp).Seconds())) // statistic may overflow (rollback to 0)
			}
			t.lastseen_PacketsSent = newPacketsSent // statistic may overflow (rollback to 0)
			t.lastseen_PacketsSent_timestamp = newPacketsSent_timestamp
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get TotalPacketsSent, %s", err.Error())
		}

		newUptime, err := GetUptime()
		if err == nil {
			t.Uptime = newUptime
		} else {
			util.MaoLogM(util.WARN, p_TPLINK_MODULE_NAME, "Fail to get Uptime, %s", err.Error())
		}

		util.MaoLogM(util.HOT_DEBUG, p_TPLINK_MODULE_NAME, "BytesSentSpeed: %d, BytesReceivedSpeed: %d, PacketsSentSpeed: %d, PacketsReceivedSpeed: %d, Uptime: %d",
			t.BytesSentSpeed, t.BytesReceivedSpeed, t.PacketsSentSpeed, t.PacketsReceivedSpeed, t.Uptime)
	}
}

func (t *TplinkGatewayModule) InitTplinkGatewayModule() bool {
	go t.controlLoop()
	return true
}