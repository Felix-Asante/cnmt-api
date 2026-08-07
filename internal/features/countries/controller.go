package countries

import "net/http"

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) getCountries(w http.ResponseWriter, r *http.Request) {}