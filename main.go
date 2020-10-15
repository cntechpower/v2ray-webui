package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Printf("hello world")
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, `{"message": "success"}`)
	})
	if err := engine.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}
