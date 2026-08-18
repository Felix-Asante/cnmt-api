package countries

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Route("/countries", func(r chi.Router) {
		r.Get("/{countryID}/destinations", c.getDestCountries)
	})
}

func (c *Controller) AdminRoutes(r chi.Router) {
	r.Route("/admin/countries", func(r chi.Router) {
		r.Get("/", c.getCountries)
		r.Post("/", c.createCountry)
	})
	r.Route("/admin/routes", func(r chi.Router) {
		r.Post("/", c.createRoute)
		r.Route("/{id}", func(r chi.Router) {
			r.Patch("/", c.updateRoute)
			r.Delete("/", c.deleteRoute)
			r.Post("/toggle-active", c.toggleRouteActive)
		})
	})
}
