package v2ray

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"sync"
	"time"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/handler/base"
	"github.com/cntechpower/v2ray-webui/model"

	"go.uber.org/atomic"
	v2ray "v2ray.com/core"
	_ "v2ray.com/core/app/proxyman/inbound"
	_ "v2ray.com/core/app/proxyman/outbound"
)

type Handler struct {
	*base.Handler

	//v2ray config template
	//support {placeholder}, any api update to will persist to v2rayConfigTemplateFilePath.
	v2rayConfigTemplateFilePath string
	v2rayConfigTemplateContent  string

	v2rayTrojanConfigTemplateFilePath string
	v2rayTrojanConfigTemplateContent  string

	//v2rayCurrentConfig is template after replace {placeholder}
	v2rayCurrentConfig string
	v2rayConfig        *v2ray.Config
	v2rayCurrentNode   *model.V2rayNode
	v2rayConfigMu      sync.Mutex

	//v2rayServer is origin v2ray server struct.
	//used for control v2ray start/stop
	v2rayServer          v2ray.Server
	v2rayServerMu        sync.Mutex
	v2rayServerStartTime time.Time

	//use this to avoid concurrency refreshing Subscription
	v2raySubscriptionRefreshing atomic.Bool

	//status
	v2rayStatusRefreshTime time.Time
}

func New(templateConfigFilePath, v2rayTrojanConfigTemplatePath string) (h *Handler, err error) {
	var bs, bs1 []byte
	bs, err = ioutil.ReadFile(templateConfigFilePath)
	if err != nil {
		return nil, err
	}

	bs1, err = ioutil.ReadFile(v2rayTrojanConfigTemplatePath)
	if err != nil {
		return nil, err
	}
	h = &Handler{
		Handler:                           &base.Handler{},
		v2rayConfigTemplateFilePath:       templateConfigFilePath,
		v2rayConfigTemplateContent:        string(bs),
		v2rayTrojanConfigTemplateFilePath: v2rayTrojanConfigTemplatePath,
		v2rayTrojanConfigTemplateContent:  string(bs1),
		v2rayConfig:                       nil,
		v2rayConfigMu:                     sync.Mutex{},
		v2rayServer:                       nil,
		v2rayServerMu:                     sync.Mutex{},
		v2raySubscriptionRefreshing:       atomic.Bool{},
	}
	go h.refreshStatusLoop()
	return

}

func (h *Handler) TearDown() {
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

func (h *Handler) StartV2ray() error {
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
		log.Errorf(header, "Failed to start: %v", err)
		return err
	}
	h.v2rayServer = server
	h.v2rayServerStartTime = time.Now()
	log.Infof(header, "started")

	// FROM v2ray: Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()
	return nil
}

var ErrV2rayNotStarted = fmt.Errorf("v2ray server is not started")

func (h *Handler) StopV2ray() error {
	h.v2rayServerMu.Lock()
	defer h.v2rayServerMu.Unlock()
	if h.v2rayServer == nil {
		return ErrV2rayNotStarted
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
