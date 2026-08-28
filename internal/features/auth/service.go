package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/common/env"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"
	"cnmt/internal/infra/password"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type Service struct {
	queries         *db.Queries
	logger          *slog.Logger
	jwtAuth         *jwtauth.JWTAuth
	tokenTTL        time.Duration
	bootstrapSecret string
}

func NewService(queries *db.Queries, logger *slog.Logger, jwtAuth *jwtauth.JWTAuth, tokenTTL time.Duration) *Service {
	return &Service{
		queries:         queries,
		logger:          logger,
		jwtAuth:         jwtAuth,
		tokenTTL:        tokenTTL,
		bootstrapSecret: env.GetString("BOOTSTRAP_SECRET", ""),
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	user, err := s.getActiveUserByEmail(ctx, NormalizeEmail(req.Email))
	if err != nil {
		return LoginResponse{}, errInvalidCredentials()
	}

	if !password.Compare(user.PasswordHash, req.Password) {
		return LoginResponse{}, errInvalidCredentials()
	}

	accessToken, expiresAt, err := s.issueToken(user)
	if err != nil {
		s.logger.Error("failed to issue access token", "err", err)
		return LoginResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	mapped := mapUser(user)
	return LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		User:        toUserResponse(mapped),
	}, nil
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest, bootstrapSecret string) (UserResponse, error) {
	if err := s.validateBootstrapSecret(bootstrapSecret); err != nil {
		return UserResponse{}, err
	}

	role := Role(req.Role)
	if !role.IsValid() {
		return UserResponse{}, fmt.Errorf("%w: invalid role", httpx.BadRequestError)
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", "err", err)
		return UserResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	row, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        NormalizeEmail(req.Email),
		PasswordHash: hash,
		Role:         roleToDB(role),
	})
	if err != nil {
		s.logger.Error("failed to create user", "err", err)
		return UserResponse{}, common.TranslateDBError(err)
	}

	return toUserResponse(mapUser(row)), nil
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (UserResponse, error) {
	user, err := s.getActiveUserByID(ctx, userID)
	if err != nil {
		return UserResponse{}, err
	}
	return toUserResponse(user), nil
}

func (s *Service) GetActiveUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	return s.getActiveUserByID(ctx, id)
}

func (s *Service) issueToken(user db.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenTTL)
	claims := map[string]interface{}{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  string(user.Role),
	}
	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiry(claims, expiresAt)

	_, token, err := s.jwtAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (s *Service) validateBootstrapSecret(provided string) error {
	if s.bootstrapSecret == "" {
		return fmt.Errorf("%w: user bootstrap is not configured", httpx.ServiceUnavailableError)
	}

	if !secureCompare(provided, s.bootstrapSecret) {
		return fmt.Errorf("%w: invalid bootstrap secret", httpx.ForbiddenError)
	}

	return nil
}

func (s *Service) getActiveUserByEmail(ctx context.Context, email string) (db.User, error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, errInvalidCredentials()
	}
	if !user.IsActive {
		return db.User{}, errInvalidCredentials()
	}
	return user, nil
}

func (s *Service) getActiveUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, common.TranslateDBError(err)
	}
	if !row.IsActive {
		return User{}, fmt.Errorf("%w", httpx.UnauthorizedError)
	}
	return mapUser(row), nil
}

func mapUser(row db.User) User {
	return User{
		ID:        row.ID,
		Email:     row.Email,
		Role:      roleFromDB(row.Role),
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
	}
}

func errInvalidCredentials() error {
	return fmt.Errorf("%w: invalid email or password", httpx.UnauthorizedError)
}

func secureCompare(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
