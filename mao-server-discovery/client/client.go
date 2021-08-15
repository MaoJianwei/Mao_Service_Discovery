package main

import (
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	parent "MaoServerDiscovery/util"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

const (
	serverAddr = "127.0.0.1:28888"
)

func main() {
	parent.MaoLog(parent.INFO, "Connect to center ...")
	for {
		ctx, cancelCtx := context.WithTimeout(context.Background(), 3*time.Second)
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
			err := streamClient.Send(report)
			if err != nil {
				parent.MaoLog(parent.ERROR, fmt.Sprintf("Fail to report, %s", err))
				break
			}
			parent.MaoLog(parent.INFO, fmt.Sprintf("ServerReport - %v", report))
			parent.MaoLog(parent.DEBUG, fmt.Sprintf("%d: Sent", count))
			count++
			time.Sleep(1 * time.Second)
		}
		time.Sleep(1 * time.Second)
	}
}
