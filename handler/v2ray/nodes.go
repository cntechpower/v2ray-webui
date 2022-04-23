package v2ray

import (
	"sync"
	"time"

	core "github.com/v2fly/v2ray-core/v4"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/model"
	"github.com/cntechpower/v2ray-webui/persist"

	"github.com/go-ping/ping"
)

func (h *Handler) GetAllV2rayNodes() ([]*model.V2rayNode, error) {
	res := make([]*model.V2rayNode, 0)
	return res, persist.DB.Find(&res).Error
}

func (h *Handler) AddNode(node *model.V2rayNode) error {
	return persist.DB.Save(node).Error
}

func (h *Handler) PingAllV2rayNodes() error {
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

func (h *Handler) pingSingleNode(header *log.Header, n *model.V2rayNode) (err error) {
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

func (h *Handler) refreshCurrentNodePing(header *log.Header) {
	h.v2rayConfigMu.Lock()
	defer h.v2rayConfigMu.Unlock()
	if h.v2rayCurrentNode != nil {
		_ = h.pingSingleNode(header, h.v2rayCurrentNode)
	}
}

func (h *Handler) SwitchNode(nodeId int64) (err error) {
	node := &model.V2rayNode{
		Id: nodeId,
	}
	if err := persist.Get(node); err != nil {
		return err
	}
	var config *core.Config
	switch node.SubscriptionType {
	case model.SubscriptionTypeVmess:
		config, err = h.validateConfig(h.v2rayConfigTemplateContent, node)
	case model.SubscriptionTypeTrojan:
		config, err = h.validateConfig(h.v2rayTrojanConfigTemplateContent, node)
	default:
		config, err = h.validateConfig(h.v2rayConfigTemplateContent, node)
	}

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
