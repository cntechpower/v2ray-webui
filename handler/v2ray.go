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
	v2ray "v2ray.com/core"
	_ "v2ray.com/core/app/proxyman/inbound"
	_ "v2ray.com/core/app/proxyman/outbound"
	v2rayConf "v2ray.com/core/infra/conf/serial"

	"cntechpower.com/api-server/log"
	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/persist"
)

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

func (h *V2rayHandler) validateConfig(config string, node *model.V2rayNode) (*v2ray.Config, error) {
	header := log.NewHeader("V2rayHandler.validateConfig")
	config = strings.ReplaceAll(config, "{serverHostName}", node.Host)
	config = strings.ReplaceAll(config, "{serverName}", node.Name)
	config = strings.ReplaceAll(config, "{serverPath}", node.Path)
	config = strings.ReplaceAll(config, "9495945", strconv.FormatInt(node.Port, 10))
	config = strings.ReplaceAll(config, "{serverId}", node.ServerId)
	log.Infof(header, "validate config: %v", config)
	return v2rayConf.LoadJSONConfig(strings.NewReader(config))

}

func (h *V2rayHandler) SwitchNode(nodeId int64) error {
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

func (h *V2rayHandler) GetAllV2rayNodes() ([]*model.V2rayNode, error) {
	res := make([]*model.V2rayNode, 0)
	return res, persist.MySQL().Find(&res).Error
}

func (h *V2rayHandler) AddSubscription(subscriptionName, subscriptionAddr string) error {
	return persist.Save(model.NewSubscription(subscriptionName, subscriptionAddr))
}

func (h *V2rayHandler) DelSubscription(subscriptionId int64) error {
	return persist.Delete(&model.Subscription{Id: subscriptionId})

}

func (h *V2rayHandler) PingAllV2rayNodes() error {
	header := log.NewHeader("PingAllV2rayNodes")
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
			if totalCount == 0 {
				n.PingRTT = 999
			} else {
				n.PingRTT = totalMs / totalCount
			}
			if err := persist.Save(n); err != nil {
				log.Errorf(header, "save result to db error: %v", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return nil

}

func (h *V2rayHandler) GetAllSubscriptions() ([]*model.Subscription, error) {
	res := make([]*model.Subscription, 0)
	return res, persist.MySQL().Find(&res).Error
}

func (h *V2rayHandler) EditSubscription(subscriptionId int64, subscriptionName, subscriptionAddr string) error {
	subscriptionConfig := model.NewSubscription(subscriptionName, subscriptionAddr)
	subscriptionConfig.Id = subscriptionId
	return persist.Save(subscriptionConfig)

}

func (h *V2rayHandler) RefreshV2raySubscription(subscriptionId int64) error {
	subscriptionConfig := &model.Subscription{Id: subscriptionId}
	if err := persist.Get(subscriptionConfig); err != nil {
		return err
	}
	subscriptionName := subscriptionConfig.Name
	subscriptionUrl := subscriptionConfig.Addr
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

func (h *V2rayHandler) UpdateConfig(ConfigContent string) error {
	h.v2rayConfigMu.Lock()
	defer h.v2rayConfigMu.Unlock()
	if _, err := h.validateConfig(ConfigContent, h.v2rayCurrentNode); err != nil {
		return err
	}
	h.v2rayConfigTemplateContent = ConfigContent
	f, err := os.Create(h.v2rayConfigTemplateFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(ConfigContent)
	return err
}

func (h *V2rayHandler) GetV2rayConfigTemplateContent() string {
	return h.v2rayConfigTemplateContent
}
