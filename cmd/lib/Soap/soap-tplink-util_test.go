package Soap

import (
	"log"
	"testing"
	"time"
)

func TestGetExternalIPAddress(t *testing.T) {
	for {
		totalBytesSent, err := GetTotalBytesSent()
		log.Printf("TotalBytesSent: %d, err: %d", totalBytesSent, err)

		totalBytesReceived, err := GetTotalBytesReceived()
		log.Printf("TotalBytesReceived: %d, err: %d", totalBytesReceived, err)

		totalPacketsSent, err := GetTotalPacketsSent()
		log.Printf("TotalPacketsSent: %d, err: %d", totalPacketsSent, err)

		totalPacketsReceived, err := GetTotalPacketsReceived()
		log.Printf("TotalPacketsReceived: %d, err: %d", totalPacketsReceived, err)

		uptime, err := GetUptime()
		log.Printf("Uptime: %d, err: %d", uptime, err)

		externalIPAddress, err := GetExternalIPAddress()
		log.Printf("ExternalIPAddress: %s, err: %d", externalIPAddress, err)

		time.Sleep(500 * time.Millisecond)
		log.Printf("\n\n\n\n\n\n==============================")
	}
}