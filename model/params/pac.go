package params

type UpdatePacCronParam struct {
	CronString string `form:"cron" binding:"required"`
}

type AddCustomPacWebsitesParam struct {
	WebSite string `form:"web_site" binding:"required"`
}

type DelCustomPacWebsitesParam struct {
	WebSiteId int64 `form:"website_id" binding:"required"`
}

type UpdatePacProxyAddrParam struct {
	PacAddr string `form:"addr" binding:"required"`
}
