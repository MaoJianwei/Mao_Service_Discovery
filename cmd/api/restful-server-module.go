package MaoApi

import "github.com/gin-gonic/gin"

var (
	RestfulServerRegisterName = "api-restful-server-module"
)

type RestfulServerModule interface {
	RegisterUiPage(relativePath string, handlers ...gin.HandlerFunc)
	RegisterGetApi(relativePath string, handlers ...gin.HandlerFunc)
	RegisterPostApi(relativePath string, handlers ...gin.HandlerFunc)
}