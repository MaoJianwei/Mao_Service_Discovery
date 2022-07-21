package Wechat

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

const (
	MODULE_NAME = "Wechat-Message-module"

	URL_WECHAT_HOMEPAGE = "/configWechat"
	URL_WECHAT_CONFIG   = "/addWechatInfo"
	URL_WECHAT_SHOW   = "/getWechatInfo"

	WECHAT_INFO_CONFIG_PATH = "/wechat"
)

const (

	URL_TEMPLATE_GET_ACCESS_TOKEN = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"
	URL_TEMPLATE_SEND_MESSAGE = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s"

	TEXT_CARD_JSON_TEMPLATE =
		"{" +
			"\"touser\":\"%s\"," +
			"\"toparty\":\"\"," +
			"\"totag\":\"\"," +
			"\"msgtype\":\"textcard\"," +
			"\"agentid\": %s," +
			"\"textcard\": {" +
					"\"title\":\"%s\"," +
					"\"description\":\"%s\"," +
					"\"url\":\"%s\"" +
			"}" +
		"}"
)

type WechatMessageModule struct {

	corpId		string
	agentId		string
	agentSecret	string // corpsecret. Attention: agentSecret can't be outputted !!!
	globalReceivers []string  // length == 0 and nil stand for all receivers.


	// input the message
	sendWechatMessageChannel chan *MaoApi.WechatMessage
	lastSendTimestamp        time.Time

	// same as the other module, it is expected to be global
	//checkInterval uint32

	needShutdown bool
}

func (w *WechatMessageModule) RequireShutdown() {
	w.needShutdown = true
}


func (w *WechatMessageModule) SendWechatMessage(message *MaoApi.WechatMessage) {
	w.sendWechatMessageChannel <- message
}

func (w *WechatMessageModule) checkWechatInfo() bool {

	if w.corpId == "" || w.agentId == "" || w.agentSecret == ""  {
		return false
	}
	return true
}


//func (w *WechatMessageModule) getAccessToken;

func (w *WechatMessageModule) sendWechatMessage(m *MaoApi.WechatMessage) {

	if !w.checkWechatInfo() {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to send wechat message, please config wechat info first.")
		return
	}

	receiverConfig := w.globalReceivers
	if m.Receivers != nil {
		receiverConfig = m.Receivers
	}

	receivers := ""
	if len(receiverConfig) == 0 {
		receivers = "@all"
	} else {
		receivers = fmt.Sprintf("%s%s", receivers, receiverConfig[0])
		for i := 1; i < len(receiverConfig); i++ {
			receivers = fmt.Sprintf("%s%s", receivers, "|")
			receivers = fmt.Sprintf("%s%s", receivers, receiverConfig[i])
		}
	}

	wechatJson := fmt.Sprintf(TEXT_CARD_JSON_TEMPLATE, receivers, w.agentId, m.Title, m.ContentHttp, m.Url)
	util.MaoLogM(util.HOT_DEBUG, MODULE_NAME, "Prepare wechat json:\n%s", wechatJson)
	

	//http.Get(fmt.Sprintf(URL_TEMPLATE_GET_ACCESS_TOKEN, ))


	//req := http.Post("https://qyapi.weixin.qq.com/cgi-bin/message/send", )
	
	return

	// TODO

	//// Set up authentication information everytime, that allows to update:
	//// username, password, smtpServerAddrPort, sender, receiver
	//auth := AuthLOGIN(s.username, s.password)
	//
	//// TODO: how to build multiple receivers string?
	//// TODO: support multiple receivers
	//msg := fmt.Sprintf("To: %s\r\n" +
	//	"Subject: %s%s\r\n" + "\r\n" +
	//	"%s\r\n",
	//	s.receiver[0], SUBJECT_FIX_PREFIX, m.Subject, m.Content)
	//
	////msg := []byte("To: @.com\r\n" +
	////	"Subject: MaoReport: beijing tower\r\n" +
	////	"\r\n" +
	////	"This is the email body.\r\n")
	//
	//// TODO: make it configurable
	//tlsConfig := &tls.Config{
	//	InsecureSkipVerify: true,
	//}
	//
	//// Connect to the server, authenticate, set the sender and recipient,
	//// and send the email all in one step.
	//err := MaoEnhancedGolang.SendMail(s.smtpServerAddrPort, auth, s.sender, s.receiver, []byte(msg), tlsConfig)
	//if err != nil {
	//	util.MaoLogM(util.WARN, MODULE_NAME, "Fail to send email, %s", err.Error())
	//}
}

func (w *WechatMessageModule) sendWechatMessageLoop() {
	checkInterval := time.Duration(1000) * time.Millisecond
	checkShutdownTimer := time.NewTimer(checkInterval)
	for {
		select {
		case message := <-w.sendWechatMessageChannel:
			if time.Now().Sub(w.lastSendTimestamp) < 1 * time.Second {
				freezeTimer := time.NewTimer(time.Duration(1) * time.Second)
				sent := false
				for !sent {
					select {
					case <-freezeTimer.C:
						util.MaoLogM(util.INFO, MODULE_NAME, "Sending wechat message, pending: %d", len(w.sendWechatMessageChannel))
						w.sendWechatMessage(message)
						w.lastSendTimestamp = time.Now()
						sent = true
					case <-checkShutdownTimer.C:
						util.MaoLogM(util.HOT_DEBUG, MODULE_NAME, "CheckShutdown while freezing, event queue len %d", len(w.sendWechatMessageChannel))
						if w.needShutdown {
							util.MaoLogM(util.WARN, MODULE_NAME, "Exit while freezing, the sendWechatMessageChannel len: %d", len(w.sendWechatMessageChannel))
							return
						}
						checkShutdownTimer.Reset(checkInterval)
					}
				}
			} else {
				util.MaoLogM(util.INFO, MODULE_NAME, "Sending wechat message, pending: %d", len(w.sendWechatMessageChannel))
				w.sendWechatMessage(message)
				w.lastSendTimestamp = time.Now()
			}
		case <-checkShutdownTimer.C:
			util.MaoLogM(util.HOT_DEBUG, MODULE_NAME, "CheckShutdown, event queue len %d", len(w.sendWechatMessageChannel))
			if w.needShutdown {
				if len(w.sendWechatMessageChannel) != 0 {
					util.MaoLogM(util.WARN, MODULE_NAME, "Exiting, but the sendWechatMessageChannel is not empty, len: %d", len(w.sendWechatMessageChannel))
				}
				util.MaoLogM(util.INFO, MODULE_NAME, "Exit.")
				return
			}
			checkShutdownTimer.Reset(checkInterval)
		}
	}
}

func (w *WechatMessageModule) InitWechatMessageModule() bool {
	w.sendWechatMessageChannel = make(chan *MaoApi.WechatMessage, 1024)
	w.needShutdown = false

	go w.sendWechatMessageLoop()

	w.configRestControlInterface()

	return true
}

func (w *WechatMessageModule) configRestControlInterface() {
	restfulServer := MaoCommon.ServiceRegistryGetRestfulServerModule()
	if restfulServer == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get RestfulServerModule, unable to register restful apis.")
		return
	}

	restfulServer.RegisterGetApi(URL_WECHAT_HOMEPAGE, w.showWechatPage)
	restfulServer.RegisterGetApi(URL_WECHAT_SHOW, w.showWechatInfo)
	restfulServer.RegisterPostApi(URL_WECHAT_CONFIG, w.processWechatInfo)
}

func (w *WechatMessageModule) showWechatPage(c *gin.Context) {
	c.HTML(200, "index-wechat.html", nil)
}

func (w *WechatMessageModule) showWechatInfo(c *gin.Context) {
	data := make(map[string]interface{})
	data["corpId"] = w.corpId
	data["agentId"] = w.agentId
	data["globalReceivers"] = w.globalReceivers

	// Attention: agentSecret can't be outputted !!!
	c.JSON(200, data)
}

func (w *WechatMessageModule) processWechatInfo(c *gin.Context) {

	// TODO: check and limit the length of corpId/agentId/agentSecret. Prevent injection attack

	corpId, ok := c.GetPostForm("corpId")
	if ok {
		w.corpId = corpId
	}

	agentId, ok := c.GetPostForm("agentId")
	if ok {
		w.agentId = agentId
	}

	agentSecret, ok := c.GetPostForm("agentSecret")
	if ok {
		w.agentSecret = agentSecret
	}

	globalReceiversStr, ok := c.GetPostForm("globalReceivers")
	if ok {
		globalReceivers := strings.Fields(globalReceiversStr)
		w.globalReceivers = globalReceivers
	}


	configModule := MaoCommon.ServiceRegistryGetConfigModule()
	if configModule == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get config module instance, can't save wechat info")
	} else {
		data := make(map[string]interface{})
		data["corpId"] = w.corpId
		data["agentId"] = w.agentId
		data["globalReceivers"] = w.globalReceivers

		// Attention: agentSecret can't be outputted !!!
		configModule.PutConfig(WECHAT_INFO_CONFIG_PATH, data)
	}

	w.showWechatPage(c)
}