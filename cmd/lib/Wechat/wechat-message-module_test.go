package Wechat

import (
	"log"
	"strings"
	"testing"
)

func TestWechatMessageModule_SendWechatMessage(t *testing.T) {
	globalReceiversStr := ""
	globalReceivers := strings.Fields(globalReceiversStr)
	log.Println(globalReceivers)
}