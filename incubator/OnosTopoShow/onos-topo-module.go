package OnosTopoShow

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/MaoCommon"
	"MaoServerDiscovery/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	MODULE_NAME = "ONOS-Topology-module"

	ADD_DEVICE_API_SUFFIX = "/MaoIntegration/addDevice"
	REMOVE_DEVICE_API_SUFFIX = "/MaoIntegration/removeDevice"
	ADD_LINK_API_SUFFIX = "/MaoIntegration/addBiLink" // bidirectional link
	REMOVE_LINK_API_SUFFIX = "/MaoIntegration/removeBiLink" // bidirectional link

	URL_ONOS_HOMEPAGE = "/configOnos"
	URL_ONOS_CONFIG   = "/addOnosInfo"
	URL_ONOS_SHOW   = "/getOnosInfo"

	ONOS_CONFIG_PATH     = "/ONOS"
	ONOS_API_CONFIG_PATH = "/ONOS/APIs"
)

type OnosTopoModule struct {

	/* Local info */
	hostname string
	version string

	/* API Config */
	addrPort			string
	ADD_DEVICE_API		string
	REMOVE_DEVICE_API	string
	DELETE_DEVICE_API	string
	ADD_LINK_API		string
	REMOVE_LINK_API		string

	/* Internal */
	portIterators	map[string]uint
	portMapping		map[string]uint // local-remote-protocol => local port number

	needShutdown bool
	topoEventChannel chan *MaoApi.TopoEvent
}

func (o *OnosTopoModule) RequireShutdown() {
	o.needShutdown = true
}

func (o *OnosTopoModule) InitOnosTopoModule(hostname string, version string) bool {
	o.hostname = hostname
	o.version = version
	o.needShutdown = false
	o.topoEventChannel = make(chan *MaoApi.TopoEvent, 1024)

	o.portIterators = make(map[string]uint)
	o.portIterators[hostname] = 1

	o.portMapping = make(map[string]uint)

	o.configRestControlInterface()

	go o.topoEventLoop()
	return true
}





func (o *OnosTopoModule) configRestControlInterface() {
	restfulServer := MaoCommon.ServiceRegistryGetRestfulServerModule()
	if restfulServer == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get RestfulServerModule, unable to register restful apis.")
		return
	}

	restfulServer.RegisterGetApi(URL_ONOS_HOMEPAGE, o.showOnosPage)
	restfulServer.RegisterGetApi(URL_ONOS_SHOW, o.showOnosInfo)
	restfulServer.RegisterPostApi(URL_ONOS_CONFIG, o.processOnosInfo)
}

func (o *OnosTopoModule) showOnosPage(c *gin.Context) {
	c.HTML(200, "index-onos.html", nil)
}

func (o *OnosTopoModule) showOnosInfo(c *gin.Context) {
	// TODO: TEST
	data := make(map[string]interface{})
	data["addrPort"] = o.addrPort
	data["ADD_DEVICE_API"] = o.ADD_DEVICE_API
	data["REMOVE_DEVICE_API"] = o.REMOVE_DEVICE_API
	data["DELETE_DEVICE_API"] = o.DELETE_DEVICE_API
	data["ADD_LINK_API"] = o.ADD_LINK_API
	data["REMOVE_LINK_API"] = o.REMOVE_LINK_API

	// Attention: password can't be outputted !!!
	c.JSON(200, data)
}

func (o *OnosTopoModule) processOnosInfo(c *gin.Context) {
	// TODO: TEST

	addrPort, ok := c.GetPostForm("addrPort")
	if ok {
		o.addrPort = addrPort
		o.configOnosEndpointAPI(addrPort)
	}

	configModule := MaoCommon.ServiceRegistryGetConfigModule()
	if configModule == nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get config module instance, can't save email info")
	} else {
		data := make(map[string]interface{})
		data["addrPort"] = o.addrPort
		data["ADD_DEVICE_API"] = o.ADD_DEVICE_API
		data["REMOVE_DEVICE_API"] = o.REMOVE_DEVICE_API
		data["DELETE_DEVICE_API"] = o.DELETE_DEVICE_API
		data["ADD_LINK_API"] = o.ADD_LINK_API
		data["REMOVE_LINK_API"] = o.REMOVE_LINK_API

		configModule.PutConfig(ONOS_CONFIG_PATH, data)
	}

	o.showOnosPage(c)
}

func (o *OnosTopoModule) configOnosEndpointAPI(addrPort string) {
	o.ADD_DEVICE_API = fmt.Sprintf("http://%s/onos/mao%s", addrPort, ADD_DEVICE_API_SUFFIX)
	o.REMOVE_DEVICE_API = fmt.Sprintf("http://%s/onos/mao%s", addrPort, REMOVE_DEVICE_API_SUFFIX)
	o.DELETE_DEVICE_API =  fmt.Sprintf("http://%s/onos/v1/devices/mao:%%s", addrPort)
	o.ADD_LINK_API = fmt.Sprintf("http://%s/onos/mao%s", addrPort, ADD_LINK_API_SUFFIX)
	o.REMOVE_LINK_API = fmt.Sprintf("http://%s/onos/mao%s", addrPort, REMOVE_LINK_API_SUFFIX)

	// Add local node to ONOS after updating the API URLs
	// Init my local node to ONOS instance.
	o.topoAddDevice(o.hostname, o.version, time.Now().String())
}



//func (o *OnosTopoModule) initConfig() (success bool) {
//	apiConfig := o.getApiConfig()
//	if apiConfig != nil {
//		// TODO: confirm and apply config
//		return true
//	}
//
//	// the config doesn't exist, init it.
//
//	configModule := MaoCommon.ServiceRegistryGetConfigModule()
//	if configModule == nil {
//		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to get config module instance")
//		return false
//	}
//
//	_, errCode := configModule.PutConfig(ONOS_API_CONFIG_PATH, make([]string, 0))
//	if errCode != Config.ERR_CODE_SUCCESS {
//		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to init ONOS_API_CONFIG_PATH, errCode: %d", errCode)
//		return false
//	}
//
//	return true
//}



func (o *OnosTopoModule) SendEvent(event *MaoApi.TopoEvent) {
	o.topoEventChannel <- event
}

func (o *OnosTopoModule) topoEventLoop() {
	kaInterval := time.Duration(1000) * time.Millisecond
	kaShutdownTimer := time.NewTimer(kaInterval)

	// Init my local node to ONOS instance.
	o.topoAddDevice(o.hostname, o.version, time.Now().String())
	for {
		select {
		case event := <-o.topoEventChannel:
			qingdao := len(o.topoEventChannel)
			util.MaoLogM(util.DEBUG, MODULE_NAME, "buffer len: %d", qingdao)
			switch event.EventType {
			case MaoApi.SERVICE_UP:
				localPort, ok1 := o.portMapping[fmt.Sprintf("%s-%s-%s", o.hostname, event.ServiceName, event.EventSource)]
				if !ok1 {
					localPort = o.portIterators[o.hostname]
					o.portIterators[o.hostname] = localPort + 1
					o.portMapping[fmt.Sprintf("%s-%s-%s", o.hostname, event.ServiceName, event.EventSource)] = localPort
				}

				servicePort, ok2 := o.portMapping[fmt.Sprintf("%s-%s-%s", event.ServiceName, o.hostname, event.EventSource)]
				if !ok2 {
					servicePort, ok2 = o.portIterators[event.ServiceName]
					if !ok2 {
						servicePort = 1
					}
					o.portIterators[event.ServiceName] = servicePort + 1
					o.portMapping[fmt.Sprintf("%s-%s-%s", event.ServiceName, o.hostname, event.EventSource)] = servicePort
				}

				go func() {
					o.topoAddDevice(event.ServiceName, event.Timestamp.String(), event.EventSource)
					o.topoAddLink(o.hostname, localPort, event.ServiceName, servicePort)
				}()
			case MaoApi.SERVICE_DOWN:
				go func(){
					o.topoAddDevice(event.ServiceName, event.Timestamp.String(), event.EventSource)
					o.topoOfflineDevice(event.ServiceName)
				}()
			case MaoApi.SERVICE_DELETE:
				go o.topoDeleteDevice(event.ServiceName)
			}
		case <-kaShutdownTimer.C:
			if o.needShutdown {
				if len(o.topoEventChannel) != 0 {
					util.MaoLogM(util.WARN, MODULE_NAME, "Exiting, but the topoEventChannel is not empty, len: %d", len(o.topoEventChannel))
				}
				util.MaoLogM(util.INFO, MODULE_NAME, "Exit.")
				return
			}
			go o.topoAddDevice(o.hostname, o.version, time.Now().String())
			kaShutdownTimer.Reset(kaInterval)
		}
	}
}

func (o *OnosTopoModule) topoAddDevice(serviceName string, timestamp string, eventSource string) {
	if serviceName == "" {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to add device, service name can't be empty")
		return
	}

	jsonMap := make(map[string]string)
	jsonMap["deviceId"] = serviceName
	jsonMap["deviceName"] = serviceName
	jsonMap["swVersion"] = timestamp
	jsonMap["manageProtocol"] = eventSource
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to marshal device data, %s", err.Error())
		return
	}

	if sendRequest("POST", o.ADD_DEVICE_API, jsonBytes) {
		//routerDatas = append(routerDatas, jsonBytes)
		//routerNames = append(routerNames, router.deviceName)
		util.MaoLogM(util.INFO, MODULE_NAME, "Add device: %s", serviceName)
	}
}

func (o *OnosTopoModule) topoOfflineDevice(serviceName string) {
	if serviceName == "" {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to offline device, service name can't be empty")
		return
	}

	jsonMap := make(map[string]string)
	jsonMap["deviceId"] = serviceName
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to marshal device data, %s", err.Error())
		return
	}

	if sendRequest("POST", o.REMOVE_DEVICE_API, jsonBytes) {
		util.MaoLogM(util.INFO, MODULE_NAME, "Offline device: %s", serviceName)
	}
}

func (o *OnosTopoModule) topoDeleteDevice(serviceName string) {
	if serviceName == "" {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to delete device, service name can't be empty")
		return
	}

	if sendRequest("DELETE", fmt.Sprintf(o.DELETE_DEVICE_API, serviceName), nil) {
		util.MaoLogM(util.INFO, MODULE_NAME, "Delete device: %s", serviceName)
	}
}

func (o *OnosTopoModule) topoAddLink(serviceName1 string, portId1 uint, serviceName2 string, portId2 uint) {
	if serviceName1 == "" || serviceName2 == ""{
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to add link, service name can't be empty, %s - %s", serviceName1, serviceName2)
		return
	}

	jsonMap := make(map[string]interface{})
	jsonMap["srcDeviceId"] = serviceName1
	jsonMap["srcPortId"] = portId1
	jsonMap["srcPortName"] = fmt.Sprintf("%s-%s", serviceName1, serviceName2)
	jsonMap["dstDeviceId"] = serviceName2
	jsonMap["dstPortId"] = portId2
	jsonMap["dstPortName"] = fmt.Sprintf("%s-%s", serviceName2, serviceName1)

	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to marshal link data, %s", err.Error())
		return
	}

	if sendRequest("POST", o.ADD_LINK_API, jsonBytes) {
		//linkDatas = append(linkDatas, jsonBytes)
		//linkNames = append(linkNames, GenerateLinkName(r1.deviceName, r2.deviceName))
		util.MaoLogM(util.INFO, MODULE_NAME, "Add link: %s - %s", serviceName1, serviceName2)
	}
}

func sendRequest(method string, url string, body []byte) bool {
	if url == "" {
		util.MaoLogM(util.DEBUG, MODULE_NAME, "API URL is not configured, not send request.")
		return false
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to create request: %s %s, %s",
			method, url, err.Error())
		return false
	}

	// ONOS Web default password, karaf:karaf, basic authentication
	req.Header.Add("Authorization", "Basic a2FyYWY6a2FyYWY=")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to do request: %s %s, err: %s",
			req.Method, req.URL.String(), err.Error())
		return false
	}

	if resp.StatusCode != 200 {
		util.MaoLogM(util.WARN, MODULE_NAME, "Fail to finish request: %s %s, http code: %d, err: %s",
			req.Method, req.URL.String(), resp.StatusCode, resp.Status)
		return false
	}

	return true
}
