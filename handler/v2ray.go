package handler

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"

	"cntechpower.com/api-server/log"

	"github.com/gin-gonic/gin"

	v2ray "v2ray.com/core"
	_ "v2ray.com/core/app/proxyman/inbound"
	_ "v2ray.com/core/app/proxyman/outbound"
	v2rayConf "v2ray.com/core/infra/conf/serial"
)

func AddV2rayHandler(engine *gin.Engine) (tearDown func()) {
	h, err := NewV2rayHandler()
	//v2ray handler
	{
		v2rayGroup := engine.Group("/v2ray")
		if err != nil {
			panic(err)
		}
		v2rayConfigGroup := v2rayGroup.Group("/config")
		v2rayConfigGroup.POST("/reload", h.GenericWrapper(h.ReloadConfig))
		v2rayConfigGroup.GET("/get", h.GetConfig)
		v2rayServerGroup := v2rayGroup.Group("/server")
		v2rayServerGroup.POST("/start", h.GenericWrapper(h.StartV2ray))
		v2rayServerGroup.POST("/stop", h.GenericWrapper(h.StopV2ray))
	}
	return h.TearDown

}

type V2rayHandler struct {
	*baseHandler
	v2rayConfig   *v2ray.Config
	v2rayConfigMu sync.Mutex
	v2rayServer   v2ray.Server
	v2rayServerMu sync.Mutex
}

func NewV2rayHandler() (*V2rayHandler, error) {
	return &V2rayHandler{
		baseHandler:   &baseHandler{},
		v2rayConfig:   nil,
		v2rayConfigMu: sync.Mutex{},
		v2rayServer:   nil,
		v2rayServerMu: sync.Mutex{},
	}, nil
}

func (h *V2rayHandler) TearDown() {
	h.v2rayServerMu.Lock()
	defer h.v2rayServerMu.Unlock()
	if h.v2rayServer != nil {
		header := log.NewHeader("V2rayHandler.TearDown")
		log.Infof(header, "stopping v2ray...")
		_ = h.v2rayServer.Close()
		h.v2rayServer = nil
		log.Infof(header, "stopped v2ray...")
	}
}

func (h *V2rayHandler) ReloadConfig() error {
	f, err := os.Open("./conf/v2ray.json")
	if err != nil {
		return err
	}
	defer f.Close()

	tmpConfig, err := v2rayConf.LoadJSONConfig(f)
	if err != nil {
		return err
	}
	h.v2rayConfigMu.Lock()
	h.v2rayConfig = tmpConfig
	h.v2rayConfigMu.Unlock()
	return nil
}

func (h *V2rayHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.v2rayConfig)
}

func (h *V2rayHandler) StartV2ray() error {
	header := log.NewHeader("StartV2ray")
	h.v2rayServerMu.Lock()
	defer h.v2rayServerMu.Unlock()
	if h.v2rayServer != nil {
		return fmt.Errorf("v2ray server already started")
	}

	//generate a new server
	h.v2rayConfigMu.Lock()
	if h.v2rayConfig == nil {
		h.v2rayConfigMu.Unlock()
		return fmt.Errorf("v2ray config not init")
	}
	server, err := v2ray.New(h.v2rayConfig)
	h.v2rayConfigMu.Unlock()
	if err != nil {
		return err
	}
	if err := server.Start(); err != nil {
		log.Errorf(header, "Failed to start", err)
		return err
	}
	h.v2rayServer = server
	log.Infof(header, "started")

	// FROM v2ray: Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()
	return nil
}

func (h *V2rayHandler) StopV2ray() error {
	h.v2rayServerMu.Lock()
	defer h.v2rayServerMu.Unlock()
	if h.v2rayServer == nil {
		return fmt.Errorf("v2ray server is not started")
	}
	if err := h.v2rayServer.Close(); err != nil {
		header := log.NewHeader("StopV2ray")
		errMsg := fmt.Errorf("v2ray server stop failed: %v", err)
		log.Errorf(header, "%v", errMsg)
		return errMsg
	}
	h.v2rayServer = nil
	return nil
}
