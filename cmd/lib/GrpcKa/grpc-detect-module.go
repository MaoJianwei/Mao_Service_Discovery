package GrpcKa

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	"MaoServerDiscovery/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
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


func (g *GrpcDetectModule) mergeAliveServer() {
	for serverNode := range g.mergeChannel {
		g.serverInfo.Store(serverNode.Hostname, serverNode)
	}
}


func (g *GrpcDetectModule) GetServiceInfo() []*MaoApi.GrpcServiceNode {
	servers := make([]*MaoApi.GrpcServiceNode, 0)
	g.serverInfo.Range(func(key, value interface{}) bool {
		servers = append(servers, value.(*MaoApi.GrpcServiceNode))
		return true
	})
	return servers
}

func (g *GrpcDetectModule) InitGrpcModule(addrPort string) bool {
	g.mergeChannel = make(chan *MaoApi.GrpcServiceNode, 1024)

	listener, err := net.Listen("tcp", addrPort)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to create listener at %s, err: %s", addrPort, err.Error())
		return false
	}

	g.server = grpc.NewServer()
	pb.RegisterMaoServerDiscoveryServer(g.server, g)
	go g.runGrpcServer(listener)

	go g.mergeAliveServer()

	return true
}

