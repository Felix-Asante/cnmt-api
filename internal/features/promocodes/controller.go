package promocodes

import (
	"fmt"
	"net/http"

	"cnmt/internal/common/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[CreatePromoCodeRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	promoCode, err := c.service.Create(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, promoCode)
}

func (c *Controller) getByCode(w http.ResponseWriter, r *http.Request) {
	promoCode, err := c.service.GetByCode(r.Context(), chi.URLParam(r, "code"))
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, promoCode)
}

func (c *Controller) getByID(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}
	promoCode, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, promoCode)
}

func (c *Controller) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}
	if err := c.service.Delete(r.Context(), id); err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) getAll(w http.ResponseWriter, r *http.Request) {
	promoCodes, err := c.service.GetAll(r.Context())
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, promoCodes)
}

func (c *Controller) update(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseID(w, r)
	if !ok {
		return
	}
	body, err := httpx.DecodeAndValidate[UpdatePromoCodeRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	promoCode, err := c.service.Update(r.Context(), id, body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, promoCode)
}

func (c *Controller) parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid promo code id", httpx.BadRequestError))
		return uuid.Nil, false
	}
	return id, true
}
