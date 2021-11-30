package IcmpKa

import (
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
	addServiceChan chan string
	delServiceChan chan string
	configService  []*MaoIcmpService
)

const (
	URL_CONFIG_HOMEPAGE        string = "/"
	URL_CONFIG_ADD_SERVICE_IP  string = "/addServiceIp"
	URL_CONFIG_DEL_SERVICE_IP  string = "/delServiceIp"
	URL_CONFIG_SHOW_SERVICE_IP string = "/showServiceIP"

	PROTO_ICMP    = 1
	PROTO_ICMP_V6 = 58

	ICMP_DETECT_ID    = 0x1994
	ICMP_V6_DETECT_ID = 0x1996
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

	//ControlPort uint32 // Todo: can be moved out
	AddChan *chan string
	DelChan *chan string
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
		time.Sleep(500 * time.Millisecond)
		round++
	}
}

/**
 * For IPv6: PROTO_ICMP, m.connV4
 * For IPv4: PROTO_ICMP_V6, m.connV6
 */
func (m *IcmpDetectModule) receiveProcessIcmpLoop(protoNum int, conn *icmp.PacketConn) {
	freeze_period := 500 // ms
	recvBuf := make([]byte, 2000)
	for {
		count, addr, err := conn.ReadFrom(recvBuf)
		lastseen := time.Now()
		if err != nil {
			util.MaoLog(util.WARN, fmt.Sprintf("Fail to recv ICMP, freeze %d ms, %s", freeze_period, err.Error()))
			time.Sleep(time.Duration(freeze_period) * time.Millisecond)
			continue
		}

		msg, err := icmp.ParseMessage(protoNum, recvBuf)
		if err != nil {
			util.MaoLog(util.WARN, fmt.Sprintf("Fail to parse ICMP, %s", err.Error()))
			continue
		}

		icmpEcho, ok := msg.Body.(*icmp.Echo)
		if !ok {
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
	checkPeriod := 1 * time.Second
	checkTimer := time.NewTimer(checkPeriod)
	for {
		select {
		case addService := <-*m.AddChan:
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
			}
		case delService := <-*m.DelChan:
			m.serviceStore.Delete(delService)
			util.MaoLog(util.DEBUG, fmt.Sprintf("Del service %s", delService))
		case <-checkTimer.C:
			m.serviceStore.Range(func(key, value interface{}) bool {
				service := value.(*MaoIcmpService)
				if service.Alive && time.Since(service.LastSeen) > 3*time.Second {
					service.Alive = false
				}
				return true
			})
			checkTimer.Reset(checkPeriod)
		}
	}
}

func (m *IcmpDetectModule) refreshShowingService() {
	for {
		time.Sleep(1 * time.Second)
		newConfigService := []*MaoIcmpService{}
		m.serviceStore.Range(func(_, value interface{}) bool {
			newConfigService = append(newConfigService, value.(*MaoIcmpService))
			return true
		})
		configService = newConfigService
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

	go m.receiveProcessIcmpLoop(PROTO_ICMP, m.connV4)
	go m.receiveProcessIcmpLoop(PROTO_ICMP_V6, m.connV6)
	go m.sendIcmpLoop()
	go m.controlLoop()

	go m.refreshShowingService()

	return true
}





func showConfigPage(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func showServiceIp(c *gin.Context) {
	tmp := configService
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Address < tmp[j].Address
	})
	c.JSON(200, tmp)
}

func processServiceIp(c *gin.Context) {
	v4Ip, ok := c.GetPostForm("ipv4v6")
	if ok {
		v4IpArr := strings.Fields(v4Ip)
		for _, s := range v4IpArr {
			ip := net.ParseIP(s)
			if ip != nil {
				if c.FullPath() == URL_CONFIG_ADD_SERVICE_IP {
					addServiceChan <- s
				} else {
					delServiceChan <- s
				}
			}
		}
	}
	c.HTML(200, "index.html", nil)
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

func ConfigRestControlInterface(restful *gin.Engine) {
	restful.LoadHTMLFiles("resource/index.html")
	restful.Static("/static/", "resource")

	restful.GET(URL_CONFIG_HOMEPAGE, showConfigPage)
	restful.GET(URL_CONFIG_SHOW_SERVICE_IP, showServiceIp)
	restful.POST(URL_CONFIG_ADD_SERVICE_IP, processServiceIp)
	restful.POST(URL_CONFIG_DEL_SERVICE_IP, processServiceIp)
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
