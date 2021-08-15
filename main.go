package main
//
//import (
//	"MaoServerDiscovery/util"
//	"context"
//	"fmt"
//	etcd "go.etcd.io/etcd/client/v3"
//	"time"
//)
//
//const (
//	KEEP_ALIVE_TTL = 10 // Second
//)
//
//var (
//	last_hostname, last_addrStr string
//)
//
//func putKV(client *etcd.Client, lease etcd.LeaseID, key, value string) {
//	util.MaoLog(util.DEBUG, fmt.Sprintf("To put { %s -> %s }", key, value))
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	putResp, err := client.Put(ctx, key, value, etcd.WithLease(lease))
//	cancel()
//	if err != nil {
//		util.MaoLog(util.WARN, fmt.Sprintf("Fail to put { %s -> %s }, Resp: %s, err: %s", key, value, putResp, err))
//		return
//	}
//
//	util.MaoLog(util.INFO, fmt.Sprintf("Success to put { %s -> %s }", key, value))
//}
//
//func watchKeyPrefix(client *etcd.Client, key string) {
//
//	// Mao: If using this, it will close channel automatically.
//	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
//	ctx := context.Background()
//
//	util.MaoLog(util.INFO, fmt.Sprintf("To watch: %s", key))
//	watchChannel := client.Watch(ctx, key, etcd.WithPrefix())
//	for watchResp := range watchChannel {
//		util.MaoLog(util.DEBUG, fmt.Sprintf("New WatchResponse"))
//		for _, ev := range watchResp.Events {
//			util.MaoLog(util.INFO, fmt.Sprintf("New change, Type: %s, { %s -> %s }", ev.Type, ev.Kv.Key, ev.Kv.Value)) // "ev.PrevKv.Key nil", "ev.PrevKv.Value nil"
//		}
//	}
//
//	util.MaoLog(util.INFO, fmt.Sprintf("Finish watch"))
//}
//
//// useless
//func showLease(client *etcd.Client) {
//	leasesResp, err := client.Leases(context.Background())
//	if err != nil {
//		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to get leases, err: %s", err))
//	}
//
//	util.MaoLog(util.INFO, fmt.Sprintf("Found %d leases", len(leasesResp.Leases)))
//	for _, leaseStatus := range leasesResp.Leases {
//		util.MaoLog(util.INFO, fmt.Sprintf("Lease id: %d", leaseStatus.ID))
//	}
//}
//
//func keepAliveLease(client *etcd.Client, id etcd.LeaseID) {
//	util.MaoLog(util.INFO, fmt.Sprintf("Start keep-alive for lease %d ...", id))
//	kaRespChan, err := client.KeepAlive(context.Background(), id)
//	if err != nil {
//		util.MaoLog(util.WARN, fmt.Sprintf("Fail to start keep-alive for lease %d", id))
//		return
//	}
//
//	for kaResp := range kaRespChan {
//		util.MaoLog(util.DEBUG, fmt.Sprintf("KA - ID: %d, TTL: %d", kaResp.ID, kaResp.TTL))
//	}
//}
//
//func grantAndKeepAliveLease(client *etcd.Client) (etcd.LeaseID, error) {
//
//	lgResp, err := client.Grant(context.Background(), KEEP_ALIVE_TTL)
//	if err != nil {
//		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to grant a new lease, err: %s", err))
//		return -1, err
//	}
//
//	util.MaoLog(util.INFO, fmt.Sprintf("Granted lease ID: %d, TTL: %d", lgResp.ID, lgResp.TTL))
//	go keepAliveLease(client, lgResp.ID)
//
//	return lgResp.ID, nil
//}
//
//func createClient() (*etcd.Client, error) {
//	client, err := etcd.New(etcd.Config{
//		Endpoints:   []string{"10.107.10.155:22379"},
//		DialTimeout: 3 * time.Second,
//	})
//	if err != nil {
//		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to generate client, err: %s", err))
//		return nil, err
//	}
//	return client, nil
//	//defer client.Close()
//
//}
//
//func announceNodeInfo(client *etcd.Client, lease etcd.LeaseID, hostname string, addrs []string) {
//	addrStr := ""
//	for _, addr := range addrs {
//		addrStr = fmt.Sprintf("%s%s,", addrStr, addr)
//	}
//
//	addrBytes := []byte(addrStr)
//	addrBytes = addrBytes[:len(addrBytes)-1]
//
//	addrStr = string(addrBytes)
//
//	if last_hostname != hostname || last_addrStr != addrStr {
//		putKV(client, lease, fmt.Sprintf("/node/%s/addrs", hostname), addrStr)
//	}
//	putKV(client, lease, fmt.Sprintf("/node/%s/lastseen", hostname), time.Now().String())
//
//	last_hostname = hostname
//	last_addrStr = addrStr
//}
//
//func main() {
//	client, err := createClient()
//	if err != nil {
//		return
//	}
//	defer func(client *etcd.Client) {
//		err := client.Close()
//		if err != nil {
//			util.MaoLog(util.WARN, fmt.Sprintf("Fail to close client, err: %s", err))
//		}
//	}(client)
//
//	lease, err := grantAndKeepAliveLease(client)
//	if err != nil {
//		return
//	}
//
//	go watchKeyPrefix(client, "/node")
//	time.Sleep(1 * time.Second)
//
//	for {
//		hostname, err := util.GetHostname()
//		if err != nil {
//			return
//		}
//
//		addrs, err := util.GetUnicastIp()
//		if err != nil {
//			return
//		}
//
//		announceNodeInfo(client, lease, hostname, addrs)
//
//		time.Sleep(5 * time.Second)
//	}
//}
