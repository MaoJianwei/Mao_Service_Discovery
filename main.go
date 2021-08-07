package main

import (
	"context"
	"fmt"
	etcd "go.etcd.io/etcd/client/v3"
	"log"
	"net"
	"os"
	"time"
)

const (
	KEEP_ALIVE_TTL = 10 // Second
)

func putKV(client *etcd.Client, lease etcd.LeaseID, key, value string) {
	log.Printf("To put { %s -> %s }\n", key, value)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	putResp, err := client.Put(ctx, key, value, etcd.WithLease(lease))
	cancel()
	if err != nil {
		log.Printf("Fail to put { %s -> %s }, Resp: %s, err: %s\n", key, value, putResp, err)
		return
	}

	log.Printf("Success to put { %s -> %s }, Resp: %s, err: %s\n", key, value, putResp, err)
}

func watchKeyPrefix(client *etcd.Client, key string) {

	// Mao: If using this, it will close channel automatically.
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	ctx := context.Background()

	log.Printf("To watch: %s\n", key)
	watchChannel := client.Watch(ctx, key, etcd.WithPrefix())
	for watchResp := range watchChannel {
		log.Printf("New WatchResponse, %s", watchResp)
		for _, ev := range watchResp.Events {
			log.Printf("New change { %s -> %s }, Old: { %s -> %s }, Type: %s, ev:%s\n",
				ev.Kv.Key, ev.Kv.Value, "ev.PrevKv.Key nil", "ev.PrevKv.Value nil", ev.Type, ev)
		}
	}

	log.Printf("Finish watch")
}

// useless
func showLease(client *etcd.Client) {
	leasesResp, err := client.Leases(context.Background())
	if err != nil {
		log.Printf("Fail to get leases, err: %s\n", err)
	}

	log.Printf("Found %d leases", len(leasesResp.Leases))
	for _, leaseStatus := range leasesResp.Leases {
		log.Printf("Lease id: %d", leaseStatus.ID)
	}
}

func keepAliveLease(client *etcd.Client, id etcd.LeaseID) {
	log.Printf("Start keep-alive for lease %d ...", id)
	kaRespChan, err := client.KeepAlive(context.Background(), id)
	if err != nil {
		log.Printf("Fail to start keep-alive for lease %d", id)
		return
	}

	for kaResp := range kaRespChan {
		log.Printf("KA - ID: %d, TTL: %d", kaResp.ID, kaResp.TTL)
	}
}

func grantAndKeepAliveLease(client *etcd.Client) (etcd.LeaseID, error) {

	lgResp, err := client.Grant(context.Background(), KEEP_ALIVE_TTL)
	if err != nil {
		log.Printf("Fail to grant a new lease, err: %s", err)
		return -1, err
	}

	log.Printf("Granted lease ID: %d, TTL: %d", lgResp.ID, lgResp.TTL)
	go keepAliveLease(client, lgResp.ID)

	return lgResp.ID, nil
}

func createClient() (*etcd.Client, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"pi-dpdk.maojianwei.com:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		log.Printf("Fail to generate client, %s.\n", err)
		return nil, err
	}
	return client, nil
	//defer client.Close()

}

func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Fail to get hostname")
		return "", err
	}
	log.Printf("%s", hostname)
	return hostname, nil
}

func getUnicastIp() ([]string, error) {
	ret := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Fail to get addresses, err: %s", err)
		return nil, err
	}

	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok {
			if ip.IP.IsGlobalUnicast() {
				log.Printf("%s --- %s === %s", addr.String(), addr.Network(), ip.IP.String())
				ret = append(ret, ip.IP.String())
			}
		}
	}
	return ret, nil
}

func announceNodeInfo(client *etcd.Client, lease etcd.LeaseID, hostname string, addrs []string) {
	addrStr := ""
	for _, addr := range addrs {
		addrStr = fmt.Sprintf("%s%s,", addrStr, addr)
	}

	addrBytes := []byte(addrStr)
	addrBytes = addrBytes[:len(addrBytes)-1]

	addrStr = string(addrBytes)

	putKV(client, lease, fmt.Sprintf("/node/%s/addrs", hostname), addrStr)
}

func main() {

	hostname, err := getHostname()
	if err != nil {
		return
	}

	addrs, err := getUnicastIp()
	if err != nil {
		return
	}

	client, err := createClient()
	if err != nil {
		return
	}
	defer func(client *etcd.Client) {
		err := client.Close()
		if err != nil {
			log.Printf("Fail to close client, err: %s", err)
		}
	}(client)

	lease, err := grantAndKeepAliveLease(client)
	if err != nil {
		return
	}

	go watchKeyPrefix(client, "/node")

	time.Sleep(3 * time.Second)
	announceNodeInfo(client, lease, hostname, addrs)

	for {
		time.Sleep(3600 * time.Second)
	}
}
