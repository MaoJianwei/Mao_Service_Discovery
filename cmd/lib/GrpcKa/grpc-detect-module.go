package GrpcKa

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	"MaoServerDiscovery/util"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

	server *grpc.Server
	pb.UnimplementedMaoServerDiscoveryServer

	checkInterval uint32 // milliseconds
	leaveTimeout uint32 // milliseconds
	refreshShowingInterval uint32 // milliseconds

	// only for web showing, i.e. external get operation
	serverInfoMirror  []*MaoApi.GrpcServiceNode
}

// implement pb.UnimplementedMaoServerDiscoveryServer
func (g *GrpcDetectModule) Report(reportStream pb.MaoServerDiscovery_ReportServer) error {
	util.MaoLogM(util.DEBUG, MODULE_NAME, "Triggered new report round")
	ctx := reportStream.Context()
	peerCtx, okbool := peer.FromContext(ctx)
	peerMetadata, okbool := metadata.FromIncomingContext(ctx)
	transportstream := grpc.ServerTransportStreamFromContext(ctx)
	util.MaoLogM(util.INFO, MODULE_NAME, "New server comming: %s", peerCtx.Addr.String())
	util.MaoLogM(util.DEBUG, MODULE_NAME, "\n%v\n%v\n%v", peerCtx, peerMetadata, transportstream)
	if !okbool {
	}
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
			}
			g.serverInfo.Store(serverNode.Hostname, serverNode)
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

