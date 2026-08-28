package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Role      Role
	IsActive  bool
	CreatedAt time.Time
}

type contextKey string

const userContextKey contextKey = "user"

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

func ActorFromContext(ctx context.Context) string {
	user, ok := UserFromContext(ctx)
	if !ok {
		return "unknown"
	}
	return user.Email
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
