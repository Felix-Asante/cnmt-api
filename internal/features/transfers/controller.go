package transfers

import (
	"fmt"
	"net/http"

	"cnmt/internal/common/httpx"
	"cnmt/internal/features/auth"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Controller struct {
	svc *Service
}

func NewController(service *Service) *Controller {
	return &Controller{svc: service}
}

func (c *Controller) createTransfer(w http.ResponseWriter, r *http.Request) {

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: idempotency key is required", httpx.BadRequestError))
		return
	}

	body, err := httpx.DecodeAndValidate[createTransferRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	transfer, err := c.svc.CreateTransfer(r.Context(), body, idemKey)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, transfer)
}


func (c *Controller) getAllTransfers(w http.ResponseWriter, r *http.Request) {
	query, err := parseGetAllTransfersQuery(r)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	resp, err := c.svc.GetAllTransfers(r.Context(), query)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) getTransferByReference(w http.ResponseWriter, r *http.Request) {
	reference := chi.URLParam(r, "reference")
	if reference == "" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: reference is required", httpx.BadRequestError))
		return
	}

	transfer, err := c.svc.GetByReference(r.Context(), reference)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, transfer)
}

func (c *Controller) createPaymentProofSignedUrl(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[createPaymentProofSignedUrlRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := c.svc.CreatePaymentProofSignedUrl(r.Context(), body.Reference, body.ContentType)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) confirmPaymentProof(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[confirmPaymentProofRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = c.svc.ConfirmPaymentProof(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK,map[string]bool{"success": true})
}

func (c *Controller) parseTransferID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := chi.URLParam(r, "id")
	id, err := uuid.Parse(raw)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: invalid transfer id", httpx.BadRequestError))
		return uuid.Nil, false
	}
	return id, true
}

func (c *Controller) verifyPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseTransferID(w, r)
	if !ok {
		return
	}

	resp, err := c.svc.VerifyPayment(r.Context(), id, auth.ActorFromContext(r.Context()))
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) rejectPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseTransferID(w, r)
	if !ok {
		return
	}

	body, err := httpx.DecodeAndValidate[adminActionRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if body.Reason == nil || *body.Reason == "" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: reason is required", httpx.BadRequestError))
		return
	}

	resp, err := c.svc.RejectPayment(r.Context(), id, auth.ActorFromContext(r.Context()), *body.Reason)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) processTransfer(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseTransferID(w, r)
	if !ok {
		return
	}

	resp, err := c.svc.ProcessTransfer(r.Context(), id, auth.ActorFromContext(r.Context()))
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) completeTransfer(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseTransferID(w, r)
	if !ok {
		return
	}

	resp, err := c.svc.CompleteTransfer(r.Context(), id, auth.ActorFromContext(r.Context()))
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) cancelTransfer(w http.ResponseWriter, r *http.Request) {
	id, ok := c.parseTransferID(w, r)
	if !ok {
		return
	}

	body, err := httpx.DecodeAndValidate[adminActionRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if body.Reason == nil || *body.Reason == "" {
		httpx.WriteError(w, http.StatusBadRequest, fmt.Errorf("%w: reason is required", httpx.BadRequestError))
		return
	}

	resp, err := c.svc.CancelTransfer(r.Context(), id, auth.ActorFromContext(r.Context()), *body.Reason)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) getTransferOptions(w http.ResponseWriter, r *http.Request) {
	resp, err := c.svc.GetTransferOptions(r.Context())
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}