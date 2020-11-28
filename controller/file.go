package controller

import (
	"cntechpower.com/api-server/handler"
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
	p := new(struct {
		Type int64 `form:"type" binding:"required"`
		From int64 `form:"from"`
		To   int64 `form:"to" binding:"required"`
	})
	if err := ctx.Bind(p); err != nil {
		return
	}
	c.DoStringFunc(ctx, func() (string, error) {
		return c.service.ReadFile(int(p.Type), int(p.From), int(p.To))
	})
}
