package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PacWebSite struct {
	gorm.Model
	WebSiteUrl string `validate:"required"`
}

type PacWebSiteToRender struct {
	Id         uint   `json:"id"`
	WebSiteUrl string `validate:"required" json:"url"`
}

func NewPacWebSite(s string) *PacWebSite {
	return &PacWebSite{WebSiteUrl: s}
}

func NewPacWebSiteForDelete(id int) *PacWebSite {
	return &PacWebSite{
		Model: gorm.Model{ID: uint(id)},
	}
}

func (p *PacWebSite) GetCacheKey() string {
	return fmt.Sprintf("proxy_web_site_%v", p.ID)
}

func (p *PacWebSite) GetCacheDuration() time.Duration {
	return time.Minute
}

func (p *PacWebSite) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}
func (p *PacWebSite) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p *PacWebSite) TableName() string {
	return "pac_websites"
}

func (p *PacWebSite) Render() *PacWebSiteToRender {
	return &PacWebSiteToRender{
		Id:         p.ID,
		WebSiteUrl: p.WebSiteUrl,
	}
}

func RenderPacWebSites(before []*PacWebSite) []*PacWebSiteToRender {
	res := make([]*PacWebSiteToRender, 0, len(before))
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

func (c *PacContent) TableName() string {
	return "pac_contents"
}
