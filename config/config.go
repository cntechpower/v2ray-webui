package config

import (
	"io/ioutil"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

const defaultConfPath = "./conf/api.config"

var Config *config
var checker *validator.Validate

func init() {
	checker = validator.New()
}

type config struct {
	ListenAddr       string `validate:"required"`
	DebugMode        bool
	PacHandlerConfig *pacHandlerConfig
}

type pacHandlerConfig struct {
	PacGenerateCron string `validate:"required"`
	PacProxyAddr    string `validate:"required"`
}

func (c *config) Validate() (err error) {
	if err = checker.Struct(c); err != nil {
		return
	}
	if c.PacHandlerConfig != nil {
		if err := checker.Struct(c.PacHandlerConfig); err != nil {
			panic(err)
		}
	}
	return nil
}

func (c *config) Save() error {
	f, err := os.Create(defaultConfPath)
	if err != nil {
		return err
	}
	defer f.Close()
	bs, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	_, err = f.Write(bs)
	return err
}

func Init() {
	f, err := os.Open(defaultConfPath)
	if err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	Config = &config{}
	if err := yaml.Unmarshal(bs, Config); err != nil {
		panic(err)
	}
	if err := Config.Validate(); err != nil {
		panic(err)
	}
}

func Default() *config {
	return &config{
		ListenAddr: "0.0.0.0:8888",
		DebugMode:  true,
		PacHandlerConfig: &pacHandlerConfig{
			PacGenerateCron: "0 0 * * *",
			PacProxyAddr:    "SOCKS5 10.0.0.2:1081",
		},
	}
}
