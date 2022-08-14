package branch

import (
	"MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/AuxDataProcessor"
	config "MaoServerDiscovery/cmd/lib/Config"
	"MaoServerDiscovery/cmd/lib/Email"
	"MaoServerDiscovery/cmd/lib/GrpcKa"
	icmpKa "MaoServerDiscovery/cmd/lib/IcmpKa"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/cmd/lib/Restful"
	"MaoServerDiscovery/util"
	parent "MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"time"
)

const (
	s_MODULE_NAME = "General-Server"
)

//var (
	// serviceAlive []*MaoApi.GrpcServiceNode // Mao: Deprecated, 2022.07.08.
//)

/**
Mao: Deprecated, 2022.07.08.
Using grpcModule.GetServiceInfo() and check service.Alive instead.

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
*/


func getGrpcAliveService() []*MaoApi.GrpcServiceNode {
	serviceAliveTmp := make([]*MaoApi.GrpcServiceNode, 0)

	grpcModule := MaoCommon.ServiceRegistryGetGrpcKaModule()
	if grpcModule == nil {
		util.MaoLogM(util.WARN, s_MODULE_NAME, "Fail to get GrpcKaModule")
		return serviceAliveTmp
	}

	serviceInfo := grpcModule.GetServiceInfo()
	for _, s := range serviceInfo {
		if s.Alive {
			serviceAliveTmp = append(serviceAliveTmp, s)
		}
	}

	return serviceAliveTmp
}

func startCliOutput(dump_interval uint32) {
	count := 1
	for {
		services := getGrpcAliveService()

		dump := ""
		for _, s := range services {
			dump = fmt.Sprintf("%s%s => %s - %s\n", dump, s.Hostname, s.LocalLastSeen, s.Ips)
		}
		util.MaoLogM(util.INFO, s_MODULE_NAME, "========== %d ==========\n%s", count, dump)

		count++
		time.Sleep(time.Duration(dump_interval) * time.Millisecond)
	}
}

/**
Mao: Deprecated, 2022.07.08.
Deprecated "/json" and "/plain", use "/" and "/showMergeServiceIP" instead.

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
 */

func showMergeServer(c *gin.Context) {
	c.HTML(200, "index-server.html", nil)
}
func showMergeServiceIP(c *gin.Context) {
	ret := make([]interface{}, 0)

	icmpModule := MaoCommon.ServiceRegistryGetIcmpKaModule()
	if icmpModule == nil {
		util.MaoLogM(util.WARN, s_MODULE_NAME, "Fail to get IcmpKaModule")
		c.JSON(202, ret)
	}
	services := icmpModule.GetServices()
	for _, s := range services {
		ret = append(ret, s)
	}

	// TODO: because we haven't provided a method to remove dead/alive Grpc services yet,
	// so let us just show alive services to WebUI now.
	serviceAliveTmp := getGrpcAliveService()
	for _, s := range serviceAliveTmp {
		ret = append(ret, s)
	}

	c.JSON(200, ret)
}


func RunServer(
	report_server_addr *net.IP, report_server_port uint32, web_server_addr *net.IP, web_server_port uint32,
	influxdbUrl string, influxdbToken string, influxdbOrgBucket string,
	dump_interval uint32, refresh_interval uint32, minLogLevel util.MaoLogLevel, silent bool) {

	util.InitMaoLog(minLogLevel)


	// ====== Restful Server module - part 1/2 ======
	restfulServer := &Restful.RestfulServerImpl{}
	restfulServer.InitRestfulServer()

	MaoCommon.RegisterService(MaoApi.RestfulServerRegisterName, restfulServer)

	// register server.go's api
	//restfulServer.RegisterGetApi("/json", showServers) // Mao: Deprecated, 2022.07.08.
	//restfulServer.RegisterGetApi("/plain", showServerPlain) // Mao: Deprecated, 2022.07.08.
	restfulServer.RegisterGetApi("/showMergeServiceIP", showMergeServiceIP)
	restfulServer.RegisterGetApi("/", showMergeServer)
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


	// ====== SMTP Email module ======
	smtpEmailModule := &Email.SmtpEmailModule{}
	if !smtpEmailModule.InitSmtpEmailModule() {
		return
	}

	MaoCommon.RegisterService(MaoApi.EmailModuleRegisterName, smtpEmailModule)
	// ============================


	// ====== Wechat Message module ======
	//wechatMessageModule := &Wechat.WechatMessageModule{}
	//if !wechatMessageModule.InitWechatMessageModule() {
	//	return
	//}
	//
	//MaoCommon.RegisterService(MaoApi.WechatModuleRegisterName, wechatMessageModule)
	// ============================


	// ====== Restful Server module - part 2/2 ======
	restfulServer.StartRestfulServerDaemon(parent.GetAddrPort(web_server_addr, web_server_port))
	// ==============================================



	// ====== Aux Data Processor module ======
	AuxDataProcessor.ConfigInfluxdbUtils(influxdbUrl, influxdbToken, influxdbOrgBucket)

	auxDataModule := &AuxDataProcessor.AuxDataProcessorModule{}
	auxDataModule.InitAuxDataProcessor()

	MaoCommon.RegisterService(MaoApi.AuxDataModuleRegisterName, auxDataModule)

	envTempProcessor := AuxDataProcessor.EnvTempProcessor{}
	var envTempProcessorAux MaoApi.AuxDataProcessor = envTempProcessor
	auxDataModule.AddProcessor(&envTempProcessorAux)
	// =======================================


	if !silent {
		go startCliOutput(dump_interval)
	}

	// updateServerAlive(refresh_interval) // Mao: Deprecated, 2022.07.08.
	for {
		time.Sleep(1 * time.Minute)
	}
}
