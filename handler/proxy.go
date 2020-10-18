package handler

import (
	"net/http"
	"strconv"

	"cntechpower.com/api-server/model"
	"cntechpower.com/api-server/persist"
	"github.com/gin-gonic/gin"
)

const (
	fqdn = "required,fqdn"
)

func ListCustomProxyWebsites(c *gin.Context) {
	res, err := persist.GetAllCustomProxyWebsites()
	if err != nil {
		errorWith500(c, err)
	}
	c.JSON(http.StatusOK, model.RenderProxyWebSites(res))
}

func ListCustomProxyWebsitesInOneCache(c *gin.Context) {
	res, err := persist.GetAllCustomProxyWebsitesInOneCache()
	if err != nil {
		errorWith500(c, err)
	}
	c.JSON(http.StatusOK, model.RenderProxyWebSites(res))
}

func ListCustomProxyWebsitesWithoutCache(c *gin.Context) {
	res := make([]*model.ProxyWebSite, 0)
	err := persist.MySQL().Find(&res).Error
	if err != nil {
		errorWith500(c, err)
	}
	c.JSON(http.StatusOK, model.RenderProxyWebSites(res))
}

func AddCustomProxyWebsites(c *gin.Context) {
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

func DelCustomProxyWebsites(c *gin.Context) {
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
