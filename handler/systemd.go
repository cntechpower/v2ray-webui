package handler

import "github.com/coreos/go-systemd/v22/dbus"

type SystemdHandler struct {
	*baseHandler
	serviceList []string
}

func NewSystemdHandler(serviceList []string) *SystemdHandler {
	return &SystemdHandler{
		baseHandler: &baseHandler{},
		serviceList: serviceList,
	}
}

func (h *SystemdHandler) CheckService(serviceNames []string) ([]dbus.UnitStatus, error) {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.ListUnitsByNames(serviceNames)

}
