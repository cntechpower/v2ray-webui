package handler

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

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

type cronLogger struct {
	h *log.Header
}

func newCronLogger() *cronLogger {
	return &cronLogger{h: log.NewHeader("cron")}
}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof(l.h, msg, keysAndValues...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Errorf(l.h, fmt.Sprintf("got error %v, msg: %v", err, msg), keysAndValues...)
}

type PacHandler struct {
	currentPac  *atomic.String
	cronSpec    *atomic.String
	cronMu      sync.Mutex
	cron        *cron.Cron
	cronEntryId cron.EntryID
	proxyAddr   *atomic.String
}

func newPacHandler() (*PacHandler, error) {
	currentPac := model.NewPacContent("")
	if err := persist.DB.Order("id desc").Limit(1).Find(&currentPac).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h := &PacHandler{
		currentPac:  atomic.NewString(currentPac.Content),
		cronSpec:    atomic.NewString(config.Config.PacHandlerConfig.PacGenerateCron),
		cronMu:      sync.Mutex{},
		cron:        cron.New(cron.WithLogger(newCronLogger())),
		cronEntryId: 0,
		proxyAddr:   atomic.NewString(config.Config.PacHandlerConfig.PacProxyAddr),
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

func (h *PacHandler) ListCustomWebsites() ([]*model.PacWebSite, error) {
	res := make([]*model.PacWebSite, 0)
	return res, persist.DB.Find(&res).Error
}

func (h *PacHandler) AddCustomWebsite(webSite string) error {
	if err := checker.Var(webSite, fqdn); err != nil {
		return err
	}

	if err := persist.Create(model.NewPacWebSite(webSite)); err != nil {
		return err
	}

	return nil
}

func (h *PacHandler) DelCustomWebsites(id int64) error {
	return persist.Delete(model.NewPacWebSiteForDelete(id))
}

func (h *PacHandler) GetCurrentCron() string {
	return h.cronSpec.Load()
}

func (h *PacHandler) UpdateCron(cronString string) error {
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

func (h *PacHandler) getPacGenerateCmd() (string, error) {
	domainInDB := make([]*model.PacWebSite, 0)
	err := persist.DB.Find(&domainInDB).Error
	if err != nil {
		return "", err
	}
	domains := make([]string, 0, len(domainInDB))
	for _, domain := range domainInDB {
		domains = append(domains, "||"+domain.WebSiteUrl)
	}
	addr := h.proxyAddr.Load()
	return fmt.Sprintf(pacGenerateCommand, addr, addr, strings.Join(domains, ",")), nil
}

func (h *PacHandler) updatePacFunc(cmd string) error {
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

func (h *PacHandler) ManualGeneratePac() error {
	cmd, err := h.getPacGenerateCmd()
	if err != nil {
		return err
	}
	return h.updatePacFunc(cmd)
}

func (h *PacHandler) GetCurrentPAC() string {
	return h.currentPac.Load()
}

func (h *PacHandler) UpdateConfig(c *model.PacHandlerConfig) error {
	if c.PacGenerateCron == "" || c.PacProxyAddr == "" {
		return fmt.Errorf("at least one config is empty")
	}
	h.proxyAddr.Store(c.PacProxyAddr)
	h.cronSpec.Store(c.PacGenerateCron)
	config.Config.PacHandlerConfig.PacProxyAddr = c.PacProxyAddr
	config.Config.PacHandlerConfig.PacGenerateCron = c.PacGenerateCron
	return config.Config.Save()
}

func (h *PacHandler) GetConfig() (*model.PacHandlerConfig, error) {
	res := &model.PacHandlerConfig{}
	res.PacProxyAddr = h.proxyAddr.Load()
	res.PacGenerateCron = h.cronSpec.Load()
	return res, nil
}
