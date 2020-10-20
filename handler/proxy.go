package handler

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm"

	"cntechpower.com/api-server/log"

	"github.com/robfig/cron/v3"

	"go.uber.org/atomic"

	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/persist"
	"github.com/gin-gonic/gin"
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

type ProxyHandler struct {
	currentPac  *atomic.String
	cronSpec    *atomic.String
	cronMu      sync.Mutex
	cron        *cron.Cron
	cronEntryId cron.EntryID
}

func NewProxyHandler() (*ProxyHandler, error) {
	currentPac := model.NewPacContent("")
	if err := persist.MySQL().Order("id desc").Limit(1).Find(&currentPac).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	config := &model.Config{}
	if err := persist.MySQL().Find(&config).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	h := &ProxyHandler{
		currentPac:  atomic.NewString(currentPac.Content),
		cronSpec:    atomic.NewString(config.PacGenerateCron),
		cronMu:      sync.Mutex{},
		cron:        cron.New(cron.WithLogger(NewCronLogger())),
		cronEntryId: 0,
	}
	if config.PacGenerateCron != "" {
		entryId, err := h.cron.AddFunc(config.PacGenerateCron, func() {
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

func (h *ProxyHandler) ListCustomProxyWebsites(c *gin.Context) {
	res, err := persist.GetAllCustomProxyWebsites()
	if err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, model.RenderProxyWebSites(res))
}

func (h *ProxyHandler) ListCustomProxyWebsitesInOneCache(c *gin.Context) {
	res, err := persist.GetAllCustomProxyWebsitesInOneCache()
	if err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, model.RenderProxyWebSites(res))
}

func (h *ProxyHandler) ListCustomProxyWebsitesWithoutCache(c *gin.Context) {
	res := make([]*model.ProxyWebSite, 0)
	err := persist.MySQL().Find(&res).Error
	if err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, model.RenderProxyWebSites(res))
}

func (h *ProxyHandler) AddCustomProxyWebsites(c *gin.Context) {
	for _, webSite := range c.PostFormArray("web_site") {
		if err := checker.Var(webSite, fqdn); err != nil {
			errorWith500(c, err)
			return
		}
	}
	successNames := make([]string, 0)
	for _, webSite := range c.PostFormArray("web_site") {
		if err := persist.Create(model.NewProxyWebSite(webSite)); err != nil {
			errorWith500(c, err)
			return
		}
		successNames = append(successNames, webSite)
	}
	ok(c, "add custom proxy websites %v success", successNames)
}

func (h *ProxyHandler) DelCustomProxyWebsites(c *gin.Context) {
	successIds := make([]int, 0)
	for _, webSiteId := range c.PostFormArray("web_site_id") {
		id, err := strconv.Atoi(webSiteId)
		if err != nil {
			errorWith500(c, err)
			return
		}
		if err := persist.Delete(model.NewProxyWebSiteForDelete(id)); err != nil {
			errorWith500(c, err)
			return
		}
		successIds = append(successIds, id)
	}
	ok(c, "delete custom proxy websites %v success", successIds)
}

func (h *ProxyHandler) GetCurrentCron(c *gin.Context) {
	ok(c, h.cronSpec.Load())
}

func (h *ProxyHandler) UpdateCron(c *gin.Context) {
	h.cronMu.Lock()
	defer h.cronMu.Unlock()
	cronSpec := c.PostForm("cron")
	entryId, err := h.cron.AddFunc(cronSpec, func() {
		_ = h.updatePacFunc("")
	})
	if err != nil {
		errorWith500(c, err)
		return
	}
	h.cronSpec.Store(cronSpec)
	if h.cronEntryId != 0 {
		h.cron.Remove(h.cronEntryId)
	}
	h.cronEntryId = entryId
	config := model.NewConfig()
	if err := persist.Get(config); err != nil && err != gorm.ErrRecordNotFound {
		errorWith500(c, err)
		return
	}
	config.PacGenerateCron = cronSpec
	if err := persist.MySQL().Where("type=?", "global").Save(config).Error; err != nil {
		errorWith500(c, err)
		return
	}
	ok(c, h.cronSpec.Load())
}

func (h *ProxyHandler) getPacGenerateCmd() (string, error) {
	domainInDB := make([]*model.ProxyWebSite, 0)
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

func (h *ProxyHandler) updatePacFunc(cmd string) error {
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

func (h *ProxyHandler) ManualGeneratePac(c *gin.Context) {
	cmd, err := h.getPacGenerateCmd()
	if err != nil {
		errorWith500(c, err)
		return
	}
	dryRun := c.PostForm("dry-run")
	if strings.ToUpper(dryRun) == "TRUE" {
		c.String(http.StatusOK, cmd)
		return
	}
	if err := h.updatePacFunc(cmd); err != nil {
		errorWith500(c, err)
		return
	}
	ok(c, "manual generate pac success")
}

func (h *ProxyHandler) GetCurrentPAC(c *gin.Context) {
	c.String(http.StatusOK, h.currentPac.Load())
}
