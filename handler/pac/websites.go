package pac

import (
	"github.com/cntechpower/v2ray-webui/model"
	"github.com/cntechpower/v2ray-webui/persist"
)

func (h *Handler) ListCustomWebsites() ([]*model.PacWebSite, error) {
	res := make([]*model.PacWebSite, 0)
	return res, persist.DB.Find(&res).Error
}

func (h *Handler) AddCustomWebsite(webSite string) error {
	if err := h.checker.Var(webSite, fqdn); err != nil {
		return err
	}

	if err := persist.Create(model.NewPacWebSite(webSite)); err != nil {
		return err
	}

	return nil
}

func (h *Handler) DelCustomWebsites(id int64) error {
	return persist.Delete(model.NewPacWebSiteForDelete(id))
}
