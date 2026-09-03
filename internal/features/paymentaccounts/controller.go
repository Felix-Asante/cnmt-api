package paymentaccounts

import (
	"fmt"
	"net/http"
	"strconv"

	"cnmt/internal/common/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func (c *Controller) update(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}

	body, err := httpx.DecodeAndValidate[UpdatePaymentAccountRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	account, err := c.svc.Update(r.Context(), id, body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, account)
}

func (c *Controller) activate(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}

	account, err := c.svc.Activate(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, account)
}

func (c *Controller) deactivate(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}

	account, err := c.svc.Deactivate(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, account)
}

func (c *Controller) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}

	if err := c.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid payment account id", httpx.BadRequestError))
		return uuid.Nil, false
	}
	return id, true
}
