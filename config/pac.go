package config

import "fmt"

type pacHandlerConfig struct {
	PacGenerateCron string
	PacProxyAddr    string
}

func (p *pacHandlerConfig) Validate() error {
	if p.PacGenerateCron == "" {
		return fmt.Errorf("pac generate cron is empty")
	}
	return nil
}
