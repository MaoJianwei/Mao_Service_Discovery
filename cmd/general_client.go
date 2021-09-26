package branch

import (
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	parent "MaoServerDiscovery/util"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os/exec"
	"strconv"
	"time"
)


func prepareNat66GatewayData(report *pb.ServerReport) {
	ccc := exec.Command("/bin/bash", "-c", "ip6tables -nvxL FORWARD | grep MaoIPv6In | awk '{printf $2}'")
	ininin, err := ccc.CombinedOutput()
	if err != nil {
		parent.MaoLog(parent.ERROR, "Fail to get MaoIPv6In, " + err.Error())
		return
	}
	ccc = exec.Command("/bin/bash", "-c", "ip6tables -nvxL FORWARD | grep MaoIPv6Out | awk '{printf $2}'")
	outoutout, err := ccc.CombinedOutput()
	if err != nil {
		parent.MaoLog(parent.ERROR, "Fail to get MaoIPv6Out, " + err.Error())
		return
	}
	v6In, err := strconv.ParseUint(string(ininin), 10, 64)
	if err != nil {
		parent.MaoLog(parent.ERROR, "Fail to parse MaoIPv6In, " + err.Error())
		return
	}
	v6Out, err := strconv.ParseUint(string(outoutout), 10, 64)
	if err != nil {
		parent.MaoLog(parent.ERROR, "Fail to parse MaoIPv6Out, " + err.Error())
		return
	}
	parent.MaoLog(parent.DEBUG, fmt.Sprintf("v6In: %d , v6Out: %d", v6In, v6Out))
	report.AuxData = fmt.Sprintf("{ \"v6In\":%d, \"v6Out\":%d}", v6In, v6Out)
}

func RunGeneralClient(report_server_addr *net.IP, report_server_port uint32, report_interval uint32, silent bool,
	nat66Gateway bool) {
	parent.MaoLog(parent.INFO, "Connect to center ...")
	for {
		serverAddr := parent.GetAddrPort(report_server_addr, report_server_port)
		parent.MaoLog(parent.INFO, fmt.Sprintf("Connect to %s ...", serverAddr))

		ctx, cancelCtx := context.WithTimeout(context.Background(), 3 * time.Second)
		connect, err := grpc.DialContext(ctx, serverAddr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			parent.MaoLog(parent.WARN, fmt.Sprintf("Retry, %s ...", err))
			continue
		}
		cancelCtx()
		parent.MaoLog(parent.INFO, "Connected.")

		client := pb.NewMaoServerDiscoveryClient(connect)
		streamClient, err := client.Report(context.Background())
		if err != nil {
			parent.MaoLog(parent.ERROR, fmt.Sprintf("Fail to get streamClient, %s", err))
			continue
		}
		parent.MaoLog(parent.INFO, "Got StreamClient.")

		count := 1
		for {
			dataOk := true
			hostname, _ := parent.GetHostname()
			if err != nil {
				hostname = "Mao-Unknown"
				dataOk = false
			}

			ips, _ := parent.GetUnicastIp()
			if err != nil {
				ips = []string{"Mao-Fail", err.Error()}
				dataOk = false
			}

			parent.MaoLog(parent.DEBUG, fmt.Sprintf("%d: To send", count))
			report := &pb.ServerReport{
				Ok:          dataOk,
				Hostname:    hostname,
				Ips:         ips,
				NowDatetime: time.Now().String(),
			}
			if nat66Gateway == true {
				prepareNat66GatewayData(report)
			}

			err := streamClient.Send(report)
			if err != nil {
				parent.MaoLog(parent.ERROR, fmt.Sprintf("Fail to report, %s", err))
				break
			}
			if silent == false {
				parent.MaoLog(parent.INFO, fmt.Sprintf("ServerReport - %v", report))
			}
			parent.MaoLog(parent.DEBUG, fmt.Sprintf("%d: Sent", count))

			count++
			time.Sleep(time.Duration(report_interval) * time.Millisecond)
		}
		time.Sleep(1 * time.Second)
	}
}
