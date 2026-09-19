package promocodes

import "net/http"

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	
}

func (c *Controller) get(w http.ResponseWriter, r *http.Request) {
	
}

