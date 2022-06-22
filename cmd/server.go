package branch

import (
	"MaoServerDiscovery/cmd/api"
	config "MaoServerDiscovery/cmd/lib/Config"
	icmpKa "MaoServerDiscovery/cmd/lib/IcmpKa"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/cmd/lib/Restful"
	pb "MaoServerDiscovery/grpc.maojianwei.com/server/discovery/api"
	"MaoServerDiscovery/util"
	parent "MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

//const (
//	addr    = "[::]:28888"
//	addrWeb = "[::]:29999"
//)

var (
	serverAlive []*ServerNode
)

type ServerNode struct {
	Hostname    string
	ReportTimes uint64

	Ips            []string
	RealClientAddr string

	OtherData string

	ServerDateTime string
	LocalLastSeen  time.Time
}

type RealMaoServerDiscoveryHWServer struct {
	mergeChannel chan *ServerNode
	pb.UnimplementedMaoServerDiscoveryServer
}

func (s *RealMaoServerDiscoveryHWServer) Report(reportStream pb.MaoServerDiscovery_ReportServer) error {
	util.MaoLog(util.DEBUG, fmt.Sprintf("Triggered new report round"))
	ctx := reportStream.Context()
	peerCtx, okbool := peer.FromContext(ctx)
	peerMetadata, okbool := metadata.FromIncomingContext(ctx)
	transportstream := grpc.ServerTransportStreamFromContext(ctx)
	util.MaoLog(util.INFO, fmt.Sprintf("New server comming: %s", peerCtx.Addr.String()))
	util.MaoLog(util.DEBUG, fmt.Sprintf("\n%v\n%v\n%v", peerCtx, peerMetadata, transportstream))
	if !okbool {
	}
	_ = dealRecv(reportStream, s.mergeChannel)
	return nil
}

func dealRecv(reportStream pb.MaoServerDiscovery_ReportServer, mergeChannel chan *ServerNode) error {
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
			util.MaoLog(util.ERROR, fmt.Sprintf("Report err: <%s> %s", clientAddr, err))
			return err
		}
		util.MaoLog(util.DEBUG, fmt.Sprintf("Report get: <%s> %s, %v", clientAddr, report.GetHostname(), report.GetIps()))
		if report.GetOk() {
			mergeChannel <- &ServerNode{
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


func mergeAliveServer(mergeChannel chan *ServerNode, serverInfo *sync.Map) {
	for serverNode := range mergeChannel {
		serverInfo.Store(serverNode.Hostname, serverNode)
	}
}

func updateServerAlive(serverInfo *sync.Map, refresh_interval uint32) {
	//count := 1
	for {
		servers := make([]*ServerNode, 0)
		serverInfo.Range(func(key, value interface{}) bool {
			servers = append(servers, value.(*ServerNode))
			return true
		})

		serverAliveTmp := make([]*ServerNode, 0)
		for _, s := range servers {
			if time.Now().Sub(s.LocalLastSeen) < 5 * time.Second {
				serverAliveTmp = append(serverAliveTmp, s)
			}
		}

		sort.Slice(serverAliveTmp, func(i, j int) bool {
			return serverAliveTmp[i].Hostname < serverAliveTmp[j].Hostname
		})

		serverAlive = serverAliveTmp

		//count++
		time.Sleep(time.Duration(refresh_interval) * time.Millisecond)
	}
}


func startCliOutput(dump_interval uint32) {
	count := 1
	for {
		servers := serverAlive

		dump := ""
		for _, s := range servers {
			dump = fmt.Sprintf("%s%s => %s - %s\n", dump, s.Hostname, s.LocalLastSeen, s.Ips)
		}
		util.MaoLog(util.INFO, fmt.Sprintf("========== %d ==========\n%s", count, dump))

		count++
		time.Sleep(time.Duration(dump_interval) * time.Millisecond)
	}
}

func showServers(c *gin.Context) {
	serverTmp := serverAlive
	c.IndentedJSON(200, serverTmp)
}
func showServerPlain(c *gin.Context) {
	serverTmp := serverAlive

	dump := "<html><meta http-equiv=\"refresh\" content=\"1\"><title>Mao Service Discovery</title><body>"
	for _, s := range serverTmp {
		dump = fmt.Sprintf("%s%s => %s - %s %s<br/>", dump, s.Hostname, s.LocalLastSeen, s.Ips, s.OtherData)
	}
	dump += "</body></html>"
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, dump)
}

func runGrpcServer(server *grpc.Server, listener net.Listener) {
	util.MaoLog(util.INFO, fmt.Sprintf("Server running %s ...", listener.Addr().String()))
	if err := server.Serve(listener); err != nil {
		util.MaoLog(util.ERROR, fmt.Sprintf("%s", err))
	}
	util.MaoLog(util.INFO, "Serve over")
}

func RunServer(
	report_server_addr *net.IP, report_server_port uint32, web_server_addr *net.IP, web_server_port uint32,
	dump_interval uint32, refresh_interval uint32, silent bool) {

	util.InitMaoLog()

	//
	mergeChannel := make(chan *ServerNode, 1024)
	serverInfo := sync.Map{}

	listener, err := net.Listen("tcp", parent.GetAddrPort(report_server_addr, report_server_port))
	if err != nil {
		log.Printf("%s", err)
		return
	}

	server := grpc.NewServer()
	pb.RegisterMaoServerDiscoveryServer(server, &RealMaoServerDiscoveryHWServer{
		mergeChannel: mergeChannel,
	})
	go runGrpcServer(server, listener)

	go mergeAliveServer(mergeChannel, &serverInfo)



	// ====== Restful Server module - part 1/2 ======
	restfulServer := &Restful.RestfulServerImpl{}
	restfulServer.InitRestfulServer()

	// register serger.go's api
	restfulServer.RegisterGetApi("/json", showServers)
	restfulServer.RegisterGetApi("/", showServerPlain)
	// ==============================================


	// ====== Config(YAML) module ======
	configModule := &config.ConfigYamlModule{}
	if !configModule.InitConfigModule("beijing.yaml") {
		return
	}

	MaoCommon.RegisterService(MaoApi.ConfigModuleRegisterName, configModule)
	// =================================


	// ====== ICMP KA module ======
	icmpDetectModule := &icmpKa.IcmpDetectModule{}
	if !icmpDetectModule.InitIcmpModule() {
		return
	}

	MaoCommon.RegisterService(MaoApi.IcmpKaModuleRegisterName, icmpDetectModule)
	// ============================

	// ====== Restful Server module - part 2/2 ======
	restfulServer.StartRestfulServerDaemon(parent.GetAddrPort(web_server_addr, web_server_port))
	// ==============================================



	if !silent {
		go startCliOutput(dump_interval)
	}

	updateServerAlive(&serverInfo, refresh_interval)
}
