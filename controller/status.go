package controller

import (
	"github.com/cntechpower/v2ray-webui/handler"
	"github.com/gin-gonic/gin"
)

func AddStatusHandler(engine *gin.RouterGroup) (teardownFunc func()) {
	group := engine.Group("/status")
	controller := newStatusController()
	group.GET("/ping", controller.Ping)
	group.GET("/v2ray", controller.V2ray)
	return func() {}
}

type statusController struct {
	*baseController
}

func newStatusController() *statusController {
	return &statusController{}
}

func (c *statusController) Ping(ctx *gin.Context) {
	c.DoJSONFunc(ctx, func() (interface{}, error) {
		return "pong", nil
	})
}

func (c *statusController) V2ray(ctx *gin.Context) {
	c.DoJSONFunc(ctx, func() (interface{}, error) {
		return handler.V2ray.Status()
	})
}
