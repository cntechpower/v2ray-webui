package config

type systemdHandlerConfig struct {
	monitorServiceNames []string
}

func (s *systemdHandlerConfig) Validate() error {
	return nil
}
