package countries

import (
	"cnmt/internal/common/httpx"
	"net/http"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) getCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := c.svc.GetCountries(r.Context())
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, countries)
}