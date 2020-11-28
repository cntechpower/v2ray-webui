package controller

import (
	"net/http"

	"github.com/cntechpower/v2ray-webui/model"
	"github.com/gin-gonic/gin"
)

type baseController struct {
}

func (h *baseController) DoFunc(c *gin.Context, f func() error) {
	if err := f(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewGenericStatus(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, "Operation Succeed."))
}

func (h *baseController) DoJSONFunc(c *gin.Context, f func() (interface{}, error)) {
	res, err := f()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewGenericStatus(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *baseController) DoStringFunc(c *gin.Context, f func() (string, error)) {
	res, err := f()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewGenericStatus(http.StatusInternalServerError, err.Error()))
		return
	}
	c.String(http.StatusOK, res)
}

func (h *baseController) GenericWrapper(f func() error) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := f(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewGenericStatus(http.StatusInternalServerError, err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, "Operation Succeed."))
	}
}
