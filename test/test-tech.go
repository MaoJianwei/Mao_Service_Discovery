
package main

import (
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"time"
)

func sendV4Ping(connV4 *icmp.PacketConn) {
	//conn packetConn
	//conn.SetFlagTTL()

	icmpPayloadData := append([]byte(time.Now().String()))
	echoMsg := icmp.Echo{
		ID:   0xabcd,
		Seq:  0x1994,
		Data: icmpPayloadData,
	}

	icmpMsg := icmp.Message{
		Type:     ipv4.ICMPTypeEcho,
		Code:     0,
		//Checksum: 0,
		Body:     &echoMsg,
	}

	// do le->be in the Marshal
	icmpMsgByte, err := icmpMsg.Marshal(nil)
	if err != nil {
		log.Printf("Fail to marshal icmpMsg: %s", err.Error())
	}

	v4Addr, err := net.ResolveIPAddr("ip", "127.0.0.1")
	if err != nil {
		log.Printf("Fail to ResolveIPAddr v4Addr: %s", err.Error())
	}

	count, err := connV4.WriteTo(icmpMsgByte, v4Addr)
	if err != nil {
		log.Printf("Fail to WriteTo connV4: %s", err.Error())
	}
	log.Printf("WriteTo connV4 %d: %s: %s --- %v \n--- %v", count, v4Addr.String(), v4Addr.Network(), icmpMsgByte, icmpMsg)
}

func main() {
	connV4, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Printf("Listen for v4 error: %s", err.Error())
	}
	log.Printf("v4 conn: %v", connV4)
	connV6, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		log.Printf("Listen for v6 error: %s", err.Error())
	}
	log.Printf("v6 conn: %v", connV6)

	for {
		sendV4Ping(connV4)
		time.Sleep(500 * time.Millisecond)
	}
}
