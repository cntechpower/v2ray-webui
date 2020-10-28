package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"

	"go.uber.org/atomic"

	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/model/params"
	"cntechpower.com/api-server/persist"

	"cntechpower.com/api-server/log"

	"github.com/gin-gonic/gin"

	v2ray "v2ray.com/core"
	_ "v2ray.com/core/app/proxyman/inbound"
	_ "v2ray.com/core/app/proxyman/outbound"
	v2rayConf "v2ray.com/core/infra/conf/serial"
)

func AddV2rayHandler(engine *gin.Engine, templateConfigFilePath string) (tearDown func()) {
	h, err := NewV2rayHandler(templateConfigFilePath)
	if err != nil {
		panic(err)
	}

	v2rayGroup := engine.Group("/v2ray")
	v2rayGroup.POST("/start", h.GenericWrapper(h.StartV2ray))
	v2rayGroup.POST("/stop", h.GenericWrapper(h.StopV2ray))

	v2rayConfigGroup := v2rayGroup.Group("/config")
	v2rayConfigGroup.POST("/switch_node", h.SwitchNode)
	v2rayConfigGroup.GET("/get", h.GetConfig)
	v2rayConfigGroup.POST("/update", h.UpdateConfig)

	subscriptionGroup := v2rayGroup.Group("/subscription")
	subscriptionGroup.GET("/nodes/list", h.GetAllV2rayNodes)
	subscriptionGroup.POST("/nodes/ping", h.PingAllV2rayNodes)
	subscriptionGroup.POST("/add", h.AddSubscription)
	subscriptionGroup.POST("/delete", h.DelSubscription)
	subscriptionGroup.GET("/list", h.GetAllSubscriptions)
	subscriptionGroup.POST("/edit", h.EditSubscription)
	subscriptionGroup.POST("/refresh", h.RefreshV2raySubscription)

	return h.TearDown

}

type V2rayHandler struct {
	*baseHandler

	//v2ray config template
	//support {placeholder}, any api update to will persist to v2rayConfigTemplateFilePath.
	v2rayConfigTemplateFilePath string
	v2rayConfigTemplateContent  string

	//v2rayCurrentConfig is template after replace {placeholder}
	v2rayCurrentConfig string
	v2rayConfig        *v2ray.Config
	v2rayCurrentNode   *model.V2rayNode
	v2rayConfigMu      sync.Mutex

	//v2rayServer is origin v2ray server struct.
	//used for control v2ray start/stop
	v2rayServer   v2ray.Server
	v2rayServerMu sync.Mutex

	//use this to avoid concurrency refreshing Subscription
	v2raySubscriptionRefreshing atomic.Bool
}

func NewV2rayHandler(templateConfigFilePath string) (*V2rayHandler, error) {
	bs, err := ioutil.ReadFile(templateConfigFilePath)
	if err != nil {
		return nil, err
	}
	return &V2rayHandler{
		baseHandler:                 &baseHandler{},
		v2rayConfigTemplateFilePath: templateConfigFilePath,
		v2rayConfigTemplateContent:  string(bs),
		v2rayConfig:                 nil,
		v2rayConfigMu:               sync.Mutex{},
		v2rayServer:                 nil,
		v2rayServerMu:               sync.Mutex{},
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

func (h *V2rayHandler) SwitchNode(c *gin.Context) {
	param := &params.V2raySwitchNodeParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return h.switchNode(param.NodeId)
	})
}

func (h *V2rayHandler) validateConfig(config string, node *model.V2rayNode) (*v2ray.Config, error) {
	header := log.NewHeader("V2rayHandler.validateConfig")
	config = strings.ReplaceAll(config, "{serverName}", node.Host)
	config = strings.ReplaceAll(config, "{serverPath}", node.Path)
	config = strings.ReplaceAll(config, "{serverPort}", strconv.FormatInt(node.Port, 10))
	config = strings.ReplaceAll(config, "{serverId}", node.ServerId)
	log.Infof(header, "validate config: %v", config)
	return v2rayConf.LoadJSONConfig(strings.NewReader(config))

}

func (h *V2rayHandler) switchNode(nodeId int64) error {
	node := model.NewV2rayNode(nodeId, "")
	if err := persist.Get(node); err != nil {
		return err
	}
	config, err := h.validateConfig(h.v2rayConfigTemplateContent, node)
	if err != nil {
		return err
	}
	h.v2rayConfigMu.Lock()
	h.v2rayConfig = config
	h.v2rayCurrentNode = node
	h.v2rayConfigMu.Unlock()
	if err := h.StopV2ray(); err != nil {
		return err
	}
	return h.StartV2ray()
}

func (h *V2rayHandler) GetConfig(c *gin.Context) {
	c.String(http.StatusOK, h.v2rayConfigTemplateContent)
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

func (h *V2rayHandler) GetAllV2rayNodes(c *gin.Context) {
	res := make([]*model.V2rayNode, 0)
	if err := persist.MySQL().Find(&res).Error; err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *V2rayHandler) AddSubscription(c *gin.Context) {
	param := &params.AddV2raySubscriptionParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return persist.Save(model.NewSubscription(param.SubscriptionName, param.SubscriptionAddr))
	})
}

func (h *V2rayHandler) DelSubscription(c *gin.Context) {
	param := &params.V2raySubscriptionIdParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		return persist.Delete(&model.Subscription{Id: param.SubscriptionId})
	})
}

func (h *V2rayHandler) PingAllV2rayNodes(c *gin.Context) {
	header := log.NewHeader("PingAllV2rayNodes")
	h.DoFunc(c, func() error {
		nodes := make([]*model.V2rayNode, 0)
		if err := persist.MySQL().Find(&nodes).Error; err != nil {
			return err
		}
		wg := sync.WaitGroup{}
		for _, node := range nodes {
			wg.Add(1)
			n := node
			go func() {
				p, err := ping.NewPinger(n.Host)
				if err != nil {
					n.PingRTT = 999
					if err := persist.Save(n); err != nil {
						log.Errorf(header, "save result to db error: %v", err)
					}
					wg.Done()
					return
				}
				totalMs := int64(0)
				totalCount := int64(0)
				totalMu := sync.Mutex{}
				p.Count = 9
				p.Timeout = time.Second * 10
				p.OnRecv = func(packet *ping.Packet) {
					totalMu.Lock()
					totalMs += packet.Rtt.Milliseconds()
					totalCount += 1
					totalMu.Unlock()
				}
				if err := p.Run(); err != nil {
					n.PingRTT = 999
					if err := persist.Save(n); err != nil {
						log.Errorf(header, "save result to db error: %v", err)
					}
					wg.Done()
					return
				}
				n.PingRTT = totalMs / totalCount
				if err := persist.Save(n); err != nil {
					log.Errorf(header, "save result to db error: %v", err)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		return nil
	})
}

func (h *V2rayHandler) GetAllSubscriptions(c *gin.Context) {
	res := make([]*model.Subscription, 0)
	if err := persist.MySQL().Find(&res).Error; err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *V2rayHandler) EditSubscription(c *gin.Context) {
	param := &params.UpdateV2raySubscriptionParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		subscriptionConfig := model.NewSubscription(param.SubscriptionName, param.SubscriptionAddr)
		subscriptionConfig.Id = param.SubscriptionId
		return persist.Save(subscriptionConfig)
	})
}

func (h *V2rayHandler) RefreshV2raySubscription(c *gin.Context) {
	param := &params.V2raySubscriptionIdParam{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		subscriptionConfig := &model.Subscription{Id: param.SubscriptionId}
		if err := persist.Get(subscriptionConfig); err != nil {
			return err
		}
		return h.refreshSubscription(subscriptionConfig.Id, subscriptionConfig.Name, subscriptionConfig.Addr)
	})
}

func (h *V2rayHandler) refreshSubscription(subscriptionId int64, subscriptionName, subscriptionUrl string) error {
	if h.v2raySubscriptionRefreshing.Load() {
		return fmt.Errorf("refreshing is already doing")
	}
	h.v2raySubscriptionRefreshing.Store(true)
	header := log.NewHeader("RefreshV2raySubscription")
	log.Infof(header, "starting fetch subscription %v: %v", subscriptionName, subscriptionUrl)
	resp, err := http.Get(subscriptionUrl)
	if err != nil {
		return err
	}
	log.Infof(header, "fetch %v response code: %v, status: %v, content length: %v", subscriptionUrl, resp.StatusCode, resp.Status, resp.ContentLength)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http request fail")
	}
	res := make([]*model.V2rayNode, 0)
	decodeBs, err := ioutil.ReadAll(base64.NewDecoder(base64.RawStdEncoding, resp.Body))
	if err != nil {
		errMsg := fmt.Errorf("decode response body error: %v", err)
		log.Errorf(header, "%v", errMsg)
		return errMsg
	}

	for _, line := range strings.Split(string(decodeBs), "\n") {
		if line == "" {
			continue
		}
		s := strings.TrimRight(strings.TrimPrefix(line, "vmess://"), "=")
		bs, err := base64.RawStdEncoding.DecodeString(s)
		if err != nil {
			return err
		}
		if len(bs) == 0 {
			continue
		}
		server := model.NewV2rayNode(subscriptionId, subscriptionName)
		if err := json.Unmarshal(bs, &server); err != nil {
			errMsg := fmt.Errorf("unmarshal %v error: %v", string(bs), err)
			log.Errorf(header, "%v", errMsg)
			return errMsg
		}
		res = append(res, server)
	}
	if err := persist.MySQL().Exec("delete from v2ray_nodes where subscription_id =?", subscriptionId).Error; err != nil {
		log.Errorf(header, "truncate table v2ray_nodes fail: %v", err)
		return err
	}
	for _, server := range res {
		if err := persist.MySQL().Create(&server).Error; err != nil {
			log.Errorf(header, "save v2ray server to db error: %v", err)
		}
	}
	return nil
}

func (h *V2rayHandler) UpdateConfig(c *gin.Context) {
	param := &params.V2rayConfig{}
	if err := c.Bind(param); err != nil {
		return
	}
	h.DoFunc(c, func() error {
		h.v2rayConfigMu.Lock()
		defer h.v2rayConfigMu.Unlock()
		if _, err := h.validateConfig(param.ConfigContent, h.v2rayCurrentNode); err != nil {
			return err
		}
		h.v2rayConfigTemplateContent = param.ConfigContent
		f, err := os.Create(h.v2rayConfigTemplateFilePath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(param.ConfigContent)
		return err
	})

}
