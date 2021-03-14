package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cntechpower/v2ray-webui/handler"
	"github.com/cntechpower/v2ray-webui/model"
	"github.com/cntechpower/v2ray-webui/model/params"
)

func AddPacHandler(engine *gin.RouterGroup) (teardownFunc func()) {
	//proxy handler
	{
		pacGroup := engine.Group("/pac")
		controller, err := NewPacController()
		if err != nil {
			panic(err)
		}
		pacGroup.GET("/get", controller.GetCurrentPAC)
		pacGroup.POST("/update", controller.ManualGeneratePac)
		cronGroup := pacGroup.Group("/cron")
		{
			cronGroup.POST("/update", controller.UpdateCron)
			cronGroup.GET("/get", controller.GetCurrentCron)

		}
		configGroup := pacGroup.Group("/config")
		{
			configGroup.GET("/get", controller.GetConfig)
			configGroup.POST("/update", controller.UpdateConfig)
		}

		webSiteGroup := pacGroup.Group("/website")
		{
			webSiteGroup.GET("/list", controller.ListCustomWebsites)
			webSiteGroup.POST("/add", controller.AddCustomWebsites)
			webSiteGroup.POST("/del", controller.DelCustomWebsites)
		}

	}
	//PAC service has no tear down func.
	return func() {}
}

type PacController struct {
	*baseController
}

func NewPacController() (*PacController, error) {
	return &PacController{
		baseController: &baseController{},
	}, nil
}

func (h *PacController) GetCurrentPAC(c *gin.Context) {
	c.String(http.StatusOK, handler.Pac.GetCurrentPAC())
}

func (h *PacController) ManualGeneratePac(c *gin.Context) {
	h.DoFunc(c, func() error {
		return handler.Pac.ManualGeneratePac()
	})
}

func (h *PacController) UpdateCron(c *gin.Context) {
	param := &params.UpdatePacCronParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.Pac.UpdateCron(param.CronString)
	})
}

func (h *PacController) GetCurrentCron(c *gin.Context) {
	c.String(http.StatusOK, handler.Pac.GetCurrentCron())
}

func (h *PacController) ListCustomWebsites(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		res, err := handler.Pac.ListCustomWebsites()
		if err != nil {
			return nil, err
		}
		return model.RenderPacWebSites(res), nil
	})
}

func (h *PacController) AddCustomWebsites(c *gin.Context) {
	param := &params.AddCustomPacWebsitesParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.Pac.AddCustomWebsite(param.WebSite)
	})
}

func (h *PacController) DelCustomWebsites(c *gin.Context) {
	param := &params.DelCustomPacWebsitesParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.Pac.DelCustomWebsites(param.WebSiteId)
	})
}

func (h *PacController) UpdateConfig(c *gin.Context) {
	param := &model.PacHandlerConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.Pac.UpdateConfig(param)
	})
}

func (h *PacController) GetConfig(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return handler.Pac.GetConfig()
	})
}
