package handler

import (
	"fmt"
	"net/http"

	"cntechpower.com/api-server/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var checker *validator.Validate

func init() {
	checker = validator.New()
}

type baseHandler struct {
}

func (h *baseHandler) GenericWrapper(f func() error) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := f(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewGenericStatus(http.StatusInternalServerError, err.Error()))
			return
		}
		c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, "Operation Succeed."))
	}

}
func errorWith500(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewGenericStatus(http.StatusInternalServerError, err.Error()))
}
func ok(c *gin.Context, message string, a ...interface{}) {
	c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, fmt.Sprintf(message, a...)))
}
