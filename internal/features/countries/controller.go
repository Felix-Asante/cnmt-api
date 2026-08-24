package countries

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

func (c *Controller) getCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := c.svc.GetCountries(r.Context())
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, countries)
}

func (c *Controller) getDestCountries(w http.ResponseWriter, r *http.Request) {
	srcCountryID, err := strconv.ParseInt(chi.URLParam(r, "countryID"), 10, 64)
	if err != nil || srcCountryID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid country id", httpx.BadRequestError))
		return
	}

	destCountries, err := c.svc.GetDestCountriesBySrcCountryID(r.Context(), srcCountryID)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, destCountries)
}

func (c *Controller) createCountry(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[CreateCountryRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	country, err := c.svc.CreateCountry(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, country)
}

func (c *Controller) createRoute(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[CreateRouteRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	route, err := c.svc.CreateRoute(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, route)
}

func (c *Controller) listRoutes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	sourceCountryID, err := parseOptionalInt64(q.Get("source_country_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid source_country_id", httpx.BadRequestError))
		return
	}
	destCountryID, err := parseOptionalInt64(q.Get("dest_country_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid dest_country_id", httpx.BadRequestError))
		return
	}

	isActive := q.Get("is_active")
	if isActive != "" && isActive != "true" && isActive != "false" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: is_active must be true or false", httpx.BadRequestError))
		return
	}

	routes, err := c.svc.ListRoutes(r.Context(), sourceCountryID, destCountryID, isActive)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, routes)
}

func (c *Controller) updateRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseRouteID(w, r)
	if !ok {
		return
	}

	body, err := httpx.DecodeAndValidate[UpdateRouteRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	route, err := c.svc.UpdateRoute(r.Context(), id, body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, route)
}

func (c *Controller) deleteRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseRouteID(w, r)
	if !ok {
		return
	}

	if err := c.svc.DeleteRoute(r.Context(), id); err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) toggleRouteActive(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseRouteID(w, r)
	if !ok {
		return
	}

	route, err := c.svc.ToggleRouteActive(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, route)
}

func (c *Controller) parseRouteID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid route id", httpx.BadRequestError))
		return uuid.Nil, false
	}
	return id, true
}

func parseOptionalInt64(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n <= 0 {
		return 0, err
	}
	return n, nil
}
