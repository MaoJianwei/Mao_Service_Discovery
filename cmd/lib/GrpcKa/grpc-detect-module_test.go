package GrpcKa

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestGrpcDetectModule_GetServiceInfo(t *testing.T) {
	serverNode := &MaoApi.GrpcServiceNode{
		ReportTimes:    666,
		Hostname:       "qingdao-hostname",
		Ips:            nil,
		ServerDateTime: time.Now().String(),
		OtherData:		"OtherDataOtherData",
		RealClientAddr: "1.1.1.1",
		LocalLastSeen:  time.Now(),
		Alive:          true,
	}
	s := fmt.Sprintf("Service: %s\r\nUp Time: %s\r\nDetail: %v\r\n",
		serverNode.Hostname, time.Now().String(), serverNode)

	log.Println(s)
}