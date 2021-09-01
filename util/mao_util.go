package util

import (
	"fmt"
	"net"
	"os"
)

func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		MaoLog(ERROR, fmt.Sprintf("Fail to get hostname"))
		return "", err
	}
	MaoLog(DEBUG, fmt.Sprintf("Hostname: %s", hostname))
	return hostname, nil
}

func GetUnicastIp() ([]string, error) {
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
				MaoLog(DEBUG, fmt.Sprintf("IP %d: %s", i, ip.IP.String()))
				ret = append(ret, ip.IP.String())
			}
		}
	}
	return ret, nil
}

func JudgeIPv6(ip *net.IP) bool {
	return ip.To4() == nil
}

func GetAddrPort(addr *net.IP, port uint32) string {
	if JudgeIPv6(addr) {
		return fmt.Sprintf("[%s]:%d", addr.String(), port)
	} else {
		return fmt.Sprintf("%s:%d", addr.String(), port)
	}
}