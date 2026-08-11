package countries

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
