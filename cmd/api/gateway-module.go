package MaoApi

var (
	GatewayModuleRegisterName = "gateway-module"
)

const (
	GATEWAY_MEASUREMENT = "Gateway"
	GATEWAY_TAG_GEO = "Geo"
	GATEWAY_FIELD_BytesReceivedSpeed = "BytesReceivedSpeed"
	GATEWAY_FIELD_BytesReceived = "BytesReceived"
	GATEWAY_FIELD_BytesSentSpeed = "BytesSentSpeed"
	GATEWAY_FIELD_BytesSent = "BytesSent"
	GATEWAY_FIELD_PacketsReceivedSpeed = "PacketsReceivedSpeed"
	GATEWAY_FIELD_PacketsReceived = "PacketsReceived"
	GATEWAY_FIELD_PacketsSentSpeed = "PacketsSentSpeed"
	GATEWAY_FIELD_PacketsSent = "PacketsSent"
	GATEWAY_FIELD_Uptime = "Uptime"
)

type GatewayModule interface {
	//SendEmail(message *EmailMessage)
}