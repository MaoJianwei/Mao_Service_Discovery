package MaoApi

var (
	WechatModuleRegisterName = "wechat-module"
)

type WechatMessage struct {

	Receivers []string  // length == 0 stands for all receivers; nil stands for unspecified.

	Title string
	ContentHttp string
	Url string
}

type WechatModule interface {
	SendWechatMessage(message *WechatMessage)
}