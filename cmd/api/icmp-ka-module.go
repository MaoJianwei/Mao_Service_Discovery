package MaoApi

var (
	IcmpKaModuleRegisterName = "api-icmp-ka-module"
)

type IcmpKaModule interface {
	AddService(serviceIPv4v6 string)
	DelService(serviceIPv4v6 string)
}
