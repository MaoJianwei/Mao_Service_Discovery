//
package main
//
//import (
//	"github.com/gin-gonic/gin"
//	"golang.org/x/net/icmp"
//	"golang.org/x/net/ipv4"
//	"log"
//	"net"
//	"strings"
//	"time"
//)
//
//var (
//	v4Services = map[string]bool{}
//	v6Services = map[string]bool{}
//)
//
//func receiveV4Ping() {
//
//}
//
//func sendV4Ping(connV4 *icmp.PacketConn) {
//	//conn packetConn
//	//conn.SetFlagTTL()
//
//	icmpPayloadData := append([]byte(time.Now().String()))
//	echoMsg := icmp.Echo{
//		ID:   0xabcd,
//		Seq:  0x1994,
//		Data: icmpPayloadData,
//	}
//
//	icmpMsg := icmp.Message{
//		Type:     ipv4.ICMPTypeEcho,
//		Code:     0,
//		//Checksum: 0,
//		Body:     &echoMsg,
//	}
//
//	// do le->be in the Marshal
//	icmpMsgByte, err := icmpMsg.Marshal(nil)
//	if err != nil {
//		log.Printf("Fail to marshal icmpMsg: %s", err.Error())
//	}
//
//	v4Addr, err := net.ResolveIPAddr("ip", "192.168.1.2")
//	if err != nil {
//		log.Printf("Fail to ResolveIPAddr v4Addr: %s", err.Error())
//	}
//
//	count, err := connV4.WriteTo(icmpMsgByte, v4Addr)
//	if err != nil {
//		log.Printf("Fail to WriteTo connV4: %s", err.Error())
//	}
//	log.Printf("WriteTo connV4 %d: %s: %s --- %v \n--- %v", count, v4Addr.String(), v4Addr.Network(), icmpMsgByte, icmpMsg)
//}
//
//func showConfigPage(c *gin.Context) {
//	c.HTML(200, "index.html", nil)
//}
//
//func addServiceIp(c *gin.Context) {
//	v4Ip, ok := c.GetPostForm("ipv4")
//	if ok {
//		v4IpArr := strings.Fields(v4Ip)
//		log.Printf("\n%v", v4IpArr)
//	}
//	v6Ip, ok := c.GetPostForm("ipv6")
//	if ok {
//		v6IpArr := strings.Fields(v6Ip)
//		log.Printf("\n%v", v6IpArr)
//	}
//}
//
//func startRestServer() {
//	gin.SetMode(gin.ReleaseMode)
//	restful := gin.Default()
//	restful.LoadHTMLFiles("index.html")
//	restful.GET("/", showConfigPage)
//	restful.POST("/addServiceIp", addServiceIp)
//	err := restful.Run("[::]:9876")
//	if err != nil {
//		log.Printf("qingdao %s", err.Error())
//	}
//}
//
//func main() {
//	go receiveV4Ping()
//	startRestServer()
//	return
//
//
//
//	connV4, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
//	if err != nil {
//		log.Printf("Listen for v4 error: %s", err.Error())
//	}
//	log.Printf("v4 conn: %v", connV4)
//	connV6, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
//	if err != nil {
//		log.Printf("Listen for v6 error: %s", err.Error())
//	}
//	log.Printf("v6 conn: %v", connV6)
//
//	for {
//		sendV4Ping(connV4)
//		time.Sleep(1000 * time.Millisecond)
//	}
//}
