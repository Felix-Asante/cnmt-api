package promocodes

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Route("/promocodes", func(r chi.Router) {
		r.Get("/{code}", c.getByCode)
	})
}

func (c *Controller) AdminRoutes(r chi.Router) {
	r.Route("/admin/promocodes", func(r chi.Router) {
		r.Get("/", c.getAll)
		r.Post("/", c.create)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", c.getByID)
			r.Patch("/", c.update)
			r.Delete("/", c.delete)
		})
	})
}
