package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const defaultConfPath = "./conf/api.config"

var Config *config

type config struct {
	ListenAddr       string
	DebugMode        bool
	PacHandlerConfig *pacHandlerConfig
}

func (c *config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen addr is empty")
	}
	if c.PacHandlerConfig == nil {
		return fmt.Errorf("pac handler config not exist")
	}
	if c.PacHandlerConfig != nil {
		if err := c.PacHandlerConfig.Validate(); err != nil {
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
