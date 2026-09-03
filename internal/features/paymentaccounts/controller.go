package paymentaccounts

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	countryID, err := parseOptionalInt64(q.Get("country_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid country_id", httpx.BadRequestError))
		return
	}

	paymentMethod := q.Get("payment_method")
	if paymentMethod != "" && paymentMethod != "BANK" && paymentMethod != "MOBILE_MONEY" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: payment_method must be BANK or MOBILE_MONEY", httpx.BadRequestError))
		return
	}

	isActive := q.Get("is_active")
	if isActive != "" && isActive != "true" && isActive != "false" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: is_active must be true or false", httpx.BadRequestError))
		return
	}

	currencyCode := strings.ToUpper(strings.TrimSpace(q.Get("currency_code")))
	if currencyCode != "" && len(currencyCode) != 3 {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: currency_code must be 3 characters", httpx.BadRequestError))
		return
	}

	accounts, err := c.svc.List(r.Context(), countryID, paymentMethod, isActive, currencyCode)
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

func parseOptionalInt64(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}
