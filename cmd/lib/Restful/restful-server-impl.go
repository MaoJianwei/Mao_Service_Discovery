package Restful

import (
	"MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
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
	util.MaoLog(util.INFO, fmt.Sprintf("Starting web show %s ...", r.serviceAddr))
	err := r.restful.Run(r.serviceAddr)
	if err != nil {
		util.MaoLog(util.ERROR, fmt.Sprintf("Fail to run rest server, %s", err))
	}
}

func (r *RestfulServerImpl) StartRestfulServerDaemon(webAddr string) {
	r.serviceAddr = webAddr
	go r.startRestfulServer()
}