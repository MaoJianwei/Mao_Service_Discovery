package branch

import (
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
	"os"
	"sort"
	"sync"
	"time"
)

//const (
//	addr    = "[::]:28888"
//	addrWeb = "[::]:29999"
//)

var (
	serverWebShow []*ServerNode
)

type ServerNode struct {
	Hostname    string
	ReportTimes uint64

	Ips            []string
	RealClientAddr string

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

func dumpAliveServer(serverInfo *sync.Map, dump_interval uint32) {
	count := 1
	for {
		servers := make([]*ServerNode, 0)
		serverInfo.Range(func(key, value interface{}) bool {
			servers = append(servers, value.(*ServerNode))
			return true
		})
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].Hostname < servers[j].Hostname
		})

		serverWebShow = servers

		dump := ""
		for _, s := range servers {
			if time.Now().Sub(s.LocalLastSeen) > 5 * time.Second {
				dump = fmt.Sprintf("%s%s => %s - %s\n", dump, s.Hostname, s.LocalLastSeen, s.Ips)
			}
		}
		util.MaoLog(util.INFO, fmt.Sprintf("========== %d ==========\n%s", count, dump))

		count++
		time.Sleep(time.Duration(dump_interval) * time.Millisecond)
	}
}

func startRestful(webAddr string) {
	util.MaoLog(util.INFO, fmt.Sprintf("Starting web show %s ...", webAddr))
	gin.SetMode(gin.ReleaseMode)
	restful := gin.Default()
	restful.GET("/", showServers)
	err := restful.Run(webAddr)
	if err != nil {
		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to run rest server, %s", err))
		return
	}
}

func showServers(c *gin.Context) {
	serverTmp := serverWebShow
	c.IndentedJSON(200, serverTmp)
}


func runServer(server *grpc.Server, listener net.Listener) {
	util.MaoLog(util.INFO, fmt.Sprintf("Server running %s ...", listener.Addr().String()))
	if err := server.Serve(listener); err != nil {
		util.MaoLog(util.ERROR, fmt.Sprintf("%s", err))
	}
	util.MaoLog(util.INFO, "Serve over")
}

func RunServer(report_server_addr *net.IP, report_server_port uint32, web_server_addr *net.IP, web_server_port uint32, dump_interval uint32) {

	log.SetOutput(os.Stdout)

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
	go runServer(server, listener)

	go mergeAliveServer(mergeChannel, &serverInfo)

	go startRestful(parent.GetAddrPort(web_server_addr, web_server_port))

	dumpAliveServer(&serverInfo, dump_interval)
}
