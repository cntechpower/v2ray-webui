package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/model/params"
)

func AddProxyHandler(engine *gin.RouterGroup) (teardownFunc func()) {
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
	service *handler.PacHandler
}

func NewPacController() (*PacController, error) {
	service, err := handler.NewPacHandler()
	if err != nil {
		return nil, err
	}
	return &PacController{
		baseController: &baseController{},
		service:        service,
	}, nil
}

func (h *PacController) GetCurrentPAC(c *gin.Context) {
	c.String(http.StatusOK, h.service.GetCurrentPAC())
}

func (h *PacController) ManualGeneratePac(c *gin.Context) {
	h.DoFunc(c, func() error {
		return h.service.ManualGeneratePac()
	})
}

func (h *PacController) UpdateCron(c *gin.Context) {
	param := &params.UpdatePacCronParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.UpdateCron(param.CronString)
	})
}

func (h *PacController) GetCurrentCron(c *gin.Context) {
	c.String(http.StatusOK, h.service.GetCurrentCron())
}

func (h *PacController) ListCustomWebsites(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		res, err := h.service.ListCustomWebsites()
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
		return h.service.AddCustomWebsite(param.WebSite)
	})
}

func (h *PacController) DelCustomWebsites(c *gin.Context) {
	param := &params.DelCustomPacWebsitesParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.DelCustomWebsites(param.WebSiteId)
	})
}

func (h *PacController) UpdateConfig(c *gin.Context) {
	param := &model.PacHandlerConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.UpdateConfig(param)
	})
}

func (h *PacController) GetConfig(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return h.service.GetConfig()
	})
}
