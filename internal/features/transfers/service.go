package transfers

import (
	"cnmt/internal/infra/db"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
	queries *db.Queries
}

func NewService(db *pgxpool.Pool, queries *db.Queries) *Service {
	return &Service{db: db, queries: queries}
}

func (s *Service) CreateTransfer(ctx context.Context, body createTransferRequest) (createTransferResponse, error) {
	// validate route
	// validate amount paid
	// validate amount received

	// transfer, err := s.queries.CreateTransfer(ctx)
	return createTransferResponse{
		TransferID: "1234567890",
		Reference: "1234567890",
		ExpiresIn: 1000,
	}, nil
}