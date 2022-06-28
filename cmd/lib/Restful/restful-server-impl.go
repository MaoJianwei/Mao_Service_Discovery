package Restful

import (
	"MaoServerDiscovery/util"
	"github.com/gin-gonic/gin"
)

const (
	MODULE_NAME = "Restful-Server-module"
)

type RestfulServerImpl struct {
	restful *gin.Engine

	serviceAddr string
}

func (r *RestfulServerImpl) InitRestfulServer() {
	gin.SetMode(gin.ReleaseMode)
	r.restful = gin.Default()

	r.restful.LoadHTMLGlob("resource/*")
	r.restful.Static("/static/", "resource")
}

func (r *RestfulServerImpl) RegisterGetApi(relativePath string, handlers ...gin.HandlerFunc) {
	r.restful.GET(relativePath, handlers...)

}

func (r *RestfulServerImpl) RegisterPostApi(relativePath string, handlers ...gin.HandlerFunc) {
	r.restful.POST(relativePath, handlers...)
}


func (r *RestfulServerImpl) startRestfulServer() {
	util.MaoLogM(util.INFO, MODULE_NAME, "Starting web show %s ...", r.serviceAddr)
	err := r.restful.Run(r.serviceAddr)
	if err != nil {
		util.MaoLogM(util.ERROR, MODULE_NAME, "Fail to run rest server, %s", err)
	}
}

func (r *RestfulServerImpl) StartRestfulServerDaemon(webAddr string) {
	r.serviceAddr = webAddr
	go r.startRestfulServer()
}