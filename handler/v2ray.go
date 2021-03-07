package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"go.uber.org/atomic"
	v2ray "v2ray.com/core"
	_ "v2ray.com/core/app/proxyman/inbound"
	_ "v2ray.com/core/app/proxyman/outbound"
	v2rayConf "v2ray.com/core/infra/conf/serial"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/model"
	"github.com/cntechpower/v2ray-webui/persist"
)

type v2rayHandler struct {
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
	v2rayServer          v2ray.Server
	v2rayServerMu        sync.Mutex
	v2rayServerStartTime time.Time

	//use this to avoid concurrency refreshing Subscription
	v2raySubscriptionRefreshing atomic.Bool

	//status
	v2rayStatusRefreshTime time.Time
}

func newV2rayHandler(templateConfigFilePath string) (h *v2rayHandler, err error) {
	var bs []byte
	bs, err = ioutil.ReadFile(templateConfigFilePath)
	if err != nil {
		return nil, err
	}
	h = &v2rayHandler{
		baseHandler:                 &baseHandler{},
		v2rayConfigTemplateFilePath: templateConfigFilePath,
		v2rayConfigTemplateContent:  string(bs),
		v2rayConfig:                 nil,
		v2rayConfigMu:               sync.Mutex{},
		v2rayServer:                 nil,
		v2rayServerMu:               sync.Mutex{},
		v2raySubscriptionRefreshing: atomic.Bool{},
	}
	go h.refreshStatusLoop()
	return

}

func (h *v2rayHandler) TearDown() {
	h.v2rayServerMu.Lock()
	defer h.v2rayServerMu.Unlock()
	if h.v2rayServer != nil {
		header := log.NewHeader("v2rayHandler.TearDown")
		log.Infof(header, "stopping v2ray...")
		_ = h.v2rayServer.Close()
		h.v2rayServer = nil
		log.Infof(header, "stopped v2ray...")
	}
}

func (h *v2rayHandler) refreshStatusLoop() {
	ticker := time.NewTicker(30 * time.Second)
	header := log.NewHeader("refreshStatusLoop")
	for range ticker.C {
		h.refreshCurrentNodePing(header)
		h.v2rayStatusRefreshTime = time.Now()
		header.Infof("refreshed status")
	}

}

func (h *v2rayHandler) refreshCurrentNodePing(header *log.Header) {
	h.v2rayConfigMu.Lock()
	defer h.v2rayConfigMu.Unlock()
	if h.v2rayCurrentNode != nil {
		_ = h.pingSingleNode(header, h.v2rayCurrentNode)
	}
}

func (h *v2rayHandler) validateConfig(config string, node *model.V2rayNode) (*v2ray.Config, error) {
	header := log.NewHeader("v2rayHandler.validateConfig")
	config = strings.ReplaceAll(config, "{serverHost}", node.Host)
	config = strings.ReplaceAll(config, "{serverName}", node.Name)
	config = strings.ReplaceAll(config, "{serverPath}", node.Path)
	config = strings.ReplaceAll(config, "9495945", string(node.Port))
	config = strings.ReplaceAll(config, "{serverId}", node.ServerId)
	log.Infof(header, "validate config: %v", config)
	return v2rayConf.LoadJSONConfig(strings.NewReader(config))

}

func (h *v2rayHandler) SwitchNode(nodeId int64) error {
	node := &model.V2rayNode{
		Id: nodeId,
	}
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
	if err := h.StopV2ray(); err != nil && err != ErrV2rayNotStarted {
		return err
	}
	return h.StartV2ray()
}

func (h *v2rayHandler) StartV2ray() error {
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

func (h *v2rayHandler) StopV2ray() error {
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

func (h *v2rayHandler) GetAllV2rayNodes() ([]*model.V2rayNode, error) {
	res := make([]*model.V2rayNode, 0)
	return res, persist.DB.Find(&res).Error
}

func (h *v2rayHandler) AddNode(node *model.V2rayNode) error {
	return persist.DB.Save(node).Error
}

func (h *v2rayHandler) AddSubscription(subscriptionName, subscriptionAddr string) error {
	return persist.Save(model.NewSubscription(subscriptionName, subscriptionAddr))
}

func (h *v2rayHandler) DelSubscription(subscriptionId int64) error {
	return persist.Delete(&model.Subscription{Id: subscriptionId})

}

func (h *v2rayHandler) PingAllV2rayNodes() error {
	header := log.NewHeader("PingAllV2rayNodes")
	nodes := make([]*model.V2rayNode, 0)
	if err := persist.DB.Find(&nodes).Error; err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, node := range nodes {
		wg.Add(1)
		n := node
		go func() {
			_ = h.pingSingleNode(header, n)
			wg.Done()
		}()
	}
	wg.Wait()
	return nil

}

func (h *v2rayHandler) pingSingleNode(header *log.Header, n *model.V2rayNode) (err error) {
	var p *ping.Pinger
	p, err = ping.NewPinger(n.Host)
	if err != nil {
		n.PingRTT = 999
		if err := persist.Save(n); err != nil {
			log.Errorf(header, "save result to db error: %v", err)
		}
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
	if err = p.Run(); err != nil {
		n.PingRTT = 999
		if err := persist.Save(n); err != nil {
			log.Errorf(header, "save result to db error: %v", err)
		}
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
	return
}

func (h *v2rayHandler) GetAllSubscriptions() ([]*model.Subscription, error) {
	res := make([]*model.Subscription, 0)
	return res, persist.DB.Find(&res).Error
}

func (h *v2rayHandler) EditSubscription(subscriptionId int64, subscriptionName, subscriptionAddr string) error {
	subscriptionConfig := model.NewSubscription(subscriptionName, subscriptionAddr)
	subscriptionConfig.Id = subscriptionId
	return persist.Save(subscriptionConfig)

}

func (h *v2rayHandler) RefreshV2raySubscription(subscriptionId int64) error {
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
	defer h.v2raySubscriptionRefreshing.Store(false)
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
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	res, err = h.decodeSubscription(subscriptionId, subscriptionName, bs)
	if err != nil {
		return err
	}
	if err := persist.DB.Exec("delete from v2ray_nodes where subscription_id =?", subscriptionId).Error; err != nil {
		log.Errorf(header, "truncate table v2ray_nodes fail: %v", err)
		return err
	}
	for _, server := range res {
		if err := persist.DB.Create(&server).Error; err != nil {
			log.Errorf(header, "save v2ray server to db error: %v", err)
		}
	}
	return nil
}

func (h *v2rayHandler) UpdateConfig(ConfigContent string) error {
	h.v2rayConfigMu.Lock()
	defer h.v2rayConfigMu.Unlock()
	testNode := h.v2rayCurrentNode
	if testNode == nil {
		testNode = &model.V2rayNode{
			Host:     "127.0.0.1",
			Path:     "/test",
			Port:     "9495",
			Name:     "test",
			ServerId: "aaa",
		}
	}
	if _, err := h.validateConfig(ConfigContent, testNode); err != nil {
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

func (h *v2rayHandler) GetV2rayConfigTemplateContent() string {
	return h.v2rayConfigTemplateContent
}

func (h *v2rayHandler) ValidateConfig(ConfigContent string) error {
	h.v2rayConfigMu.Lock()
	defer h.v2rayConfigMu.Unlock()
	testNode := h.v2rayCurrentNode
	if testNode == nil {
		testNode = &model.V2rayNode{
			Host:     "127.0.0.1",
			Path:     "/test",
			Port:     "9495",
			Name:     "test",
			ServerId: "aaa",
		}
	}
	_, err := h.validateConfig(ConfigContent, testNode)
	return err
}

func (h *v2rayHandler) decodeSubscription(subscriptionId int64, subscriptionName string, data []byte) (res []*model.V2rayNode, err error) {
	header := log.NewHeader("decodeSubscription")
	decodeBs, err := tryDecode(string(data))
	if err != nil {
		log.Errorf(header, "decode response body error: %v", err)
		return
	}

	var bs []byte
	for _, line := range split(string(decodeBs)) {
		if line == "" {
			continue
		}
		//s := strings.TrimRight(strings.TrimPrefix(line, "vmess://"), "=")
		s := strings.TrimPrefix(line, "vmess://")
		bs, err = tryDecode(s)
		if err != nil {
			header.Errorf("some line decode fail: %v", err)
			continue
		}
		if len(bs) == 0 {
			continue
		}
		server := model.NewV2rayNode(subscriptionId, subscriptionName)
		if err = json.Unmarshal(bs, &server); err != nil {
			log.Errorf(header, "unmarshal %v error: %v", string(bs), err)
			return
		}
		res = append(res, server)
	}
	return
}

func (h *v2rayHandler) Status() (res *model.V2rayStatus, err error) {
	res = &model.V2rayStatus{}
	h.v2rayConfigMu.Lock()
	if h.v2rayCurrentNode != nil {
		res.CurrentNode = h.v2rayCurrentNode.DeepCopy()
	}
	h.v2rayConfigMu.Unlock()

	res.Core = &model.V2rayCoreStatus{}
	h.v2rayServerMu.Lock()
	res.Core.StartTime = h.v2rayServerStartTime.Format(time.RFC3339)
	h.v2rayServerMu.Unlock()
	res.RefreshTime = h.v2rayStatusRefreshTime.Format(time.RFC3339)
	return
}

func tryDecode(str string) (de []byte, err error) {
	de, err = base64.StdEncoding.DecodeString(str)
	if err == nil {
		return
	}

	de, err = base64.RawStdEncoding.DecodeString(str)
	if err == nil {
		return
	}
	de, err = base64.URLEncoding.DecodeString(str)
	if err == nil {
		return
	}
	de, err = base64.RawURLEncoding.DecodeString(str)
	if err == nil {
		return
	}

	return nil, fmt.Errorf("no proper base64 decode method for: " + str)
}

func encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

var sep = map[rune]bool{
	' ':  true,
	'\n': true,
	',':  true,
	';':  true,
	'\t': true,
	'\f': true,
	'\v': true,
	'\r': true,
}

func split(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return sep[r]
	})
}
