package countries

import (
	"cnmt/internal/common"
	"cnmt/internal/infra/db"
	"context"
	"log/slog"
)

type Service struct {
	queries *db.Queries
	logger *slog.Logger
}

func NewService(queries *db.Queries, logger *slog.Logger) *Service {
	return &Service{queries: queries, logger: logger}
}

func (s *Service) GetCountries(ctx context.Context) ([]CountryResponse, error) {
	countries, err := s.queries.GetAllCountries(ctx)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}
	return mapCountriesToResponses(countries), nil
}