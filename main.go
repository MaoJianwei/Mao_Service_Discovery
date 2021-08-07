package main

import (
	"context"
	"fmt"
	etcd "go.etcd.io/etcd/client/v3"
	"net"
	"os"
	"time"
)

const (
	KEEP_ALIVE_TTL = 10 // Second
)

func putKV(client *etcd.Client, lease etcd.LeaseID, key, value string) {
	MaoLog(DEBUG, fmt.Sprintf("To put { %s -> %s }", key, value))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	putResp, err := client.Put(ctx, key, value, etcd.WithLease(lease))
	cancel()
	if err != nil {
		MaoLog(WARN, fmt.Sprintf("Fail to put { %s -> %s }, Resp: %s, err: %s", key, value, putResp, err))
		return
	}

	MaoLog(INFO, fmt.Sprintf("Success to put { %s -> %s }", key, value))
}

func watchKeyPrefix(client *etcd.Client, key string) {

	// Mao: If using this, it will close channel automatically.
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	ctx := context.Background()

	MaoLog(INFO, fmt.Sprintf("To watch: %s", key))
	watchChannel := client.Watch(ctx, key, etcd.WithPrefix())
	for watchResp := range watchChannel {
		MaoLog(DEBUG, fmt.Sprintf("New WatchResponse"))
		for _, ev := range watchResp.Events {
			MaoLog(INFO, fmt.Sprintf("New change, Type: %s, { %s -> %s }", ev.Type, ev.Kv.Key, ev.Kv.Value)) // "ev.PrevKv.Key nil", "ev.PrevKv.Value nil"
		}
	}

	MaoLog(INFO, fmt.Sprintf("Finish watch"))
}

// useless
func showLease(client *etcd.Client) {
	leasesResp, err := client.Leases(context.Background())
	if err != nil {
		MaoLog(ERROR, fmt.Sprintf("Fail to get leases, err: %s", err))
	}

	MaoLog(INFO, fmt.Sprintf("Found %d leases", len(leasesResp.Leases)))
	for _, leaseStatus := range leasesResp.Leases {
		MaoLog(INFO, fmt.Sprintf("Lease id: %d", leaseStatus.ID))
	}
}

func keepAliveLease(client *etcd.Client, id etcd.LeaseID) {
	MaoLog(INFO, fmt.Sprintf("Start keep-alive for lease %d ...", id))
	kaRespChan, err := client.KeepAlive(context.Background(), id)
	if err != nil {
		MaoLog(WARN, fmt.Sprintf("Fail to start keep-alive for lease %d", id))
		return
	}

	for kaResp := range kaRespChan {
		MaoLog(DEBUG, fmt.Sprintf("KA - ID: %d, TTL: %d", kaResp.ID, kaResp.TTL))
	}
}

func grantAndKeepAliveLease(client *etcd.Client) (etcd.LeaseID, error) {

	lgResp, err := client.Grant(context.Background(), KEEP_ALIVE_TTL)
	if err != nil {
		MaoLog(ERROR, fmt.Sprintf("Fail to grant a new lease, err: %s", err))
		return -1, err
	}

	MaoLog(INFO, fmt.Sprintf("Granted lease ID: %d, TTL: %d", lgResp.ID, lgResp.TTL))
	go keepAliveLease(client, lgResp.ID)

	return lgResp.ID, nil
}

func createClient() (*etcd.Client, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{"pi-dpdk.maojianwei.com:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		MaoLog(ERROR, fmt.Sprintf("Fail to generate client, err: %s", err))
		return nil, err
	}
	return client, nil
	//defer client.Close()

}

func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		MaoLog(ERROR, fmt.Sprintf("Fail to get hostname"))
		return "", err
	}
	MaoLog(INFO, fmt.Sprintf("Hostname: %s", hostname))
	return hostname, nil
}

func getUnicastIp() ([]string, error) {
	ret := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		MaoLog(ERROR, fmt.Sprintf("Fail to get addresses, err: %s", err))
		return nil, err
	}

	for i, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok {
			if ip.IP.IsGlobalUnicast() {
				MaoLog(DEBUG, fmt.Sprintf("IP %d: %s --- %s === %s", i, addr.String(), addr.Network(), ip.IP.String()))
				MaoLog(INFO, fmt.Sprintf("IP %d: %s", i, ip.IP.String()))
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
			MaoLog(WARN, fmt.Sprintf("Fail to close client, err: %s", err))
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
