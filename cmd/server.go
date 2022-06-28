package branch

import (
	"MaoServerDiscovery/cmd/api"
	config "MaoServerDiscovery/cmd/lib/Config"
	"MaoServerDiscovery/cmd/lib/GrpcKa"
	icmpKa "MaoServerDiscovery/cmd/lib/IcmpKa"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/cmd/lib/Restful"
	"MaoServerDiscovery/util"
	parent "MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"sort"
	"time"
)

const (
	s_MODULE_NAME = "General-Server"
)

var (
	serviceAlive []*MaoApi.GrpcServiceNode
)

func updateServerAlive(refresh_interval uint32) {
	for {
		time.Sleep(time.Duration(refresh_interval) * time.Millisecond)

		grpcModule := MaoCommon.ServiceRegistryGetGrpcKaModule()
		if grpcModule == nil {
			util.MaoLogM(util.WARN, s_MODULE_NAME, "Fail to get GrpcKaModule")
			continue
		}
		serviceInfo := grpcModule.GetServiceInfo()

		serviceAliveTmp := make([]*MaoApi.GrpcServiceNode, 0)
		for _, s := range serviceInfo {
			if time.Now().Sub(s.LocalLastSeen) < 5 * time.Second {
				serviceAliveTmp = append(serviceAliveTmp, s)
			}
		}

		sort.Slice(serviceAliveTmp, func(i, j int) bool {
			return serviceAliveTmp[i].Hostname < serviceAliveTmp[j].Hostname
		})

		serviceAlive = serviceAliveTmp
	}
}


func startCliOutput(dump_interval uint32) {
	count := 1
	for {
		services := serviceAlive

		dump := ""
		for _, s := range services {
			dump = fmt.Sprintf("%s%s => %s - %s\n", dump, s.Hostname, s.LocalLastSeen, s.Ips)
		}
		util.MaoLogM(util.INFO, s_MODULE_NAME, "========== %d ==========\n%s", count, dump)

		count++
		time.Sleep(time.Duration(dump_interval) * time.Millisecond)
	}
}


func showServers(c *gin.Context) {
	serverTmp := serviceAlive
	c.IndentedJSON(200, serverTmp)
}
func showServerPlain(c *gin.Context) {
	serverTmp := serviceAlive

	dump := "<html><meta http-equiv=\"refresh\" content=\"1\"><title>Mao Service Discovery</title><body>"
	for _, s := range serverTmp {
		dump = fmt.Sprintf("%s%s => %s - %s %s<br/>", dump, s.Hostname, s.LocalLastSeen, s.Ips, s.OtherData)
	}
	dump += "</body></html>"
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, dump)
}


func RunServer(
	report_server_addr *net.IP, report_server_port uint32, web_server_addr *net.IP, web_server_port uint32,
	dump_interval uint32, refresh_interval uint32, minLogLevel util.MaoLogLevel, silent bool) {

	util.InitMaoLog(minLogLevel)


	// ====== Restful Server module - part 1/2 ======
	restfulServer := &Restful.RestfulServerImpl{}
	restfulServer.InitRestfulServer()

	MaoCommon.RegisterService(MaoApi.RestfulServerRegisterName, restfulServer)

	// register server.go's api
	restfulServer.RegisterGetApi("/json", showServers)
	restfulServer.RegisterGetApi("/", showServerPlain)
	// ==============================================


	// ====== Config(YAML) module ======
	configModule := &config.ConfigYamlModule{}
	if !configModule.InitConfigModule(config.DEFAULT_CONFIG_FILE) {
		return
	}

	MaoCommon.RegisterService(MaoApi.ConfigModuleRegisterName, configModule)
	// =================================


	// ====== gRPC KA module ======
	grpcModule := &GrpcKa.GrpcDetectModule{}
	grpcModule.InitGrpcModule(parent.GetAddrPort(report_server_addr, report_server_port))

	MaoCommon.RegisterService(MaoApi.GrpcKaModuleRegisterName, grpcModule)
	// ============================


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

	updateServerAlive(refresh_interval)
}
