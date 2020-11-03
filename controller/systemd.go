package controller

import (
	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/model/params"
	"github.com/gin-gonic/gin"
)

func AddSystemdHandler(engine *gin.Engine, serviceNames []string) (tearDown func()) {
	controller, _ := NewSystemdController(serviceNames)
	systemdGroup := engine.Group("/systemd")
	systemdGroup.GET("/get", controller.GetServiceStatus)
	systemdGroup.GET("/list", controller.ListServiceStatus)

	return func() {}
}

type SystemdController struct {
	*baseController
	service *handler.SystemdHandler
}

func NewSystemdController(serviceNames []string) (*SystemdController, error) {
	service := handler.NewSystemdHandler(serviceNames)
	return &SystemdController{
		baseController: &baseController{},
		service:        service,
	}, nil
}

func (h *SystemdController) GetServiceStatus(c *gin.Context) {
	param := &params.SystemdServiceName{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoJSONFunc(c, func() (interface{}, error) {
		return h.service.CheckService([]string{param.SystemdServiceName})
	})
}

func (h *SystemdController) ListServiceStatus(c *gin.Context) {
	h.DoJSONFunc(c, func() (interface{}, error) {
		return h.service.ListService()
	})
}
