package promocodes

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Route("/promocodes", func(r chi.Router) {
		r.Get("/{id}", c.get)
	})
}

func (c *Controller) AdminRoutes(r chi.Router) {
	r.Route("/admin/promocodes", func(r chi.Router) {
		r.Post("/", c.create)
	})
}