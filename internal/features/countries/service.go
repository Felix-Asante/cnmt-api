package countries

import (
	"context"
	"log/slog"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger
	db      *pgxpool.Pool
}

func NewService(queries *db.Queries, logger *slog.Logger, db *pgxpool.Pool) *Service {
	return &Service{queries: queries, logger: logger, db: db}
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
	if len(destCountries) == 0 {
		return []DestCountryResponse{}, nil
	}

	countryIDs := make([]int64, len(destCountries))
	for i, dest := range destCountries {
		countryIDs[i] = dest.ID
	}

	channelsByCountry, err := s.loadChannelsByCountry(ctx, countryIDs)
	if err != nil {
		return nil, err
	}

	responses, err := mapDestCountriesToResponses(destCountries, channelsByCountry)
	if err != nil {
		s.logger.Error("failed to map destination countries", "error", err, "source_country_id", srcCountryID)
		return nil, err
	}
	return responses, nil
}

func (s *Service) loadChannelsByCountry(ctx context.Context, countryIDs []int64) (map[int64][]db.GetActivePaymentChannelsByCountryIDsRow, error) {
	channelsByCountry := make(map[int64][]db.GetActivePaymentChannelsByCountryIDsRow, len(countryIDs))
	if len(countryIDs) == 0 {
		return channelsByCountry, nil
	}

	channels, err := s.queries.GetActivePaymentChannelsByCountryIDs(ctx, countryIDs)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}
	for _, ch := range channels {
		channelsByCountry[ch.CountryID] = append(channelsByCountry[ch.CountryID], ch)
	}
	return channelsByCountry, nil
}

func (s *Service) CreateCountry(ctx context.Context, req CreateCountryRequest) (CountryResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CountryResponse{}, common.TranslateDBError(err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	country, err := qtx.CreateCountry(ctx, db.CreateCountryParams{
		Name:           req.Name,
		IsoCode:        req.ISOCode,
		Flag:           req.Flag,
		CurrencyName:   req.CurrencyName,
		CurrencyCode:   req.CurrencyCode,
		CurrencySymbol: req.CurrencySymbol,
	})
	if err != nil {
		return CountryResponse{}, common.TranslateDBError(err)
	}
	for _, channel := range req.PaymentChannels {
		_, err = qtx.CreatePaymentChannel(ctx, db.CreatePaymentChannelParams{
			Name:        channel.Name,
			ChannelType: channel.ChannelType,
			CountryID:   country.ID,
		})
		if err != nil {
			return CountryResponse{}, common.TranslateDBError(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return CountryResponse{}, common.TranslateDBError(err)
	}
	return mapCountryToResponse(country), nil
}
