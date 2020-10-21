package main

import (
	"cntechpower.com/api-server/config"
	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/log"
	"cntechpower.com/api-server/persist"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var version string
var configFilePath string

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

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "api config interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Default().Save("./api.config")
		},
	}
	rootCmd.AddCommand(configCmd)
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "api.config", "config file path")
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run() error {
	log.InitLogger("")
	config.Init(configFilePath)
	if err := persist.Init(config.Config.MysqlDSN, config.Config.RedisDSN); err != nil {
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
	return engine.Run(config.Config.ListenAddr)
}
