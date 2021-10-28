package main

import (
	"MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	addServiceChan chan string
	delServiceChan chan string
	configService []*MaoIcmpService
)

const (
	URL_CONFIG_HOMEPAGE        string = "/"
	URL_CONFIG_ADD_SERVICE_IP  string = "/addServiceIp"
	URL_CONFIG_DEL_SERVICE_IP  string = "/delServiceIp"
	URL_CONFIG_SHOW_SERVICE_IP string = "/showServiceIP"
)

type MaoIcmpService struct {
	Address string

	alive    bool
	lastSeen string

	detectCount uint64
	reportCount uint64

	rttDuration          uint32
	rttOutboundTimestamp time.Time
}

type IcmpDetectModule struct {
	connV4       *icmp.PacketConn
	connV6       *icmp.PacketConn
	serviceStore sync.Map // address_string -> Service object

	controlPort uint32 // Todo: can be moved out
	addChan     *chan string
	delChan     *chan string
}

func (m *IcmpDetectModule) sendIcmpLoop() {
	round := 1
	for {
		util.MaoLog(util.DEBUG, fmt.Sprintf("Detect Round %d", round))
		m.serviceStore.Range(func(key, value interface{}) bool {
			address := key.(string)
			service := value.(*MaoIcmpService)

			icmpPayloadData := append([]byte(time.Now().String()))
			echoMsg := icmp.Echo{
				ID:   0x1994,                   // 0xabcd,
				Seq:  int(service.detectCount), // 0x1994,
				Data: icmpPayloadData,
			}

			icmpMsg := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				//Checksum: 0,
				Body: &echoMsg,
			}

			// do le->be in the Marshal
			icmpMsgByte, err := icmpMsg.Marshal(nil)
			if err != nil {
				log.Printf("Fail to marshal icmpMsg: %s", err.Error())
				return true
			}

			addr, err := net.ResolveIPAddr("ip", address)
			if err != nil {
				log.Printf("Fail to ResolveIPAddr v4Addr: %s", err.Error())
				return true
			}

			if util.JudgeIPv6Addr(addr) {

			} else {
				count, err := m.connV4.WriteTo(icmpMsgByte, addr)
				if err != nil {
					log.Printf("Fail to WriteTo connV4: %s", err.Error())
					return true
				}
				log.Printf("WriteTo connV4 %d: %s: %s --- %v \n--- %v", count, addr.String(), addr.Network(), icmpMsgByte, icmpMsg)
			}
			return true
		})
		time.Sleep(500 * time.Millisecond)
		round++
	}
}

func (m *IcmpDetectModule) receiveProcessIcmpLoop() {

}

func (m *IcmpDetectModule) controlLoop() {
	for {
		util.MaoLog(util.DEBUG, "Wait for control input ...")
		select {
		case addService := <- *m.addChan:
			if _, ok := m.serviceStore.Load(addService); !ok {
				m.serviceStore.Store(addService, &MaoIcmpService{
					Address:              addService,
					alive:                false,
					lastSeen:             "",
					detectCount:          0,
					reportCount:          0,
					rttDuration:          0,
					rttOutboundTimestamp: time.Time{},
				})
				util.MaoLog(util.DEBUG, fmt.Sprintf("Get new service %s", addService))
			}
		case delService := <- *m.delChan:
			m.serviceStore.Delete(delService)
			util.MaoLog(util.DEBUG, fmt.Sprintf("Del service %s", delService))
		}
	}
}

func (m *IcmpDetectModule) InitIcmpModule() bool {
	var err error
	m.connV4, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Printf("Listen for v4 error: %s", err.Error())
		return false
	}
	log.Printf("v4 conn: %v", m.connV4)

	m.connV6, err = icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		log.Printf("Listen for v6 error: %s", err.Error())
		return false
	}
	log.Printf("v6 conn: %v", m.connV6)

	//go m.sendIcmpLoop()
	//go m.receiveProcessIcmpLoop()
	go m.controlLoop()

	return true
}

func showServiceIp(c *gin.Context) {
	c.JSON(200, configService)
}

func showConfigPage(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func processServiceIp(c *gin.Context) {
	v4Ip, ok := c.GetPostForm("ipv4")
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
		log.Printf("\n%v", v4IpArr)
	}
	v6Ip, ok := c.GetPostForm("ipv6")
	if ok {
		v6IpArr := strings.Fields(v6Ip)
		for _, s := range v6IpArr {
			ip := net.ParseIP(s)
			if ip != nil {
				if c.FullPath() == URL_CONFIG_ADD_SERVICE_IP {
					addServiceChan <- s
				} else {
					delServiceChan <- s
				}
			}
		}
		log.Printf("\n%v", v6IpArr)
	}
}

func runRestControlInterface(controlPort uint32) {
	gin.SetMode(gin.ReleaseMode)
	restful := gin.Default()
	restful.LoadHTMLFiles("resource/index.html")
	restful.Static("/static/", "resource")

	restful.GET(URL_CONFIG_HOMEPAGE, showConfigPage)
	restful.GET(URL_CONFIG_SHOW_SERVICE_IP, showServiceIp)
	restful.POST(URL_CONFIG_ADD_SERVICE_IP, processServiceIp)
	restful.POST(URL_CONFIG_DEL_SERVICE_IP, processServiceIp)

	err := restful.Run(fmt.Sprintf("[::]:%d", controlPort))
	if err != nil {
		log.Printf("qingdao %s", err.Error())
	}
}

func main() {

	addServiceChan = make(chan string, 50)
	delServiceChan = make(chan string, 50)

	icmpDetectModule := &IcmpDetectModule{
		addChan:     &addServiceChan,
		delChan:     &delServiceChan,
		controlPort: 2468,
	}

	icmpDetectModule.InitIcmpModule()

	go runRestControlInterface(icmpDetectModule.controlPort)

	for {
		time.Sleep(1 * time.Second)
		newConfigService := []*MaoIcmpService{}
		icmpDetectModule.serviceStore.Range(func(_, value interface{}) bool {
			newConfigService = append(newConfigService, value.(*MaoIcmpService))
			return true
		})
		configService = newConfigService
	}
}
