package handler

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/atomic"
	"gorm.io/gorm"

	"cntechpower.com/api-server/config"
	"cntechpower.com/api-server/log"
	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/persist"
)

const (
	fqdn               = "required,fqdn"
	pacGenerateCommand = `/usr/local/bin/genpac --format pac --gfwlist-proxy 'SOCKS5 10.0.0.2:1081' --pac-proxy 'SOCKS5 10.0.0.2:1081' --user-rule "%v"`
)

type CronLogger struct {
	h *log.Header
}

func NewCronLogger() *CronLogger {
	return &CronLogger{h: log.NewHeader("cron")}
}

func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof(l.h, msg, keysAndValues...)
}

func (l *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Errorf(l.h, fmt.Sprintf("got error %v, msg: %v", err, msg), keysAndValues...)
}

type PacHandler struct {
	currentPac  *atomic.String
	cronSpec    *atomic.String
	cronMu      sync.Mutex
	cron        *cron.Cron
	cronEntryId cron.EntryID
}

func NewPacHandler() (*PacHandler, error) {
	currentPac := model.NewPacContent("")
	if err := persist.MySQL().Order("id desc").Limit(1).Find(&currentPac).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h := &PacHandler{
		currentPac:  atomic.NewString(currentPac.Content),
		cronSpec:    atomic.NewString(config.Config.ProxyHandlerConfig.PacGenerateCron),
		cronMu:      sync.Mutex{},
		cron:        cron.New(cron.WithLogger(NewCronLogger())),
		cronEntryId: 0,
	}
	if config.Config.ProxyHandlerConfig.PacGenerateCron != "" {
		entryId, err := h.cron.AddFunc(config.Config.ProxyHandlerConfig.PacGenerateCron, func() {
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

func (h *PacHandler) ListCustomProxyWebsites() ([]*model.PacWebSite, error) {
	return persist.GetAllCustomProxyWebsites()
}

func (h *PacHandler) ListCustomProxyWebsitesInOneCache(c *gin.Context) {
	res, err := persist.GetAllCustomProxyWebsitesInOneCache()
	if err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, model.RenderPacWebSites(res))
}

func (h *PacHandler) ListCustomProxyWebsitesWithoutCache(c *gin.Context) {
	res := make([]*model.PacWebSite, 0)
	err := persist.MySQL().Find(&res).Error
	if err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, model.RenderPacWebSites(res))
}

func (h *PacHandler) AddCustomPacWebsites(webSite string) error {
	if err := checker.Var(webSite, fqdn); err != nil {
		return err
	}

	if err := persist.Create(model.NewPacWebSite(webSite)); err != nil {
		return err
	}

	return nil
}

func (h *PacHandler) DelCustomPacWebsites(id int64) error {
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

	config.Config.ProxyHandlerConfig.PacGenerateCron = cronString
	if err := config.Config.Save("./api.config"); err != nil {
		return err
	}
	return nil
}

func (h *PacHandler) getPacGenerateCmd() (string, error) {
	domainInDB := make([]*model.PacWebSite, 0)
	err := persist.MySQL().Find(&domainInDB).Error
	if err != nil {
		return "", err
	}
	domains := make([]string, 0, len(domainInDB))
	for _, domain := range domainInDB {
		domains = append(domains, "||"+domain.WebSiteUrl)
	}
	return fmt.Sprintf(pacGenerateCommand, strings.Join(domains, ",")), nil
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
