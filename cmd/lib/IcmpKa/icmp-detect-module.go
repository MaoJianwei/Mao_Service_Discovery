package IcmpKa

import (
	"MaoServerDiscovery/cmd/lib/Config"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	//addServiceChan *chan string
	//delServiceChan *chan string
	//configService  []*MaoIcmpService
)

const (
	URL_CONFIG_HOMEPAGE        = "/configIcmp"
	URL_CONFIG_ADD_SERVICE_IP  = "/addServiceIp"
	URL_CONFIG_DEL_SERVICE_IP  = "/delServiceIp"
	URL_CONFIG_SHOW_SERVICE_IP = "/showServiceIP"

	PROTO_ICMP    = 1
	PROTO_ICMP_V6 = 58

	ICMP_DETECT_ID    = 0x1994
	ICMP_V6_DETECT_ID = 0x1996

	SERVICE_LIST_CONFIG_PATH = "/icmp-ka/services"
)

type MaoIcmpService struct {
	Address string

	Alive    bool
	LastSeen time.Time

	DetectCount uint64
	ReportCount uint64

	RttDuration          int64
	RttOutboundTimestamp time.Time
}

type IcmpDetectModule struct {
	connV4       *icmp.PacketConn
	connV6       *icmp.PacketConn
	serviceStore sync.Map // address_string -> Service object

	AddChan chan string // need to be initiated when constructing
	DelChan chan string // need to be initiated when constructing

	// TODO - MAKE IT CONFIGURABLE
	// configurable parameter
	sendInterval uint32 // milliseconds
	checkInterval uint32 // milliseconds
	leaveTimeout uint32 // milliseconds
	refreshShowingInterval uint32 //

	// TODO - MAKE IT CONFIGURABLE
	// tunable configurable parameter
	receiveFreezePeriod uint32 // milliseconds - mitigate attack with malformed packets.

	// only for web showing
	configService  []*MaoIcmpService
}

func (m *IcmpDetectModule) sendIcmpLoop() {
	round := 1
	for {
		util.MaoLog(util.DEBUG, fmt.Sprintf("Detect Round %d", round))
		m.serviceStore.Range(func(_, value interface{}) bool {
			service := value.(*MaoIcmpService)

			addr, err := net.ResolveIPAddr("ip", service.Address)
			if err != nil {
				util.MaoLog(util.WARN, fmt.Sprintf("Fail to ResolveIPAddr v4v6Addr: %s", err.Error()))
				return true // for continuous iteration
			}

			var msgType icmp.Type
			var echoId int
			var conn *icmp.PacketConn
			if util.JudgeIPv6Addr(addr) {
				msgType = ipv6.ICMPTypeEchoRequest
				echoId = ICMP_V6_DETECT_ID
				conn = m.connV6
			} else {
				msgType = ipv4.ICMPTypeEcho
				echoId = ICMP_DETECT_ID
				conn = m.connV4
			}


			// To build and send ICMP Request.

			service.DetectCount++
			icmpPayloadData := []byte(time.Now().String())
			echoMsg := icmp.Echo{
				ID:   echoId,
				Seq:  int(service.DetectCount),
				Data: icmpPayloadData,
			}

			icmpMsg := icmp.Message{
				Type: msgType,
				Code: 0,
				//Checksum: 0,
				Body: &echoMsg,
			}

			// do le->be in the Marshal
			icmpMsgByte, err := icmpMsg.Marshal(nil)
			if err != nil {
				util.MaoLog(util.WARN, fmt.Sprintf("Fail to marshal icmpMsg: %s", err.Error()))
				return true
			}

			service.RttOutboundTimestamp = time.Now()
			_, err = conn.WriteTo(icmpMsgByte, addr)
			if err != nil {
				util.MaoLog(util.WARN, fmt.Sprintf("Fail to WriteTo connV6: %s", err.Error()))
				return true
			}

			return true
		})
		time.Sleep(time.Duration(m.sendInterval) * time.Millisecond)
		round++
	}
}

/**
 * For IPv6: PROTO_ICMP, m.connV4
 * For IPv4: PROTO_ICMP_V6, m.connV6
 */
func (m *IcmpDetectModule) receiveProcessIcmpLoop(protoNum int, conn *icmp.PacketConn) {
	recvBuf := make([]byte, 2000)
	for {
		count, addr, err := conn.ReadFrom(recvBuf)
		lastseen := time.Now()
		if err != nil {
			util.MaoLog(util.WARN, fmt.Sprintf("Fail to recv ICMP, freeze %d ms, %s", m.receiveFreezePeriod, err.Error()))
			time.Sleep(time.Duration(m.receiveFreezePeriod) * time.Millisecond)
			continue
		}

		msg, err := icmp.ParseMessage(protoNum, recvBuf)
		if err != nil {
			util.MaoLog(util.WARN, fmt.Sprintf("Fail to parse ICMP, freeze %d ms, %s", m.receiveFreezePeriod, err.Error()))
			time.Sleep(time.Duration(m.receiveFreezePeriod) * time.Millisecond)
			continue
		}

		icmpEcho, ok := msg.Body.(*icmp.Echo)
		if !ok {
			util.MaoLog(util.WARN, fmt.Sprintf("Fail to convert *icmp.Echo, freeze %d ms, %s", m.receiveFreezePeriod, err.Error()))
			time.Sleep(time.Duration(m.receiveFreezePeriod) * time.Millisecond)
			continue
		}
		util.MaoLog(util.DEBUG, fmt.Sprintf("%v, %v = %v, %v, %v, %v, %v, %v", count, addr, msg.Type, msg.Code, msg.Checksum, icmpEcho.ID, icmpEcho.Seq, icmpEcho.Data))

		var addrStr string
		if protoNum == PROTO_ICMP_V6 {
			addrStr = strings.Split(addr.String(), "%")[0] // for ipv6 link-local address, it is suffixed by % and interface name.
		} else {
			addrStr = addr.String()
		}
		value, ok := m.serviceStore.Load(addrStr)
		if ok && value != nil {
			service := value.(*MaoIcmpService)
			service.Alive = true
			service.LastSeen = lastseen
			service.RttDuration = service.LastSeen.Sub(service.RttOutboundTimestamp).Nanoseconds()
			service.ReportCount++
		}
	}
}

func (m *IcmpDetectModule) controlLoop() {
	checkTimer := time.NewTimer(time.Duration(m.checkInterval) * time.Millisecond)
	for {
		select {
		case addService := <-m.AddChan:
			if _, ok := m.serviceStore.Load(addService); !ok {
				m.serviceStore.Store(addService, &MaoIcmpService{
					Address:              addService,
					Alive:                false,
					LastSeen:             time.Unix(0, 0),
					DetectCount:          0,
					ReportCount:          0,
					RttDuration:          0,
					RttOutboundTimestamp: time.Time{},
				})
				util.MaoLog(util.DEBUG, fmt.Sprintf("Get new service %s", addService))
				m.addNewServiceToConfig(addService)
			}
		case delService := <-m.DelChan:
			m.serviceStore.Delete(delService)
			util.MaoLog(util.DEBUG, fmt.Sprintf("Del service %s", delService))
			m.removeOldServiceFromConfig(delService)
		case <-checkTimer.C:
			// aliveness checking
			m.serviceStore.Range(func(key, value interface{}) bool {
				service := value.(*MaoIcmpService)
				if service.Alive && time.Since(service.LastSeen) > time.Duration(m.leaveTimeout) * time.Second {
					service.Alive = false
				}
				return true
			})
			checkTimer.Reset(time.Duration(m.checkInterval) * time.Millisecond)
		}
	}
}

func (m *IcmpDetectModule) refreshShowingService() {
	for {
		time.Sleep(time.Duration(m.refreshShowingInterval) * time.Millisecond)
		newConfigService := []*MaoIcmpService{}
		m.serviceStore.Range(func(_, value interface{}) bool {
			newConfigService = append(newConfigService, value.(*MaoIcmpService))
			return true
		})
		m.configService = newConfigService
	}
}

func (m *IcmpDetectModule) getServiceConfig() (serviceList []string){
	configModule := MaoCommon.ServiceRegistryGetConfigModule()
	if configModule == nil {
		util.MaoLog(util.WARN, "Fail to get config module instance")
		return nil
	}

	serviceObj, errCode := configModule.GetConfig(SERVICE_LIST_CONFIG_PATH)
	if errCode != Config.ERR_CODE_SUCCESS {
		util.MaoLog(util.WARN, "Fail to get current services from config, errCode: %d", errCode)
		return nil
	}

	services := serviceObj.([]string)
	return services
}

func (m *IcmpDetectModule) saveServiceConfig(serviceList []string) (success bool){
	configModule := MaoCommon.ServiceRegistryGetConfigModule()
	if configModule == nil {
		util.MaoLog(util.WARN, "Fail to get config module instance")
		return false
	}

	_, errCode := configModule.PutConfig(SERVICE_LIST_CONFIG_PATH, serviceList)
	if errCode != Config.ERR_CODE_SUCCESS {
		util.MaoLog(util.WARN, "Fail to put current services to config, errCode: %d", errCode)
		return false
	}

	return true
}

func (m *IcmpDetectModule) addNewServiceToConfig(serviceAddr string) (success bool) {
	currentServices := m.getServiceConfig()
	if currentServices == nil {
		util.MaoLog(util.WARN, "Fail to get current services from config")
		return false
	}

	currentServices = append(currentServices, serviceAddr)

	return m.saveServiceConfig(currentServices)
}

func (m *IcmpDetectModule) removeOldServiceFromConfig(serviceAddr string) (success bool) {
	currentServices := m.getServiceConfig()
	if currentServices == nil {
		util.MaoLog(util.WARN, "Fail to get current services from config")
		return false
	}

	// assume that: the service address appears only once in the config
	for index, s := range currentServices {
		if s == serviceAddr {
			currentServices = append(currentServices[:index], currentServices[index+1:]...)
			m.saveServiceConfig(currentServices)
			return true
		}
	}

	util.MaoLog(util.WARN, "Can't find the service in the config, can't remove it, service: %s", serviceAddr)
	return false
}






func (m *IcmpDetectModule) AddService(serviceIPv4v6 string) {
	if net.ParseIP(serviceIPv4v6) != nil {
		m.AddChan <- serviceIPv4v6
	}
}

func (m *IcmpDetectModule) DelService(serviceIPv4v6 string) {
	if net.ParseIP(serviceIPv4v6) != nil {
		m.DelChan <- serviceIPv4v6
	}
}


func (m *IcmpDetectModule) InitIcmpModule() bool {
	var err error
	m.connV4, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to listen ICMP, %s", err.Error()))
		return false
	}
	util.MaoLog(util.INFO, "Listen ICMP ok")

	m.connV6, err = icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to listen ICMPv6, %s", err.Error()))
		return false
	}
	util.MaoLog(util.INFO, "Listen ICMPv6 ok")

	m.AddChan = make(chan string, 50)
	m.DelChan = make(chan string, 50)


	// configurable parameter
	m.sendInterval = 200
	m.checkInterval = 500
	m.leaveTimeout = 2000
	m.refreshShowingInterval = 1000

	// tunable configurable parameter
	m.receiveFreezePeriod = 10
	m.configService = []*MaoIcmpService{}


	go m.receiveProcessIcmpLoop(PROTO_ICMP, m.connV4)
	go m.receiveProcessIcmpLoop(PROTO_ICMP_V6, m.connV6)
	go m.sendIcmpLoop()
	go m.controlLoop()

	go m.refreshShowingService()

	m.configRestControlInterface()

	return true
}





func showConfigPage(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func (m *IcmpDetectModule) showServiceIps(c *gin.Context) {
	tmp := m.configService
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Address < tmp[j].Address
	})
	c.JSON(200, tmp)
}

func (m *IcmpDetectModule) processServiceIp(c *gin.Context) {
	v4Ip, ok := c.GetPostForm("ipv4v6")
	if ok {
		v4IpArr := strings.Fields(v4Ip)
		for _, s := range v4IpArr {
			if c.FullPath() == URL_CONFIG_ADD_SERVICE_IP {
				m.AddService(s)
			} else {
				m.DelService(s)
			}
		}
	}

	showConfigPage(c)
}

//func runRestControlInterface(ControlPort uint32) {
//	gin.SetMode(gin.ReleaseMode)
//	restful := gin.Default()
//	restful.LoadHTMLFiles("resource/index.html")
//	restful.Static("/static/", "resource")
//
//	restful.GET(URL_CONFIG_HOMEPAGE, showConfigPage)
//	restful.GET(URL_CONFIG_SHOW_SERVICE_IP, showServiceIp)
//	restful.POST(URL_CONFIG_ADD_SERVICE_IP, processServiceIp)
//	restful.POST(URL_CONFIG_DEL_SERVICE_IP, processServiceIp)
//
//	err := restful.Run(fmt.Sprintf("[::]:%d", ControlPort))
//	if err != nil {
//		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to run config server, %s", err.Error()))
//	}
//}

func (m *IcmpDetectModule) configRestControlInterface() {
	restfulServer := MaoCommon.ServiceRegistryGetRestfulServerModule()
	if restfulServer == nil {
		util.MaoLog(util.WARN, "Fail to get RestfulServerModule, unable to register restful apis.")
		return
	}

	restfulServer.RegisterGetApi(URL_CONFIG_HOMEPAGE, showConfigPage)
	restfulServer.RegisterGetApi(URL_CONFIG_SHOW_SERVICE_IP, m.showServiceIps)

	restfulServer.RegisterPostApi(URL_CONFIG_ADD_SERVICE_IP, m.processServiceIp)
	restfulServer.RegisterPostApi(URL_CONFIG_DEL_SERVICE_IP, m.processServiceIp)
}

//func main() {
//	addServiceChan = make(chan string, 50)
//	delServiceChan = make(chan string, 50)
//
//	icmpDetectModule := &IcmpDetectModule{
//		AddChan:     &addServiceChan,
//		DelChan:     &delServiceChan,
//		ControlPort: 2468,
//	}
//
//	icmpDetectModule.InitIcmpModule()
//
//	go runRestControlInterface(icmpDetectModule.ControlPort)
//
//	for {
//		time.Sleep(1 * time.Second)
//		newConfigService := []*MaoIcmpService{}
//		icmpDetectModule.serviceStore.Range(func(_, value interface{}) bool {
//			newConfigService = append(newConfigService, value.(*MaoIcmpService))
//			return true
//		})
//		configService = newConfigService
//	}
//}