package controller

import (
	"github.com/cntechpower/v2ray-webui/handler"

	"github.com/cntechpower/v2ray-webui/model"

	"github.com/gin-gonic/gin"

	"github.com/cntechpower/v2ray-webui/model/params"
)

func AddV2rayHandler(engine *gin.RouterGroup, templateConfigFilePath string) (tearDown func()) {
	controller, err := NewV2rayController(templateConfigFilePath)
	if err != nil {
		panic(err)
	}

	v2rayGroup := engine.Group("/v2ray")
	v2rayGroup.POST("/start", controller.GenericWrapper(handler.V2ray.StartV2ray))
	v2rayGroup.POST("/stop", controller.GenericWrapper(handler.V2ray.StopV2ray))

	v2rayConfigGroup := v2rayGroup.Group("/config")
	v2rayConfigGroup.POST("/switch_node", controller.SwitchNode)
	v2rayConfigGroup.GET("/get", controller.GetConfig)
	v2rayConfigGroup.GET("/get_trojan", controller.GetTrojanConfig)
	v2rayConfigGroup.POST("/update", controller.UpdateConfig)
	v2rayConfigGroup.POST("/validate", controller.ValidateConfig)

	subscriptionGroup := v2rayGroup.Group("/subscription")
	subscriptionGroup.POST("/add", controller.AddSubscription)
	subscriptionGroup.POST("/delete", controller.DelSubscription)
	subscriptionGroup.GET("/list", controller.GetAllSubscriptions)
	subscriptionGroup.POST("/edit", controller.EditSubscription)
	subscriptionGroup.POST("/refresh", controller.RefreshV2raySubscription)

	nodesGroup := v2rayGroup.Group("/nodes")
	nodesGroup.GET("/list", controller.GetAllV2rayNodes)
	nodesGroup.POST("/ping", controller.PingAllV2rayNodes)
	nodesGroup.POST("/add", controller.AddV2rayNode)

	return handler.V2ray.TearDown

}

type V2rayController struct {
	*baseController
}

func NewV2rayController(templateConfigFilePath string) (*V2rayController, error) {
	return &V2rayController{
		baseController: &baseController{},
	}, nil
}

func (h *V2rayController) SwitchNode(c *gin.Context) {
	param := &params.V2raySwitchNodeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.SwitchNode(param.NodeId)
	})
}

func (h *V2rayController) GetConfig(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return handler.V2ray.GetV2rayConfigTemplateContent(), nil
	})
}

func (h *V2rayController) GetTrojanConfig(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return handler.V2ray.GetV2rayTrojanConfigTemplateContent(), nil
	})
}

func (h *V2rayController) UpdateConfig(c *gin.Context) {
	param := &params.V2rayConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.UpdateConfig(param.ConfigContent, param.Type)
	})
}

func (h *V2rayController) ValidateConfig(c *gin.Context) {
	param := &params.V2rayConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.ValidateConfig(param.ConfigContent)
	})
}

func (h *V2rayController) GetAllV2rayNodes(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return handler.V2ray.GetAllV2rayNodes()
	})
}

func (h *V2rayController) AddV2rayNode(c *gin.Context) {
	param := &params.V2rayAddNodeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.AddNode(&model.V2rayNode{
			SubscriptionId:   0,
			SubscriptionName: "none",
			Host:             param.Host,
			Path:             param.Path,
			TLS:              param.TLS,
			Address:          param.Address,
			Port:             model.FlexString(param.Port),
			Aid:              model.FlexString(param.Aid),
			Net:              param.Net,
			Type:             param.Type,
			V:                param.V,
			Name:             param.Name,
			ServerId:         param.ServerId,
		})
	})
}

func (h *V2rayController) PingAllV2rayNodes(c *gin.Context) {
	h.DoFunc(c, func() error {
		return handler.V2ray.PingAllV2rayNodes()
	})
}

func (h *V2rayController) AddSubscription(c *gin.Context) {
	param := &params.AddV2raySubscriptionParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.AddSubscription(param.SubscriptionName, param.SubscriptionAddr)
	})
}

func (h *V2rayController) DelSubscription(c *gin.Context) {
	param := &params.V2raySubscriptionIdParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.DelSubscription(param.SubscriptionId)
	})
}

func (h *V2rayController) GetAllSubscriptions(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return handler.V2ray.GetAllSubscriptions()
	})
}

func (h *V2rayController) EditSubscription(c *gin.Context) {
	param := &params.UpdateV2raySubscriptionParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.EditSubscription(param.SubscriptionId, param.SubscriptionName, param.SubscriptionAddr)
	})
}

func (h *V2rayController) RefreshV2raySubscription(c *gin.Context) {
	param := &params.V2raySubscriptionIdParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return handler.V2ray.RefreshV2raySubscription(param.SubscriptionId)
	})
}
