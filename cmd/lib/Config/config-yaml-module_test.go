package Config

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestConfigYamlModule_main(t *testing.T) {

	log.Println(os.Args)

	configModule := &ConfigYamlModule{
	}

	if !configModule.InitConfigModule(DEFAULT_CONFIG_FILE) {
		return
	}

	mmm := make(map[string]interface{})
	mmm["beijing"] = "ra11111dar"
	mmm["intintint"] = 7181
	mmm["fff"] = 2.525

	vvv := make(map[int]interface{})
	vvv[6666] = "radar"
	vvv[8888] = 5511
	vvv[7181] = 2.525


	value, errCode := configModule.GetConfig("/qingdao/radar/freq") // ok
	log.Printf("Put 1 %v, %v\n", value, errCode)

	value, errCode = configModule.GetConfig("/qingdao/radar/name") // ok
	log.Printf("Put 2 %v, %v\n", value, errCode)

	value, errCode = configModule.GetConfig("/qingdao/name") // ok
	log.Printf("Put 3 %v, %v\n", value, errCode)

	value, errCode = configModule.GetConfig("/config/module/instance/object") // bad
	log.Printf("Put 4 %v, %v\n", value, errCode)

	value, errCode = configModule.GetConfig("/config/module/instance/mmm") // ok
	log.Printf("Put 5 %v, %v\n", value, errCode)

	value, errCode = configModule.GetConfig("/config/module/instance/vvv") // ok
	log.Printf("Put 6 %v, %v\n", value, errCode)


	b, v := configModule.PutConfig("/qingdao/radar/freq", 118.5) // ok
	log.Printf("Put 1 %v, %v\n", b, v)

	b, v = configModule.PutConfig("/qingdao/radar/name", "qingdao") // ok
	log.Printf("Put 2 %v, %v\n", b, v)

	b, v = configModule.PutConfig("/qingdao/name", "liuting") // ok
	log.Printf("Put 3 %v, %v\n", b, v)

	b, v = configModule.PutConfig("/config/module/instance/object", configModule) // bad
	log.Printf("Put 4 %v, %v\n", b, v)

	b, v = configModule.PutConfig("/config/module/instance/mmm", mmm) // ok
	log.Printf("Put 5 %v, %v\n", b, v)

	b, v = configModule.PutConfig("/config/module/instance/vvv", vvv) // ok
	log.Printf("Put 6 %v, %v\n", b, v)

	configModule.RequireShutdown()

	time.Sleep(3 * time.Second)





	/*// ============================ Develop & Unit tests =============================
	config := make(map[string]interface{})
	tmp := make(map[string]interface{})
	config["qingdao"] = tmp

	tmp1 := make(map[string]interface{})
	tmp["radar"] = tmp1

	tmp2 := make(map[string]interface{})
	tmp1["118.5"] = tmp2

	tmp2["beijing-2"] = "qingdaoleida beijing bigmao"


	//GET Unit Tests:
	//path := "/" // ERR_CODE_PATH_FORMAT - ok
	//path := "/qingdao" // pass - ok
	//path := "/qingdao/" // ERR_CODE_PATH_FORMAT - ok
	//
	//path := "/qingdao/radar" // pass - ok
	//path := "/qingdao/radar/" // ERR_CODE_PATH_FORMAT - ok
	//
	//path := "/qingdao/radar/118.5" // pass - ok
	//path := "/qingdao/radar/118.5/" // ERR_CODE_PATH_FORMAT - ok
	//
	//path := "/qingdao/radar/218.5" // pass but result is nil - ok
	//path := "/qingdao/padar/118.5" // ERR_CODE_PATH_TRANSIT_FAIL - ok
	//
	//path := "/qingdao/radar/118.5/beijing-2" // pass - ok
	//path := "/qingdao/radar/118.5/beijing-2/" // ERR_CODE_PATH_FORMAT - ok

	//Put Unit Tests:
	//var newData interface{} = "bigmao radio station"
	//path := "/qingdao" // pass - ok
	//path := "/qingdao/radar" // pass - ok
	//path := "/qingdao/radar/118.5" // pass - ok
	//path := "/qingdao/radar/218.5" // pass - ok
	//path := "/qingdao/padar/118.5" // pass and create all transit path - ok
	//path := "/qingdao/padar/118.5/aaaa/b/1/5a" // pass and create all transit path - ok

	//Put Unit Tests - nil:
	//var newData interface{} = nil
	//path := "/qingdao" // pass and delete all, result is an empty map - ok
	//path := "/qingdao/radar" // pass and delete radar & 118.5 - ok
	//path := "/qingdao/radar/118.5" // pass and delete 118.5 - ok
	//path := "/qingdao/padar/118.5" // pass and: create all transit path, the last is an empty map of padar - ok

	log.Println("Origin config: ", config)

	// Mao: Design Principle: it is not allowed that the config contain nil.

	paths := strings.Split(path, "/")
	if paths[0] != "" || paths[len(paths)-1] == "" {
		log.Println("eventResult: ", eventResult{
			errCode: ERR_CODE_PATH_FORMAT,
			result:  nil,
		}) // event.result <-
		util.MaoLogM(util.WARN, MODULE_NAME, "format of config path is not correct.")
		return // continue
	}

	transitPaths := paths[1 : len(paths)-1] // [a, b)
	transitConfig := config
	var ok = true

	var missPos int
	var tmpPath string
	for missPos, tmpPath = range transitPaths {
		tmpObj := transitConfig[tmpPath]
		if tmpObj == nil {
			// We meet a nonexistent path, or the data is nil.

			// Get operation: fail. (nonexistent path / data is nil)
			// Put operation: need to create all transit path to store the data. (nonexistent path / data is nil)
			// Put operation: if nil is in the config or the new data is nil, we will remove it or override it automatically,
			//                because it is not allowed that the config contain nil.
			ok = false
			break
		}

		// if obj is not map, get (nil, false) --- if it is Put operation, you need to Put nil to delete the stale data first, then retry to Put data again.
		// if obj is nil, get (nil, false) --- avoid by the above --- And, it is not allowed that the config contain nil.
		// if obj is not exist, get (nil, false) --- avoid by the above
		transitConfig, ok = tmpObj.(map[string]interface{})
		if !ok {
			// Put Operation: there is valid data, we can not override it automatically.
			// Get Operation: we cannot transit forward anymore

			log.Println("in the for, ", eventResult{
				errCode: ERR_CODE_PATH_TRANSIT_FAIL,
				result:  nil,
			})     // event.result <-
			return // continue
		}
		log.Println(transitConfig)
	}
	log.Println("We get the transitConfig: ", transitConfig, missPos)

	//// in the Get case:
	if !ok {
		log.Println("out of the for, ", eventResult{
			errCode: ERR_CODE_PATH_TRANSIT_FAIL,
			result:  nil,
		}) // event.result <-
		return
	} else {
		result, ok := transitConfig[paths[len(paths)-1]]
		log.Println("Get operation: ", result, ok)
	}

	// in the Put case:
	if !ok {
		// Create transit path, and move transitConfig forward.
		// If nil is in the config, we will remove it or override it automatically here.
		log.Println(transitConfig, transitPaths, len(transitPaths), missPos, transitPaths[missPos])
		for ; missPos < len(transitPaths); missPos++ {
			var newMap = make(map[string]interface{})
			transitConfig[transitPaths[missPos]] = newMap
			transitConfig = newMap
		}
		log.Println("Transit created, we get the transitConfig: ", transitConfig)
	}

	// put nil means to delete. It is not allowed that the config contain nil.
	if newData == nil {
		// todo: iterate from bottom to top, to delete empty map
		delete(transitConfig, paths[len(paths)-1])
	} else {
		transitConfig[paths[len(paths)-1]] = newData
	}
	log.Println("After config: ", config)

	return

	// =========================================================*/















	//paths := strings.Split("/qingdao/ra1dar/118.5/beij1ing-2", "/")
	//if paths[0] != "" || paths[len(paths)-1] == "" {
	//	util.MaoLogM(util.WARN, MODULE_NAME, "format of config path is not correct.")
	//}
	//
	//transitPaths := paths[1:len(paths)-1]
	//transitConfig := config
	//var ok bool
	//for _, p := range transitPaths {
	//	tmpvar := transitConfig[p]
	//	transitConfig, ok = tmpvar.(map[string]interface{}) // can adapt nil
	//	if !ok {
	//		// todo:fail
	//	}
	//	log.Println(transitConfig)
	//}
	//result := transitConfig[paths[len(paths)-1]]
	//log.Println(result)
	//return

	//if InitConfigModule("mao-config.yaml") == false {
	//	return
	//}
	//
	//
	//log.Println(len(eventChannel))
	//go GetConfig("getPath")
	////eventP := &configEvent{
	////	eventType: EVENT_PUT,
	////	path:      "putPath",
	////	data:      make([]string, 3),
	////	result:    make(chan interface{}, 1),
	////}
	////eventChannel <- eventP
	//
	//log.Println(len(eventChannel))
	//mapmap := make(map[string]interface{})
	//mapmap["beijing-1"] = 118.5
	//mapmap["beijing-2"] = true
	//mapmap["beijing-3"] = "radar contact"
	//go PutConfig("putPath", mapmap)
	////eventG := &configEvent{
	////	eventType: EVENT_GET,
	////	path:      "getPath",
	////	data:      nil,
	////	result:    make(chan interface{}, 1),
	////}
	////eventChannel <- eventG
	//
	//log.Println(len(eventChannel))
	//
	//log.Println(eventChannel)
	//time.Sleep(3 * time.Second)
	//RequireShutdown()
	//time.Sleep(1000 * time.Second)
}

/**
content, _ := ioutil.ReadFile(DEFAULT_CONFIG_FILE)

var c map[string]interface{}
yaml.Unmarshal(content, &c)
fmt.Printf("%v", c)

orderer := c["orderer"].(map[string]interface{})
orderer2 := c["orderer"].(map[string]interface{})

delete(orderer, )

mmm := make(map[string]interface{})
mmm["beijing"] = "lei da yin dao"
mmm["xichang"] = 6858
(c["orderer"].(map[interface{}]interface{}))["baas"] = mmm

fmt.Printf("%v", c)
data, _ := yaml.Marshal(c)

fmt.Printf("%v", data)
ioutil.WriteFile("yamlAfter.yaml", data, 0666)

fmt.Printf("%v", data)
*/
