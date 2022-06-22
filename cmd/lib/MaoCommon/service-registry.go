package MaoCommon

var (
	serviceRegistry = make(map[string]interface{})
)

func RegisterService(apiName string, serviceInstancePointer interface{}) {
	serviceRegistry[apiName] = serviceInstancePointer
}

// now: return the instance, not the pointer of the instance
func GetService(apiName string) (serviceInstance interface{}) {
	// todo, test me
	return serviceRegistry[apiName]
}
