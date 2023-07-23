package MaoApi

import (
	"time"
)

var (
	GrpcKaModuleRegisterName = "api-grpc-ka-module"
)

type GrpcServiceNode struct {
	Hostname    string
	ReportTimes uint64

	Ips            []string
	RealClientAddr string

	OtherData string

	ServerDateTime string
	LocalLastSeen  time.Time
	Alive bool

	RttDuration time.Duration // nanosecond, uint64
}

type GrpcKaModule interface {
	GetServiceInfo() []*GrpcServiceNode
}