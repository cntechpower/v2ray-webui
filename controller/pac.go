package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/model/params"
)

func AddProxyHandler(engine *gin.Engine) (teardownFunc func()) {
	//proxy handler
	{
		pacGroup := engine.Group("/pac")
		controller, err := NewPacController()
		if err != nil {
			panic(err)
		}
		pacGroup.GET("", controller.GetCurrentPAC)
		pacGroup.POST("/cron", controller.UpdateCron)
		pacGroup.GET("/cron", controller.GetCurrentCron)
		pacGroup.POST("/generate", controller.ManualGeneratePac)

		webSiteGroup := pacGroup.Group("/website")
		{
			webSiteGroup.GET("/list", controller.ListCustomProxyWebsites)
			webSiteGroup.POST("/add", controller.AddCustomPacWebsites)
			webSiteGroup.POST("/del", controller.DelCustomPacWebsites)
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

func (h *PacController) ListCustomProxyWebsites(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		res, err := h.service.ListCustomProxyWebsites()
		if err != nil {
			return nil, err
		}
		return model.RenderPacWebSites(res), nil
	})
}

func (h *PacController) AddCustomPacWebsites(c *gin.Context) {
	param := &params.AddCustomPacWebsitesParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.AddCustomPacWebsites(param.WebSite)
	})
}

func (h *PacController) DelCustomPacWebsites(c *gin.Context) {
	param := &params.DelCustomPacWebsitesParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.DelCustomPacWebsites(param.WebSiteId)
	})
}
