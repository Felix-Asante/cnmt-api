package paymentaccounts

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Get("/countries/{countryID}/payment-accounts", c.listByCountry)
}

func (c *Controller) AdminRoutes(r chi.Router) {
	r.Route("/admin/payment-accounts", func(r chi.Router) {
		r.Post("/", c.create)
	})
}
