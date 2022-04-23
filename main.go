package main

import (
	"github.com/cntechpower/v2ray-webui/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/utils/os"
	"github.com/cntechpower/v2ray-webui/config"
	"github.com/cntechpower/v2ray-webui/controller"
	"github.com/cntechpower/v2ray-webui/persist"
)

var version string
var v2rayConfigTemplatePath, v2rayTrojanConfigTemplatePath string

func main() {

	var rootCmd = &cobra.Command{
		Use:   "v2ray-webui",
		Short: "v2ray-webui is v2ray client management tool",
		Long: `Manage proxy and many other resources
Written by dujinyang.
Version: ` + version,
		Run: run,
	}

	var configCmd = &cobra.Command{
		Use:   "reset",
		Short: "reset v2ray-webui config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Default().Save()
		},
	}
	rootCmd.AddCommand(configCmd)
	rootCmd.PersistentFlags().StringVarP(&v2rayConfigTemplatePath, "vtemplate", "v", "./conf/v2ray.json", "v2ray config template file path")
	rootCmd.PersistentFlags().StringVarP(&v2rayTrojanConfigTemplatePath, "vt_template", "t", "./conf/v2ray_trojan.json", "v2ray trojan config template file path")
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) {
	log.InitLogger("")
	h := log.NewHeader("v2ray-webui")
	config.Init()
	if err := persist.Init(); err != nil {
		panic(err)
	}
	if err := handler.Init(v2rayConfigTemplatePath, v2rayTrojanConfigTemplatePath); err != nil {
		panic(err)
	}
	engine := gin.New()
	engine.Use(gin.ErrorLogger())
	if !config.Config.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		//debug mode, set no cors check.
		log.Infof(h, "running in debug mode, turn off cors check.")
		engine.Use(cors.New(cors.Config{
			AllowAllOrigins:        true,
			AllowWildcard:          true,
			AllowBrowserExtensions: true,
			AllowWebSockets:        true,
			AllowFiles:             true,
		}))
	}
	staticHandler := static.Serve("/", static.LocalFile("./static/front-end", true))
	engine.Use(staticHandler)
	engine.NoRoute(func(c *gin.Context) {
		c.File("./static/front-end/index.html")
	})
	apiGroup := engine.Group("/api")
	tearDownFuncs := make([]func(), 0)
	tearDownFuncs = append(tearDownFuncs,
		controller.AddPacHandler(apiGroup),
		controller.AddV2rayHandler(apiGroup, v2rayConfigTemplatePath),
		controller.AddFileHandler(apiGroup),
		controller.AddStatusHandler(apiGroup),
	)
	httpExistChan := make(chan error)
	go func() {
		httpExistChan <- engine.Run(config.Config.ListenAddr)
	}()

	//wait for os kill signal. TODO: graceful shutdown
	go os.ListenTTINSignalLoop()
	serverExitChan := os.ListenKillSignal()
	select {
	case <-serverExitChan:
		log.Infof(h, "Server Existing")
	case err := <-httpExistChan:
		log.Fatalf(h, "api server exit with error: %v", err)
	}
	for _, f := range tearDownFuncs {
		f()
	}
}
