package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type Config struct {
	Type            string
	PacGenerateCron string
}

func NewConfig() *Config {
	return &Config{
		Type: "global",
	}
}

func (c *Config) GetCacheKey() string {
	return fmt.Sprintf("api_global_config")
}

func (c *Config) GetCacheDuration() time.Duration {
	return time.Hour * 24 * 365
}

func (c *Config) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
func (c *Config) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}
