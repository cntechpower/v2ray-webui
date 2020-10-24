package main

import (
	"cntechpower.com/api-server/config"
	"cntechpower.com/api-server/handler"
	"cntechpower.com/api-server/log"
	"cntechpower.com/api-server/persist"
	"cntechpower.com/api-server/util"
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
		Run: run,
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "api config interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Default().Save(configFilePath)
		},
	}
	rootCmd.AddCommand(configCmd)
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "./conf/api.config", "config file path")
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}

func run(_ *cobra.Command, _ []string) {
	log.InitLogger("")
	h := log.NewHeader("api-server")
	config.Init(configFilePath)
	if err := persist.Init(config.Config.MysqlDSN, config.Config.RedisDSN); err != nil {
		panic(err)
	}
	engine := gin.New()
	if !config.Config.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}
	tearDownFuncs := make([]func(), 0)
	tearDownFuncs = append(tearDownFuncs,
		handler.AddProxyHandler(engine),
		handler.AddV2rayHandler(engine))
	httpExistChan := make(chan error)
	go func() {
		httpExistChan <- engine.Run(config.Config.ListenAddr)
	}()

	//wait for os kill signal. TODO: graceful shutdown
	go util.ListenTTINSignalLoop()
	serverExitChan := util.ListenKillSignal()
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
