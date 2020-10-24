package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
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

func AddProxyHandler(engine *gin.Engine) (teardownFunc func()) {
	//proxy handler
	{
		proxyGroup := engine.Group("/proxy")
		h, err := NewProxyHandler()
		if err != nil {
			panic(err)
		}
		{
			webSiteGroup := proxyGroup.Group("/website")
			{
				webSiteGroup.GET("/list", h.ListCustomProxyWebsites)
				webSiteGroup.GET("/listv2", h.ListCustomProxyWebsitesWithoutCache)
				webSiteGroup.GET("/listv3", h.ListCustomProxyWebsitesInOneCache)
				webSiteGroup.POST("/add", h.AddCustomProxyWebsites)
				webSiteGroup.POST("/del", h.DelCustomProxyWebsites)
			}

		}
		{
			pacGroup := proxyGroup.Group("/pac")
			{
				pacGroup.GET("", h.GetCurrentPAC)
				pacGroup.POST("/cron", h.UpdateCron)
				pacGroup.GET("/cron", h.GetCurrentCron)
				pacGroup.POST("/generate", h.ManualGeneratePac)
			}

		}
		{
			v2rayGroup := proxyGroup.Group("/v2ray")
			{
				subscriptionGroup := v2rayGroup.Group("/subscription")
				subscriptionGroup.GET("/servers", h.GetAllV2rayServer)
				subscriptionGroup.POST("/add", h.AddSubscription)
				subscriptionGroup.POST("/delete", h.DelSubscription)
				subscriptionGroup.GET("/list", h.GetAllSubscriptions)
				subscriptionGroup.POST("/edit", h.EditSubscription)
				subscriptionGroup.POST("/refresh", h.RefreshV2raySubscription)
			}
		}
	}
	return func() {}
}

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
	h := &ProxyHandler{
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

	config.Config.ProxyHandlerConfig.PacGenerateCron = cronSpec
	if err := config.Config.Save("./api.config"); err != nil {
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

func (h *ProxyHandler) GetAllV2rayServer(c *gin.Context) {
	res := make([]*model.V2rayServer, 0)
	if err := persist.MySQL().Find(&res).Error; err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProxyHandler) AddSubscription(c *gin.Context) {
	addr := c.PostForm("addr")
	if addr == "" {
		errorWith500(c, fmt.Errorf("addr is required"))
		return
	}
	if err := persist.Save(model.NewSubscription(addr)); err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, "add subscription success"))
}

func (h *ProxyHandler) DelSubscription(c *gin.Context) {
	s := c.PostForm("id")
	if s == "" {
		errorWith500(c, fmt.Errorf("id is required"))
		return
	}
	id, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		errorWith500(c, fmt.Errorf("id is invalid"))
		return
	}
	if err := persist.Delete(&model.Subscription{Id: id}); err != nil {
		errorWith500(c, err)
	}
	c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, "delete subscription success"))
}

func (h *ProxyHandler) EditSubscription(c *gin.Context) {
	s := c.PostForm("id")
	if s == "" {
		errorWith500(c, fmt.Errorf("id is required"))
		return
	}
	id, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		errorWith500(c, fmt.Errorf("id is invalid"))
		return
	}
	addr := c.PostForm("addr")
	if addr == "" {
		errorWith500(c, fmt.Errorf("addr is required"))
	}
	subscriptionConfig := model.NewSubscription(addr)
	subscriptionConfig.Id = id
	if err := persist.Save(subscriptionConfig); err != nil {
		errorWith500(c, err)
	}
	c.JSON(http.StatusOK, model.NewGenericStatus(http.StatusOK, "edit subscription success"))
}

func (h *ProxyHandler) GetAllSubscriptions(c *gin.Context) {
	res := make([]*model.Subscription, 0)
	if err := persist.MySQL().Find(&res).Error; err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProxyHandler) RefreshV2raySubscription(c *gin.Context) {
	s := c.PostForm("subscription_id")
	if s == "" {
		errorWith500(c, fmt.Errorf("subscription_id is required"))
		return
	}
	id, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		errorWith500(c, fmt.Errorf("subscription_id is invalid"))
		return
	}
	subscriptionConfig := model.Subscription{Id: id}
	if err := persist.Get(&subscriptionConfig); err != nil {
		errorWith500(c, err)
		return
	}
	res, err := h.refreshSubscription(subscriptionConfig.Id, subscriptionConfig.Addr)
	if err != nil {
		errorWith500(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProxyHandler) refreshSubscription(subscriptionId int64, subscriptionUrl string) ([]*model.V2rayServer, error) {

	header := log.NewHeader("RefreshV2raySubscription")
	resp, err := http.Get(subscriptionUrl)
	if err != nil {
		return nil, err
	}
	log.Infof(header, "%v response code: %v, status: %v, content length: %v", subscriptionUrl, resp.StatusCode, resp.Status, resp.ContentLength)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request fail")
	}
	res := make([]*model.V2rayServer, 0)
	decodeBs, err := ioutil.ReadAll(base64.NewDecoder(base64.RawStdEncoding, resp.Body))
	if err != nil {
		errMsg := fmt.Errorf("decode response body error: %v", err)
		log.Errorf(header, "%v", errMsg)
		return nil, errMsg
	}

	for _, line := range strings.Split(string(decodeBs), "\n") {
		if line == "" {
			continue
		}
		s := strings.TrimRight(strings.TrimPrefix(line, "vmess://"), "=")
		bs, err := base64.RawStdEncoding.DecodeString(s)
		if err != nil {
			return nil, err
		}
		if len(bs) == 0 {
			continue
		}
		//TODO: support multi Subscription
		server := model.NewV2rayServer(subscriptionId)
		if err := json.Unmarshal(bs, &server); err != nil {
			errMsg := fmt.Errorf("unmarshal %v error: %v", string(bs), err)
			log.Errorf(header, "%v", errMsg)
			return nil, errMsg
		}
		res = append(res, server)
	}
	if err := persist.MySQL().Exec("delete from v2ray_servers where subscription_id =?", subscriptionId).Error; err != nil {
		log.Errorf(header, "truncate table v2ray_servers fail: %v", err)
	}
	for _, server := range res {
		if err := persist.MySQL().Create(&server).Error; err != nil {
			log.Errorf(header, "save v2ray server to db error: %v", err)
		}
	}
	return res, nil
}
