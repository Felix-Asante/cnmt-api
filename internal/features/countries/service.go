package countries

import (
	"context"
	"log/slog"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger
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

func (s *Service) GetDestCountriesBySrcCountryID(ctx context.Context, srcCountryID int64) ([]DestCountryResponse, error) {
	destCountries, err := s.queries.GetDestCountriesBySrcCountryID(ctx, srcCountryID)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}

	responses, err := mapDestCountriesToResponses(destCountries)
	if err != nil {
		s.logger.Error("failed to map destination countries", "error", err, "source_country_id", srcCountryID)
		return nil, err
	}
	return responses, nil
}
