package transfers

import "github.com/go-chi/chi/v5"

func (c *Controller) Routes(r chi.Router) {
	r.Route("/transfers", func(r chi.Router) {
		r.Post("/", c.createTransfer)
		r.Get("/", c.getAllTransfers)
		r.Get("/options", c.getTransferOptions)
		r.Get("/{reference}", c.getTransferByReference)
		r.Post("/payment-proof/upload-url", c.createPaymentProofSignedUrl)
		r.Patch("/payment-proof/confirm", c.confirmPaymentProof)
	})
}

func (c *Controller) AdminRoutes(r chi.Router) {
	r.Route("/admin/transfers", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Post("/verify-payment", c.verifyPayment)
			r.Post("/reject-payment", c.rejectPayment)
			r.Post("/process", c.processTransfer)
			r.Post("/complete", c.completeTransfer)
			r.Post("/cancel", c.cancelTransfer)
		})
	})
}