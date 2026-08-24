package countries

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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

func (s *Service) GetCountryByID(ctx context.Context, id int64) (CountryDetailResponse, error) {
	country, err := s.queries.GetAdminCountryByID(ctx, id)
	if err != nil {
		return CountryDetailResponse{}, common.TranslateDBError(err)
	}

	channels, err := s.queries.GetPaymentChannelsByCountryID(ctx, id)
	if err != nil {
		return CountryDetailResponse{}, common.TranslateDBError(err)
	}

	paymentChannels := make([]PaymentChannelResponse, 0, len(channels))
	for _, ch := range channels {
		paymentChannels = append(paymentChannels, mapPaymentChannelToResponse(ch))
	}

	return CountryDetailResponse{
		CountryResponse: mapCountryToResponse(country),
		PaymentChannels: paymentChannels,
	}, nil
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
	country, err := qtx.CreateCountry(ctx, req.toCreateParams())
	if err != nil {
		return CountryResponse{}, common.TranslateDBError(err)
	}
	for _, channel := range req.PaymentChannels {
		_, err = qtx.CreatePaymentChannel(ctx, channel.toCreateParams(country.ID))
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

func (s *Service) UpdateCountry(ctx context.Context, id int64, req UpdateCountryRequest) (CountryResponse, error) {
	country, err := s.queries.UpdateCountry(ctx, req.toUpdateParams(id))
	if err != nil {
		return CountryResponse{}, common.TranslateDBError(err)
	}
	return mapCountryToResponse(country), nil
}

func (s *Service) DeleteCountry(ctx context.Context, id int64) error {
	if _, err := s.queries.DeleteCountry(ctx, id); err != nil {
		return common.TranslateDBError(err)
	}
	return nil
}

func (s *Service) UpdatePaymentChannel(ctx context.Context, id uuid.UUID, req UpdatePaymentChannelRequest) (PaymentChannelResponse, error) {
	channel, err := s.queries.UpdatePaymentChannel(ctx, req.toUpdateParams(id))
	if err != nil {
		return PaymentChannelResponse{}, common.TranslateDBError(err)
	}
	return mapPaymentChannelToResponse(channel), nil
}

func (s *Service) DeletePaymentChannel(ctx context.Context, id uuid.UUID) error {
	if _, err := s.queries.DeletePaymentChannel(ctx, id); err != nil {
		return common.TranslateDBError(err)
	}
	return nil
}

func (s *Service) CreateRoute(ctx context.Context, req CreateRouteRequest) (RouteResponse, error) {
	if req.SourceCountryID == req.DestCountryID {
		return RouteResponse{}, fmt.Errorf("%w: source and destination country cannot be the same", httpx.BadRequestError)
	}
	if err := validateRoutePricing(req.FeeType, req.Fee, req.ExchangeRate, req.MinTransferAmount, req.MaxTransferAmount); err != nil {
		return RouteResponse{}, err
	}
	if err := s.requireActiveCountry(ctx, req.SourceCountryID, "source"); err != nil {
		return RouteResponse{}, err
	}
	if err := s.requireActiveCountry(ctx, req.DestCountryID, "destination"); err != nil {
		return RouteResponse{}, err
	}

	pricing, err := s.convertRoutePricing(req.ExchangeRate, req.Fee, req.MinTransferAmount, req.MaxTransferAmount)
	if err != nil {
		return RouteResponse{}, err
	}

	route, err := s.queries.CreateRoute(ctx, db.CreateRouteParams{
		SourceCountryID:      req.SourceCountryID,
		DestinationCountryID: req.DestCountryID,
		DefaultExchangeRate:  pricing.rate,
		Fee:                  pricing.fee,
		FeeType:              req.FeeType,
		MinTransferAmount:    pricing.minAmount,
		MaxTransferAmount:    pricing.maxAmount,
	})
	if err != nil {
		return RouteResponse{}, common.TranslateDBError(err)
	}
	return s.mapRoute(route)
}

func (s *Service) UpdateRoute(ctx context.Context, id uuid.UUID, req UpdateRouteRequest) (RouteResponse, error) {
	if err := validateRoutePricing(req.FeeType, req.Fee, req.ExchangeRate, req.MinTransferAmount, req.MaxTransferAmount); err != nil {
		return RouteResponse{}, err
	}

	pricing, err := s.convertRoutePricing(req.ExchangeRate, req.Fee, req.MinTransferAmount, req.MaxTransferAmount)
	if err != nil {
		return RouteResponse{}, err
	}

	route, err := s.queries.UpdateRoute(ctx, db.UpdateRouteParams{
		ID:                  id,
		DefaultExchangeRate: pricing.rate,
		Fee:                 pricing.fee,
		FeeType:             req.FeeType,
		MinTransferAmount:   pricing.minAmount,
		MaxTransferAmount:   pricing.maxAmount,
	})
	if err != nil {
		return RouteResponse{}, common.TranslateDBError(err)
	}
	return s.mapRoute(route)
}

func (s *Service) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	if _, err := s.queries.DeleteRoute(ctx, id); err != nil {
		return common.TranslateDBError(err)
	}
	return nil
}

func (s *Service) ToggleRouteActive(ctx context.Context, id uuid.UUID) (RouteResponse, error) {
	route, err := s.queries.ToggleRouteActive(ctx, id)
	if err != nil {
		return RouteResponse{}, common.TranslateDBError(err)
	}
	return s.mapRoute(route)
}

func (s *Service) ListRoutes(ctx context.Context, sourceCountryID, destCountryID int64, isActive string) ([]RouteResponse, error) {
	rows, err := s.queries.ListRoutes(ctx, db.ListRoutesParams{
		SourceCountryID:      sourceCountryID,
		DestinationCountryID: destCountryID,
		IsActive:             isActive,
	})
	if err != nil {
		return nil, common.TranslateDBError(err)
	}

	routes := make([]RouteResponse, 0, len(rows))
	for _, row := range rows {
		route, err := s.mapRoute(row)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}
	return routes, nil
}

type routePricing struct {
	rate      pgtype.Numeric
	fee       pgtype.Numeric
	minAmount pgtype.Numeric
	maxAmount pgtype.Numeric
}

func validateRoutePricing(feeType db.FeeType, fee, rate, minAmount, maxAmount decimal.Decimal) error {
	if rate.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: exchange rate must be greater than 0", httpx.BadRequestError)
	}
	if fee.LessThan(decimal.Zero) {
		return fmt.Errorf("%w: fee cannot be negative", httpx.BadRequestError)
	}
	if feeType == db.FeeTypePercentage && fee.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("%w: fee cannot be greater than 100", httpx.BadRequestError)
	}
	if minAmount.LessThan(decimal.Zero) || maxAmount.LessThan(decimal.Zero) {
		return fmt.Errorf("%w: transfer amounts cannot be negative", httpx.BadRequestError)
	}
	if minAmount.GreaterThan(maxAmount) {
		return fmt.Errorf("%w: min transfer amount cannot be greater than max transfer amount", httpx.BadRequestError)
	}
	return nil
}

func (s *Service) convertRoutePricing(rate, fee, minAmount, maxAmount decimal.Decimal) (routePricing, error) {
	numericRate, err := common.DecimalToPgRate(rate)
	if err != nil {
		s.logger.Error("failed to convert exchange rate", "error", err)
		return routePricing{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	numericFee, err := common.DecimalToPgNumeric(fee)
	if err != nil {
		s.logger.Error("failed to convert fee", "error", err)
		return routePricing{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	numericMin, err := common.DecimalToPgNumeric(minAmount)
	if err != nil {
		s.logger.Error("failed to convert min transfer amount", "error", err)
		return routePricing{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	numericMax, err := common.DecimalToPgNumeric(maxAmount)
	if err != nil {
		s.logger.Error("failed to convert max transfer amount", "error", err)
		return routePricing{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	return routePricing{
		rate:      numericRate,
		fee:       numericFee,
		minAmount: numericMin,
		maxAmount: numericMax,
	}, nil
}

func (s *Service) mapRoute(route db.Route) (RouteResponse, error) {
	resp, err := mapRouteToResponse(route)
	if err != nil {
		s.logger.Error("failed to map route", "error", err, "route_id", route.ID)
		return RouteResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}
	return resp, nil
}

func (s *Service) requireActiveCountry(ctx context.Context, countryID int64, label string) error {
	if _, err := s.queries.GetCountryByID(ctx, countryID); err != nil {
		translated := common.TranslateDBError(err)
		if errors.Is(translated, httpx.NotFoundError) {
			return fmt.Errorf("%w: %s country not found", httpx.NotFoundError, label)
		}
		return translated
	}
	return nil
}
