package MaoApi

var (
	ConfigModuleRegisterName = "api-config-module"
)

type ConfigModule interface {
	GetConfig(path string) (object interface{}, errCode int)
	PutConfig(path string, data interface{}) (success bool, errCode int)
}

