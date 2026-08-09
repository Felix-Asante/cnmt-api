package transfers

import (
	"net/http"

	"cnmt/internal/common/httpx"
)

type Controller struct {
	svc *Service
}

func NewController(service *Service) *Controller {
	return &Controller{svc: service}
}

func (c *Controller) createTransfer(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[createTransferRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	transfer, err := c.svc.CreateTransfer(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, transfer)
}
