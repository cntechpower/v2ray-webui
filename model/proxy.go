package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ProxyWebSite struct {
	gorm.Model
	WebSiteUrl string `validate:"required"`
}

type ProxyWebSiteToRender struct {
	Id         uint
	WebSiteUrl string `validate:"required"`
}

func NewProxyWebSite(s string) *ProxyWebSite {
	return &ProxyWebSite{WebSiteUrl: s}
}

func NewProxyWebSiteForDelete(id int) *ProxyWebSite {
	return &ProxyWebSite{
		Model: gorm.Model{ID: uint(id)},
	}
}

func (p *ProxyWebSite) GetCacheKey() string {
	return fmt.Sprintf("proxy_web_site_%v", p.ID)
}

func (p *ProxyWebSite) GetCacheDuration() time.Duration {
	return time.Minute
}

func (p *ProxyWebSite) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}
func (p *ProxyWebSite) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *ProxyWebSite) Render() *ProxyWebSiteToRender {
	return &ProxyWebSiteToRender{
		Id:         p.ID,
		WebSiteUrl: p.WebSiteUrl,
	}
}

func RenderProxyWebSites(before []*ProxyWebSite) []*ProxyWebSiteToRender {
	res := make([]*ProxyWebSiteToRender, 0, len(before))
	for _, p := range before {
		res = append(res, p.Render())
	}
	return res
}

type PacContent struct {
	gorm.Model
	Content string
}

func NewPacContent(content string) *PacContent {
	return &PacContent{Content: content}
}

func (c *PacContent) GetCacheKey() string {
	return fmt.Sprintf("pac_content_%v", c.ID)
}

func (c *PacContent) GetCacheDuration() time.Duration {
	return time.Minute
}

func (c *PacContent) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
func (c *PacContent) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

type V2rayServer struct {
	Id             int64  `json:"primary_key"`
	SubscriptionId int64  `json:"subscription_id"`
	Host           string `json:"host"`
	Path           string `json:"path"`
	TLS            string `json:"tls"`
	Address        string `json:"add"`
	Port           int64  `json:"port"`
	Aid            int64  `json:"aid"`
	Net            string `json:"net"`
	Type           string `json:"type"`
	V              string `json:"v"`
	Name           string `json:"ps"`
	ServerId       string `json:"id"`
}

func NewV2rayServer(subscriptionId int64) *V2rayServer {
	return &V2rayServer{SubscriptionId: subscriptionId}
}

func (s *V2rayServer) GetCacheKey() string {
	return fmt.Sprintf("pac_content_%v", s.Id)
}

func (s *V2rayServer) GetCacheDuration() time.Duration {
	return time.Minute
}

func (s *V2rayServer) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}
func (s *V2rayServer) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

type Subscription struct {
	Id   int64  `json:"id"`
	Addr string `json:"subscription_addr"`
}

func NewSubscription(addr string) *Subscription {
	return &Subscription{Addr: addr}
}

func (s *Subscription) GetCacheKey() string {
	return fmt.Sprintf("pac_content_%v", s.Id)
}

func (s *Subscription) GetCacheDuration() time.Duration {
	return time.Minute
}

func (s *Subscription) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}
func (s *Subscription) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
