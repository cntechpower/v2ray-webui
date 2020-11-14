package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var Config *config

type config struct {
	ListenAddr         string
	DebugMode          bool
	MysqlDSN           string
	RedisDSN           string
	ProxyHandlerConfig *proxyHandlerConfig
}

func (c *config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen addr is empty")
	}
	if c.ProxyHandlerConfig != nil {
		if err := c.ProxyHandlerConfig.Validate(); err != nil {
			panic(err)
		}
	}
	return nil
}

func (c *config) Save(configFilePath string) error {
	f, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	bs, err := yaml.Marshal(Default())
	if err != nil {
		return err
	}
	_, err = f.Write(bs)
	return err
}

func Init(configFilePath string) {
	f, err := os.Open(configFilePath)
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
		MysqlDSN:   "api:api@tcp(127.0.0.1:3306)/api?charset=utf8mb4&parseTime=True&loc=Local",
		RedisDSN:   "127.0.0.1:6379",
		ProxyHandlerConfig: &proxyHandlerConfig{
			PacGenerateCron: "0 0 * * *",
			PacFile:         false,
			PacFilePath:     "",
		},
	}
}
