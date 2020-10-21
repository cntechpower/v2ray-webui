package config

import "fmt"

type proxyHandlerConfig struct {
	PacGenerateCron string
	PacFile         bool
	PacFilePath     string
}

func (p *proxyHandlerConfig) Validate() error {
	if p.PacFile && p.PacFilePath != "" {
		return fmt.Errorf("pac file generate is on, but pac file path is empty")
	}
	if p.PacGenerateCron == "" {
		return fmt.Errorf("pac generate cron is empty")
	}
	return nil
}
