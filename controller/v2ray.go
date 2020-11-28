package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cntechpower/v2ray-webui/handler"
	"github.com/cntechpower/v2ray-webui/model/params"
)

func AddV2rayHandler(engine *gin.RouterGroup, templateConfigFilePath string) (tearDown func()) {
	controller, err := NewV2rayController(templateConfigFilePath)
	if err != nil {
		panic(err)
	}

	v2rayGroup := engine.Group("/v2ray")
	v2rayGroup.POST("/start", controller.GenericWrapper(controller.service.StartV2ray))
	v2rayGroup.POST("/stop", controller.GenericWrapper(controller.service.StopV2ray))

	v2rayConfigGroup := v2rayGroup.Group("/config")
	v2rayConfigGroup.POST("/switch_node", controller.SwitchNode)
	v2rayConfigGroup.GET("/get", controller.GetConfig)
	v2rayConfigGroup.POST("/update", controller.UpdateConfig)
	v2rayConfigGroup.POST("/validate", controller.ValidateConfig)

	subscriptionGroup := v2rayGroup.Group("/subscription")
	subscriptionGroup.GET("/nodes/list", controller.GetAllV2rayNodes)
	subscriptionGroup.POST("/nodes/ping", controller.PingAllV2rayNodes)
	subscriptionGroup.POST("/add", controller.AddSubscription)
	subscriptionGroup.POST("/delete", controller.DelSubscription)
	subscriptionGroup.GET("/list", controller.GetAllSubscriptions)
	subscriptionGroup.POST("/edit", controller.EditSubscription)
	subscriptionGroup.POST("/refresh", controller.RefreshV2raySubscription)

	return controller.service.TearDown

}

type V2rayController struct {
	*baseController
	service *handler.V2rayHandler
}

func NewV2rayController(templateConfigFilePath string) (*V2rayController, error) {
	service, err := handler.NewV2rayHandler(templateConfigFilePath)
	if err != nil {
		return nil, err
	}
	return &V2rayController{
		baseController: &baseController{},
		service:        service,
	}, nil
}

func (h *V2rayController) SwitchNode(c *gin.Context) {
	param := &params.V2raySwitchNodeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.SwitchNode(param.NodeId)
	})
}

func (h *V2rayController) GetConfig(c *gin.Context) {
	c.String(http.StatusOK, h.service.GetV2rayConfigTemplateContent())
}

func (h *V2rayController) UpdateConfig(c *gin.Context) {
	param := &params.V2rayConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.UpdateConfig(param.ConfigContent)
	})
}

func (h *V2rayController) ValidateConfig(c *gin.Context) {
	param := &params.V2rayConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.ValidateConfig(param.ConfigContent)
	})
}

func (h *V2rayController) GetAllV2rayNodes(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return h.service.GetAllV2rayNodes()
	})
}

func (h *V2rayController) PingAllV2rayNodes(c *gin.Context) {
	h.DoFunc(c, func() error {
		return h.service.PingAllV2rayNodes()
	})
}

func (h *V2rayController) AddSubscription(c *gin.Context) {
	param := &params.AddV2raySubscriptionParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.AddSubscription(param.SubscriptionName, param.SubscriptionAddr)
	})
}

func (h *V2rayController) DelSubscription(c *gin.Context) {
	param := &params.V2raySubscriptionIdParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.DelSubscription(param.SubscriptionId)
	})
}

func (h *V2rayController) GetAllSubscriptions(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return h.service.GetAllSubscriptions()
	})
}

func (h *V2rayController) EditSubscription(c *gin.Context) {
	param := &params.UpdateV2raySubscriptionParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.EditSubscription(param.SubscriptionId, param.SubscriptionName, param.SubscriptionAddr)
	})
}

func (h *V2rayController) RefreshV2raySubscription(c *gin.Context) {
	param := &params.V2raySubscriptionIdParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.service.RefreshV2raySubscription(param.SubscriptionId)
	})
}
