package auth

import (
	"net/http"

	"cnmt/internal/common/httpx"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[LoginRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := c.svc.Login(r.Context(), body)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (c *Controller) createUser(w http.ResponseWriter, r *http.Request) {
	body, err := httpx.DecodeAndValidate[CreateUserRequest](r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err)
		return
	}

	resp, err := c.svc.CreateUser(r.Context(), body, r.Header.Get("X-Bootstrap-Secret"))
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (c *Controller) me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
		return
	}

	resp, err := c.svc.GetMe(r.Context(), user.ID)
	if err != nil {
		httpx.WriteError(w, httpx.StatusFromError(err), err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}
