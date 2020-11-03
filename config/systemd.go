package config

type systemdHandlerConfig struct {
	MonitorServiceNames []string
}

func (s *systemdHandlerConfig) Validate() error {
	return nil
}
