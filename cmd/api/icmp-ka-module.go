package MaoApi

import "time"

var (
	IcmpKaModuleRegisterName = "api-icmp-ka-module"
)

type MaoIcmpService struct {
	Address string

	Alive    bool
	LastSeen time.Time

	DetectCount uint64
	ReportCount uint64

	RttDuration          time.Duration
	RttOutboundTimestamp time.Time
}

type IcmpKaModule interface {
	AddService(serviceIPv4v6 string)
	DelService(serviceIPv4v6 string)
	GetServices() []*MaoIcmpService
}
