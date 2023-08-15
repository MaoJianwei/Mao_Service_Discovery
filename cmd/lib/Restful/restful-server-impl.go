package Restful

import (
	"MaoServerDiscovery/util"
	"fmt"
	"github.com/gin-gonic/gin"
)

const (
	MODULE_NAME = "Restful-Server-module"
)

type RestfulServerImpl struct {
	restful *gin.Engine

	serviceAddr string

	uiPageLinks []string
	getApiLinks []string
	postApiLinks []string
}

func (r *RestfulServerImpl) InitRestfulServer() {
	gin.SetMode(gin.ReleaseMode)
	r.restful = gin.Default()

	r.restful.LoadHTMLGlob("resource/html/*")
	r.restful.Static("/js", "resource/static/js")
	r.restful.Static("/css", "resource/static/css")
	r.restful.Static("/static", "resource/static")
	r.restful.StaticFile("/favicon.ico", "resource/static/favicon.ico")


	r.restful.GET("/", r.showHomePage)
	r.restful.GET("/api", r.showApiListPage)

	// not need to initiate []string
}

func (r *RestfulServerImpl) showHomePage(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func (r *RestfulServerImpl) RegisterUiPage(relativePath string, handlers ...gin.HandlerFunc) {
	r.restful.GET("/v1" + relativePath, handlers...)
	r.uiPageLinks = append(r.uiPageLinks, "/v1" + relativePath)
}

func (r *RestfulServerImpl) RegisterGetApi(relativePath string, handlers ...gin.HandlerFunc) {
	r.restful.GET("/api" + relativePath, handlers...)
	r.getApiLinks = append(r.getApiLinks, "/api" + relativePath)
}

func (r *RestfulServerImpl) RegisterPostApi(relativePath string, handlers ...gin.HandlerFunc) {
	r.restful.POST("/api" + relativePath, handlers...)
	r.postApiLinks = append(r.postApiLinks, "/api" + relativePath)
}

func (r *RestfulServerImpl) showApiListPage(c *gin.Context) {
	htmlHead := `<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8"><title>MaoServiceDiscovery: URLs</title></head><body>`

	ret := "UI:<br/>"
	for _, v := range r.uiPageLinks {
		ret = fmt.Sprintf(`%s<a href="%s">%s</a><br/>`, ret, v, v)
	}

	ret += "<br/>GET API:<br/>"
	for _, v := range r.getApiLinks {
		ret = fmt.Sprintf(`%s<a href="%s">%s</a><br/>`, ret, v, v)
	}

	ret += "<br/>POST API:<br/>"
	for _, v := range r.postApiLinks {
		ret = fmt.Sprintf("%s%s<br/>", ret, v)
	}

	htmlTail := `</body></html>`

	ret = fmt.Sprintf("%s%s%s", htmlHead, ret, htmlTail)

	//c.String(200, ret)
	c.Data(200, "text/html; charset=utf-8", []byte(ret))
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