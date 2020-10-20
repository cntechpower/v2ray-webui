package main

import (
	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/log"
	"cntechpower.com/api-server/persist"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var mysqlDSN string
var redisDSN string
var version string
var bindAddr string

func main() {

	var rootCmd = &cobra.Command{
		Use:   "api-server",
		Short: "api-server is private cloud management tool",
		Long: `Manage proxy and many other resources
Written by dujinyang.
Version: ` + version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	rootCmd.PersistentFlags().StringVarP(&mysqlDSN, "mysql", "m",
		"api:api@tcp(127.0.0.1:3306)/api?charset=utf8mb4&parseTime=True&loc=Local", "mysql dsn")
	rootCmd.PersistentFlags().StringVarP(&redisDSN, "redis", "r",
		"127.0.0.1:6379", "redis dsn")
	rootCmd.PersistentFlags().StringVarP(&bindAddr, "bind", "b",
		"0.0.0.0:8888", "bind address")
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run() error {
	log.InitLogger("")
	if err := persist.Init(mysqlDSN, redisDSN); err != nil {
		panic(err)
	}
	engine := gin.New()
	//proxy handler
	{
		proxyGroup := engine.Group("/proxy")
		h, err := handler.NewProxyHandler()
		if err != nil {
			return err
		}
		{
			webSiteGroup := proxyGroup.Group("/website")
			webSiteGroup.GET("/list", h.ListCustomProxyWebsites)
			webSiteGroup.GET("/listv2", h.ListCustomProxyWebsitesWithoutCache)
			webSiteGroup.GET("/listv3", h.ListCustomProxyWebsitesInOneCache)
			webSiteGroup.POST("/add", h.AddCustomProxyWebsites)
			webSiteGroup.POST("/del", h.DelCustomProxyWebsites)
		}
		{
			pacGroup := proxyGroup.Group("/pac")
			pacGroup.GET("", h.GetCurrentPAC)
			pacGroup.POST("/cron", h.UpdateCron)
			pacGroup.GET("/cron", h.GetCurrentCron)
			pacGroup.POST("/generate", h.ManualGeneratePac)
		}

	}
	engine.Use()
	return engine.Run(bindAddr)
}
