package transfers

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Route("/transfers", func(r chi.Router) {
		r.Post("/", c.createTransfer)
		r.Get("/{reference}", c.getTransferByReference)
		r.Post("/payment-proof/upload-url", c.createPaymentProofSignedUrl)
	})
}