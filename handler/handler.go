package handler

import (
	"github.com/cntechpower/v2ray-webui/handler/file"
	"github.com/cntechpower/v2ray-webui/handler/pac"
	"github.com/cntechpower/v2ray-webui/handler/v2ray"
)

var Pac *pac.Handler
var V2ray *v2ray.Handler
var File *file.Handler

func Init(v2rayTemplateConfigFilePath, v2rayTrojanConfigTemplatePath string) (err error) {
	Pac, err = pac.New()
	if err != nil {
		return
	}
	V2ray, err = v2ray.New(v2rayTemplateConfigFilePath, v2rayTrojanConfigTemplatePath)
	if err != nil {
		return
	}
	File = file.New()
	return
}
