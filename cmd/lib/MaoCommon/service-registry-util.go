package MaoCommon

import MaoApi "MaoServerDiscovery/cmd/api"

// If you add a new api, please provide a util function here for it :)

// if fail, return nil
func ServiceRegistryGetConfigModule() (serviceInstance MaoApi.ConfigModule) {
	configModule, _ := GetService(MaoApi.ConfigModuleRegisterName).(MaoApi.ConfigModule)
	return configModule
}

// if fail, return nil
func ServiceRegistryGetGrpcKaModule() (serviceInstance MaoApi.GrpcKaModule) {
	grpcKaModule, _ := GetService(MaoApi.GrpcKaModuleRegisterName).(MaoApi.GrpcKaModule)
	return grpcKaModule
}

// if fail, return nil
func ServiceRegistryGetIcmpKaModule() (serviceInstance MaoApi.IcmpKaModule) {
	icmpKaModule, _ := GetService(MaoApi.IcmpKaModuleRegisterName).(MaoApi.IcmpKaModule)
	return icmpKaModule
}

// if fail, return nil
func ServiceRegistryGetRestfulServerModule() (serviceInstance MaoApi.RestfulServerModule) {
	restfulServer, _ := GetService(MaoApi.RestfulServerRegisterName).(MaoApi.RestfulServerModule)
	return restfulServer
}

// if fail, return nil
func ServiceRegistryGetEmailModule() (serviceInstance MaoApi.EmailModule) {
	emailModule, _ := GetService(MaoApi.EmailModuleRegisterName).(MaoApi.EmailModule)
	return emailModule
}

// if fail, return nil
func ServiceRegistryGetWechatModule() (serviceInstance MaoApi.WechatModule) {
	wechatModule, _ := GetService(MaoApi.WechatModuleRegisterName).(MaoApi.WechatModule)
	return wechatModule
}

// if fail, return nil
func ServiceRegistryGetAuxDataModule() (serviceInstance MaoApi.AuxDataModule) {
	auxDataModule, _ := GetService(MaoApi.AuxDataModuleRegisterName).(MaoApi.AuxDataModule)
	return auxDataModule
}