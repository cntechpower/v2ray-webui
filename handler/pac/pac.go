package pac

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/robfig/cron/v3"
	"go.uber.org/atomic"
	"gorm.io/gorm"

	"github.com/cntechpower/utils/log"
	"github.com/cntechpower/v2ray-webui/config"
	"github.com/cntechpower/v2ray-webui/model"
	"github.com/cntechpower/v2ray-webui/persist"
)

const (
	fqdn               = "required,fqdn"
	pacGenerateCommand = `/usr/local/bin/genpac --format pac --gfwlist-proxy '%v' --pac-proxy '%v' --user-rule "%v"`
)

type Handler struct {
	checker          *validator.Validate
	currentPac       *atomic.String
	cronSpec         *atomic.String
	cronMu           sync.Mutex
	cron             *cron.Cron
	cronEntryId      cron.EntryID
	proxyGenerateCmd *atomic.String
}

func New() (*Handler, error) {
	currentPac := model.NewPacContent("")
	if err := persist.DB.Order("id desc").Limit(1).Find(&currentPac).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h := &Handler{
		checker:          validator.New(),
		currentPac:       atomic.NewString(currentPac.Content),
		cronSpec:         atomic.NewString(config.Config.PacHandlerConfig.PacGenerateCron),
		cronMu:           sync.Mutex{},
		cron:             cron.New(cron.WithLogger(newCronLogger())),
		cronEntryId:      0,
		proxyGenerateCmd: atomic.NewString(config.Config.PacHandlerConfig.PacGenerateCmd),
	}
	if config.Config.PacHandlerConfig.PacGenerateCron != "" {
		entryId, err := h.cron.AddFunc(config.Config.PacHandlerConfig.PacGenerateCron, func() {
			_ = h.updatePacFunc("")
		})
		if err != nil {
			return nil, err
		}
		h.cronEntryId = entryId
		h.cron.Start()
	}
	return h, nil
}

func (h *Handler) GetCurrentCron() string {
	return h.cronSpec.Load()
}

func (h *Handler) UpdateCron(cronString string) error {
	h.cronMu.Lock()
	defer h.cronMu.Unlock()
	entryId, err := h.cron.AddFunc(cronString, func() {
		_ = h.updatePacFunc("")
	})
	if err != nil {
		return err
	}
	h.cronSpec.Store(cronString)
	if h.cronEntryId != 0 {
		h.cron.Remove(h.cronEntryId)
	}
	h.cronEntryId = entryId

	config.Config.PacHandlerConfig.PacGenerateCron = cronString
	if err := config.Config.Save(); err != nil {
		return err
	}
	return nil
}

func (h *Handler) getPacGenerateCmd() (string, error) {
	domainInDB := make([]*model.PacWebSite, 0)
	err := persist.DB.Find(&domainInDB).Error
	if err != nil {
		return "", err
	}
	domains := make([]string, 0, len(domainInDB))
	for _, domain := range domainInDB {
		domains = append(domains, "||"+domain.WebSiteUrl)
	}
	return fmt.Sprintf(h.proxyGenerateCmd.Load(), strings.Join(domains, ",")), nil
}

func (h *Handler) updatePacFunc(cmd string) error {
	header := log.NewHeader("update_pac")
	var err error
	if cmd == "" {
		cmd, err = h.getPacGenerateCmd()
		if err != nil {
			log.Errorf(header, "get pac cmd error: %v", err)
			return err
		}
	}

	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		errMsg := fmt.Errorf("error: %v, output: %v", err, string(output))
		log.Errorf(header, "%v", errMsg)
		return errMsg
	}
	content := string(output)
	h.currentPac.Store(content)
	if err := persist.Create(model.NewPacContent(content)); err != nil {
		log.Errorf(header, "save current pac to db error: %v", err)
		return err
	}
	return nil
}

func (h *Handler) ManualGeneratePac() error {
	cmd, err := h.getPacGenerateCmd()
	if err != nil {
		return err
	}
	return h.updatePacFunc(cmd)
}

func (h *Handler) GetCurrentPAC() string {
	return h.currentPac.Load()
}

func (h *Handler) UpdateConfig(c *model.PacHandlerConfig) error {
	if c.PacGenerateCron == "" || c.PacGenerateCmd == "" {
		return fmt.Errorf("at least one config is empty")
	}
	h.proxyGenerateCmd.Store(c.PacGenerateCmd)
	h.cronSpec.Store(c.PacGenerateCron)
	config.Config.PacHandlerConfig.PacGenerateCmd = c.PacGenerateCmd
	config.Config.PacHandlerConfig.PacGenerateCron = c.PacGenerateCron
	return config.Config.Save()
}

func (h *Handler) GetConfig() (*model.PacHandlerConfig, error) {
	res := &model.PacHandlerConfig{}
	res.PacGenerateCmd = h.proxyGenerateCmd.Load()
	res.PacGenerateCron = h.cronSpec.Load()
	return res, nil
}
