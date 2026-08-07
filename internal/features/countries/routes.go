package countries

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Route("/countries", func(r chi.Router) {
		r.Get("/", c.getCountries)
	})
}