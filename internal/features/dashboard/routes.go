package dashboard

import "github.com/go-chi/chi/v5"

func (c *Controller) AdminRoutes(r chi.Router) {
	r.Get("/admin/dashboard", c.getDashboard)
}
