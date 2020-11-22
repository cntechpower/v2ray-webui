package controller

import (
	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/model/params"
	"github.com/gin-gonic/gin"
)

func AddFileHandler(engine *gin.RouterGroup) (teardownFunc func()) {
	fileGroup := engine.Group("/file")
	controller := NewFileController()
	fileGroup.GET("/read", controller.ReadFile)
	return func() {}
}

type FileController struct {
	*baseController
	service *handler.FileHandler
}

func NewFileController() *FileController {
	return &FileController{
		service: handler.NewFileHandler(),
	}
}

func (c *FileController) ReadFile(ctx *gin.Context) {
	param := &params.ReadFileParam{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	c.DoStringFunc(ctx, func() (string, error) {
		return c.service.ReadFile(param.FileName, int(param.From), int(param.To))
	})
}
