package GrpcKa

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	"MaoServerDiscovery/util"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	MODULE_NAME = "GRPC-Detect-module"
)

type GrpcDetectModule struct {
	serverInfo sync.Map
	mergeChannel chan *MaoApi.GrpcServiceNode
	rttMergeChannel chan *MaoApi.GrpcServiceNode

	server *grpc.Server
	pb.UnimplementedMaoServerDiscoveryServer

	checkInterval uint32 // milliseconds
	leaveTimeout uint32 // milliseconds
	refreshShowingInterval uint32 // milliseconds

	// used for web showing, i.e. external get operation
	// used for processing aux data
	serverInfoMirror  []*MaoApi.GrpcServiceNode
}

// implement pb.UnimplementedMaoServerDiscoveryServer
func (g *GrpcDetectModule) Report(reportStream pb.MaoServerDiscovery_ReportServer) error {
	util.MaoLogM(util.DEBUG, MODULE_NAME, "Triggered new report session")
	ctx := reportStream.Context()
	peerCtx, okbool := peer.FromContext(ctx)
	if !okbool {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get peerCtx, @Report")
		return errors.New("Fail to get peerCtx, @Report")
	}
	//peerMetadata, okbool := metadata.FromIncomingContext(ctx)
	//if !okbool {
	//	util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get peerMetadata")
	//	return errors.New("Fail to get peerMetadata")
	//}
	//transportstream := grpc.ServerTransportStreamFromContext(ctx)

	util.MaoLogM(util.INFO, MODULE_NAME, "New server comming: %s", peerCtx.Addr.String())
	_ = g.dealRecv(reportStream)
	return nil
}

func (g *GrpcDetectModule) dealRecv(reportStream pb.MaoServerDiscovery_ReportServer) error {
	ctx := reportStream.Context()
	peerCtx, okbool := peer.FromContext(ctx)
	var clientAddr string
	if !okbool {
		clientAddr = "Client-<unknown>"
	}
	clientAddr = peerCtx.Addr.String()

	var count uint64 = 1
	for {
		report, err := reportStream.Recv()
		if err != nil {
			util.MaoLogM(util.ERROR, MODULE_NAME, "Report err: <%s> %s", clientAddr, err)
			return err
		}
		util.MaoLogM(util.DEBUG, MODULE_NAME, "Report get: <%s> %s, %v", clientAddr, report.GetHostname(), report.GetIps())
		if report.GetOk() {
			g.mergeChannel <- &MaoApi.GrpcServiceNode{
				ReportTimes:    count,
				Hostname:       report.GetHostname(),
				Ips:            report.GetIps(),
				ServerDateTime: report.GetNowDatetime(),
				OtherData:		report.GetAuxData(),
				RealClientAddr: clientAddr,
				LocalLastSeen:  time.Now(),
				Alive:          true,
			}
		}
		count++
	}
}


func (g *GrpcDetectModule) RttMeasure(rttMeasureStream pb.MaoServerDiscovery_RttMeasureServer) error {
	util.MaoLogM(util.DEBUG, MODULE_NAME, "Triggered new RTT measure session")
	ctx := rttMeasureStream.Context()
	peerCtx, okbool := peer.FromContext(ctx)
	if !okbool {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get peerCtx, @RttMeasure")
		return errors.New("Fail to get peerCtx, @RttMeasure")
	}
	//peerMetadata, okbool := metadata.FromIncomingContext(ctx)
	//transportstream := grpc.ServerTransportStreamFromContext(ctx)

	util.MaoLogM(util.INFO, MODULE_NAME, "New RTT measure session for %s", peerCtx.Addr.String())
	_ = g.doRttMeasure(rttMeasureStream, peerCtx.Addr.String())
	return nil
}

func (g *GrpcDetectModule) doRttMeasure(rttMeasureStream pb.MaoServerDiscovery_RttMeasureServer, clientAddr string) error {
	var count uint64 = 0
	for {
		count++

		echoRequest := pb.RttEchoRequest{Seq: count}
		t1 := time.Now()
		if err := rttMeasureStream.Send(&echoRequest); err != nil {
			util.MaoLogM(util.ERROR, MODULE_NAME, "Fail to send Rtt echo request to %s, %s", clientAddr, err)
			return err
		}

		// gRPC is based on TCP now, so we don't need to do timeout setting.
		// And, if the TCP connection broke, Recv() will return with error.
		echoResponse, err := rttMeasureStream.Recv()
		if err != nil {
			util.MaoLogM(util.ERROR, MODULE_NAME, "Fail to recv Rtt echo response from %s, %s", clientAddr, err)
			return err
		}
		t2 := time.Now()
		if echoResponse.Ack == echoRequest.Seq {
			duration := t2.Sub(t1) // nanosecond
			g.rttMergeChannel <- &MaoApi.GrpcServiceNode{
				Hostname:    echoResponse.GetHostname(),
				RttDuration: duration,
			}
			util.MaoLogM(util.DEBUG, MODULE_NAME, "Calculated RTT delay %s for %s", duration.String(), echoResponse.GetHostname())
		}

		time.Sleep(1 * time.Second)
	}
}

func (g *GrpcDetectModule) runGrpcServer(listener net.Listener) {
	util.MaoLogM(util.INFO, MODULE_NAME, "Server running %s ...", listener.Addr().String())
	if err := g.server.Serve(listener); err != nil {
		util.MaoLogM(util.ERROR, MODULE_NAME, "%s", err)
	}
	util.MaoLogM(util.INFO, MODULE_NAME, "Serve over")
}




func (g *GrpcDetectModule) controlLoop() {
	checkTimer := time.NewTimer(time.Duration(g.checkInterval) * time.Millisecond)
	for {
		select {
		case serverNode := <-g.rttMergeChannel:
			value, ok := g.serverInfo.Load(serverNode.Hostname)
			if ok && value != nil {
				server := value.(*MaoApi.GrpcServiceNode)
				server.RttDuration = serverNode.RttDuration
			}
		case serverNode := <-g.mergeChannel:
			value, ok := g.serverInfo.Load(serverNode.Hostname)
			if ok && value != nil {
				server := value.(*MaoApi.GrpcServiceNode)
				if !server.Alive && serverNode.Alive {
					emailModule := MaoCommon.ServiceRegistryGetEmailModule()
					if emailModule == nil {
						util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get EmailModule, can't send UP notification")
					} else {
						emailModule.SendEmail(&MaoApi.EmailMessage{
							Subject: "Grpc UP notification",
							Content: fmt.Sprintf("Service: %s\r\nUp Time: %s\r\nDetail: %v\r\n",
								serverNode.Hostname, time.Now().String(), serverNode),
						})
					}
				}
				server.ReportTimes = serverNode.ReportTimes
				server.Hostname = serverNode.Hostname
				server.Ips = serverNode.Ips
				server.ServerDateTime = serverNode.ServerDateTime
				server.OtherData = serverNode.OtherData
				server.RealClientAddr = serverNode.RealClientAddr
				server.LocalLastSeen = serverNode.LocalLastSeen
				server.Alive = serverNode.Alive
			} else {
				// Attention, serverNode instance is not created always. 2023.07.24
				// TODO: other place may need to be check.
				g.serverInfo.Store(serverNode.Hostname, serverNode)
			}
		case <-checkTimer.C:
			// aliveness checking
			g.serverInfo.Range(func(key, value interface{}) bool {
				service := value.(*MaoApi.GrpcServiceNode)
				if service.Alive && time.Since(service.LocalLastSeen) > time.Duration(g.leaveTimeout) * time.Millisecond {
					service.Alive = false
					g.mergeChannel <- service

					emailModule := MaoCommon.ServiceRegistryGetEmailModule()
					if emailModule == nil {
						util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get EmailModule, can't send DOWN notification")
					} else {
						emailModule.SendEmail(&MaoApi.EmailMessage{
							Subject: "Grpc DOWN notification",
							Content: fmt.Sprintf("Service: %s\r\nDOWN Time: %s\r\nDetail: %v\r\n",
								service.Hostname, time.Now().String(), service),
						})
					}
				}
				return true
			})
			checkTimer.Reset(time.Duration(g.checkInterval) * time.Millisecond)
		}
	}
}



func (g *GrpcDetectModule) refreshShowingService() {
	for {
		time.Sleep(time.Duration(g.refreshShowingInterval) * time.Millisecond)
		serversTmp := make([]*MaoApi.GrpcServiceNode, 0)
		g.serverInfo.Range(func(_, value interface{}) bool {
			serversTmp = append(serversTmp, value.(*MaoApi.GrpcServiceNode))
			return true
		})
		g.serverInfoMirror = serversTmp
	}
}

func (g *GrpcDetectModule) GetServiceInfo() []*MaoApi.GrpcServiceNode {
	servers := g.serverInfoMirror
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Hostname < servers[j].Hostname
	})
	return servers
}



func (g *GrpcDetectModule) InitGrpcModule(addrPort string) bool {
	g.mergeChannel = make(chan *MaoApi.GrpcServiceNode, 1024)
	g.rttMergeChannel = make(chan *MaoApi.GrpcServiceNode, 1024)

	g.checkInterval = 500
	g.leaveTimeout = 5000
	g.refreshShowingInterval = 1000
	g.serverInfoMirror = make([]*MaoApi.GrpcServiceNode, 0)


	listener, err := net.Listen("tcp", addrPort)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to create listener at %s, err: %s", addrPort, err.Error())
		return false
	}

	g.server = grpc.NewServer()
	pb.RegisterMaoServerDiscoveryServer(g.server, g)
	go g.runGrpcServer(listener)


	go g.controlLoop()
	go g.refreshShowingService()

	return true
}

