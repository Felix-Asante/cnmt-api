package paymentaccounts

import (
	"fmt"
	"net/http"
	"strconv"

	"cnmt/internal/common/httpx"

	"github.com/go-chi/chi/v5"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) listByCountry(w http.ResponseWriter, r *http.Request) {
	countryID, err := strconv.ParseInt(chi.URLParam(r, "countryID"), 10, 64)
	if err != nil || countryID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid country id", httpx.BadRequestError))
		return
	}

	accounts, err := c.svc.ListByCountryID(r.Context(), countryID)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, accounts)
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[CreatePaymentAccountRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	account, err := c.svc.Create(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, account)
}
