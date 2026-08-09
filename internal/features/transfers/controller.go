package transfers

import (
	"fmt"
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

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: idempotency key is required", httpx.BadRequestError))
		return
	}

	body, err := httpx.DecodeAndValidate[createTransferRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	transfer, err := c.svc.CreateTransfer(r.Context(), body, idemKey)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, transfer)
}
