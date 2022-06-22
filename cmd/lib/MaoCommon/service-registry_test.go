package MaoCommon

import (
	MaoApi "MaoServerDiscovery/cmd/api"
	"MaoServerDiscovery/cmd/lib/Restful"
	"testing"
)

func TestServiceRegistry_GetService(t *testing.T) {
	restfulServer := &Restful.RestfulServerImpl{}
	RegisterService(MaoApi.RestfulServerRegisterName, restfulServer)

	r, ok := GetService(MaoApi.RestfulServerRegisterName).(MaoApi.RestfulServerModule)
	if r != nil && ok {
	} else {
		t.Errorf("Fail case: get RestfulServerRegisterName and cast to RestfulServerModule, %v, %v", r, ok)
	}

	c, ok := GetService(MaoApi.RestfulServerRegisterName).(MaoApi.ConfigModule)
	if c == nil && !ok {
	} else {
		t.Errorf("Fail case: get RestfulServerRegisterName and cast to ConfigModule, %v, %v", c, ok)
	}

	ccc, ok := GetService(MaoApi.ConfigModuleRegisterName).(MaoApi.ConfigModule)
	if ccc == nil && !ok {
	} else {
		t.Errorf("Fail case: get ConfigModuleRegisterName and cast to ConfigModule, %v, %v", ccc, ok)
	}
}