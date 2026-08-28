package auth

import (
	"net/http"

	"cnmt/internal/common/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type Middleware struct {
	svc     *Service
	jwtAuth *jwtauth.JWTAuth
}

func NewMiddleware(svc *Service, jwtAuth *jwtauth.JWTAuth) *Middleware {
	return &Middleware{svc: svc, jwtAuth: jwtAuth}
}

func (m *Middleware) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(m.jwtAuth)
}

func (m *Middleware) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
			return
		}

		userID, err := userIDFromClaims(claims)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.UnauthorizedError)
			return
		}

		user, err := m.svc.GetActiveUserByID(r.Context(), userID)
		if err != nil {
			httpx.WriteError(w, httpx.StatusFromError(err), err)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

func (m *Middleware) RequireRole(roles ...Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || !hasRole(user.Role, roles...) {
				httpx.WriteError(w, http.StatusForbidden, httpx.ForbiddenError)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func userIDFromClaims(claims map[string]interface{}) (uuid.UUID, error) {
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return uuid.Nil, httpx.UnauthorizedError
	}
	return uuid.Parse(sub)
}

func (c *Controller) Routes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", c.login)
		r.Post("/users", c.createUser)
	})
}

func (c *Controller) AuthenticatedRoutes(r chi.Router) {
		r.Get("/auth/me", c.me)
}
