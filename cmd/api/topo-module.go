package MaoApi

import "time"

type EventType int

const (
	SOURCE_GRPC = "gRPC"
	SOURCE_ICMP = "ICMP"
)
const (
	SERVICE_UP EventType = iota + 1
	SERVICE_DOWN
	SERVICE_DELETE
)

var (
	TopoModuleRegisterName = "onos-topo-module"
)

type TopoEvent struct {
	EventType EventType
	EventSource string

	ServiceName string
	Timestamp time.Time
}

type TopoModule interface {
	SendEvent(event *TopoEvent)
}
