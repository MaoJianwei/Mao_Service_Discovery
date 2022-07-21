package Wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
)

func TestWechatMessageModule_SendWechatMessage(t *testing.T) {
	globalReceiversStr := ""
	globalReceivers := strings.Fields(globalReceiversStr)
	log.Println(globalReceivers)
}





type WechatAccessTokenResponse struct {
	ErrCode		int		`json:"errcode"`
	ErrMsg		string	`json:"errmsg"`
	AccessToken	string	`json:"access_token"`
	ExpiresIn	int		`json:"expires_in"`
}
func TestWechatMessageModule_SendWechatMessage2(t *testing.T) {

	req, err := http.NewRequest("GET", fmt.Sprintf(URL_TEMPLATE_GET_ACCESS_TOKEN,
		"", ""), nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	log.Println(body)

	accessTokenResponse := WechatAccessTokenResponse{}
	err = json.Unmarshal(body, &accessTokenResponse)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(body)
	log.Println(accessTokenResponse)

	if accessTokenResponse.ErrCode != 0 {
		log.Println(accessTokenResponse)
		return
	}
	// =====================================

	data :=
		"{" +
			"\"touser\":\"@all\"," +
			"\"toparty\":\"\"," +
			"\"totag\":\"\"," +
			"\"msgtype\":\"textcard\"," +
			"\"agentid\": 123456789," +
			"\"textcard\": {" +
			"\"title\":\"青岛雷达\"," +
			"\"description\":\"beijing<br>good day\"," +
			"\"url\":\"https://www.baidu.com/\"" +
			"}" +
			"}"
	postData := bytes.NewReader([]byte(data))
	req, err = http.NewRequest("POST", fmt.Sprintf(URL_TEMPLATE_SEND_MESSAGE, accessTokenResponse.AccessToken), postData)
	if err != nil {
		log.Println(err.Error())
		return
	}

	client = http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println(body)

}



