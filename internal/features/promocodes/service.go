package promocodes

import (
	"cnmt/internal/infra/db"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)
type Service struct {
	db      *pgxpool.Pool
	queries *db.Queries
	logger  *slog.Logger
}

type ServiceConfig struct {
	DB *pgxpool.Pool
	Queries *db.Queries
	Logger *slog.Logger
}

func NewService(config ServiceConfig) *Service {
	return &Service{
		db: config.DB,
		queries: config.Queries,
		logger: config.Logger,
	}
}