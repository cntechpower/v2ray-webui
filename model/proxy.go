package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type ProxyWebSite struct {
	gorm.Model
	WebSiteUrl string `validate:"required"`
}

type ProxyWebSiteToRender struct {
	WebSiteUrl string `validate:"required"`
}

func (p *ProxyWebSite) GetCacheKey() string {
	return fmt.Sprintf("proxy_web_site_%v", p.ID)
}

func (p *ProxyWebSite) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}
func (p *ProxyWebSite) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *ProxyWebSite) Render() *ProxyWebSiteToRender {
	return &ProxyWebSiteToRender{WebSiteUrl: p.WebSiteUrl}
}

func RenderProxyWebSites(before []*ProxyWebSite) []*ProxyWebSiteToRender {
	res := make([]*ProxyWebSiteToRender, 0, len(before))
	for _, p := range before {
		res = append(res, p.Render())
	}
	return res
}

func NewProxyWebSite(s string) *ProxyWebSite {
	return &ProxyWebSite{WebSiteUrl: s}
}

func NewProxyWebSiteForDelete(id int) *ProxyWebSite {
	return &ProxyWebSite{
		Model: gorm.Model{ID: uint(id)},
	}
}
