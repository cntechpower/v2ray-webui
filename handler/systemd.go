package handler

import (
	"cntechpower.com/api-server/log"
	"github.com/coreos/go-systemd/v22/dbus"
)

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
	header := log.NewHeader("check_service")
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		log.Errorf(header, "get systemd conn error: %v", err)
		return nil, err
	}
	defer conn.Close()
	res, err := conn.ListUnitsByNames(serviceNames)
	if err != nil {
		log.Errorf(header, "get service %v status error: %v", serviceNames, err)
		return nil, err
	}
	log.Infof(header, "check service %v success, state slice len: %v", serviceNames, len(res))
	return res, err
}

func (h *SystemdHandler) ListService() ([]dbus.UnitStatus, error) {
	header := log.NewHeader("check_service")
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		log.Errorf(header, "get systemd conn error: %v", err)
		return nil, err
	}
	defer conn.Close()
	res, err := conn.ListUnitsByNames(h.serviceList)
	if err != nil {
		log.Errorf(header, "list service %v status error: %v", h.serviceList, err)
		return nil, err
	}
	log.Infof(header, "list service %v success, state slice len: %v", h.serviceList, len(res))
	return res, err
}
