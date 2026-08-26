package dashboard

import (
	"net/http"
	"time"

	"cnmt/internal/common/httpx"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) getDashboard(w http.ResponseWriter, r *http.Request) {
	period, err := parsePeriod(r, time.Now())
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := c.svc.Get(r.Context(), period)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}
