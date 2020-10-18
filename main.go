package main

import (
	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/log"
	"cntechpower.com/api-server/persist"
	"github.com/gin-gonic/gin"
)

var mysqlDSN = "api:api@tcp(127.0.0.1:3306)/api?charset=utf8mb4&parseTime=True&loc=Local"
var redisDSN = "127.0.0.1:6379"

func main() {
	log.InitLogger("")
	if err := persist.Init(mysqlDSN, redisDSN); err != nil {
		panic(err)
	}
	engine := gin.New()
	//proxy handler
	{
		proxyGroup := engine.Group("/proxy")
		webSiteGroup := proxyGroup.Group("/website")
		webSiteGroup.GET("/list", handler.ListCustomProxyWebsites)
		webSiteGroup.GET("/listv2", handler.ListCustomProxyWebsitesWithoutCache)
		webSiteGroup.POST("/add", handler.AddCustomProxyWebsites)
		webSiteGroup.POST("/del", handler.DelCustomProxyWebsites)
	}
	engine.Use()
	if err := engine.Run("0.0.0.0:8888"); err != nil {
		panic(err)
	}
}
