package v2ray

import (
	"os"
	"strings"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/model"

	v2ray "v2ray.com/core"
	v2rayConf "v2ray.com/core/infra/conf/serial"
)

func (h *Handler) validateConfig(config string, node *model.V2rayNode) (*v2ray.Config, error) {
	header := log.NewHeader("V2rayHandler.validateConfig")
	config = strings.TrimRight(config, "\"")
	config = strings.TrimLeft(config, "\"")
	config = strings.ReplaceAll(config, "\\n", " ")
	config = strings.ReplaceAll(config, "\\", "")
	config = strings.ReplaceAll(config, "{serverHost}", node.Host)
	config = strings.ReplaceAll(config, "{serverName}", node.Name)
	config = strings.ReplaceAll(config, "{serverPath}", node.Path)
	config = strings.ReplaceAll(config, "{serverPass}", node.Password)
	config = strings.ReplaceAll(config, "9495945", string(node.Port))
	config = strings.ReplaceAll(config, "{serverId}", node.ServerId)
	log.Infof(header, "validate config: %v", config)
	return v2rayConf.LoadJSONConfig(strings.NewReader(config))

}

func (h *Handler) UpdateConfig(ConfigContent, typ string) error {
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
			Password: "aaa",
		}
	}
	if _, err := h.validateConfig(ConfigContent, testNode); err != nil {
		return err
	}
	var f *os.File
	var err error
	switch typ {
	case model.SubscriptionTypeVmess:
		h.v2rayConfigTemplateContent = ConfigContent
		f, err = os.Create(h.v2rayConfigTemplateFilePath)
	case model.SubscriptionTypeTrojan:
		h.v2rayTrojanConfigTemplateContent = ConfigContent
		f, err = os.Create(h.v2rayTrojanConfigTemplateFilePath)
	default:
		h.v2rayConfigTemplateContent = ConfigContent
		f, err = os.Create(h.v2rayConfigTemplateFilePath)
	}

	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(ConfigContent)
	return err
}

func (h *Handler) GetV2rayConfigTemplateContent() string {
	return h.v2rayConfigTemplateContent
}

func (h *Handler) GetV2rayTrojanConfigTemplateContent() string {
	return h.v2rayTrojanConfigTemplateContent
}

func (h *Handler) ValidateConfig(ConfigContent string) error {
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
