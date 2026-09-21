package promocodes

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewService(queries *db.Queries, logger *slog.Logger) *Service {
	return &Service{queries: queries, logger: logger}
}

func (s *Service) Create(ctx context.Context, data CreatePromoCodeRequest) (GetPromoCodeResponse, error) {
	if err := validatePromoCodeInput(data.DiscountPercentage, data.StartDate, data.EndDate, data.MaxUses, data.MaxUsesPerUser); err != nil {
		return GetPromoCodeResponse{}, err
	}

	discountPercentage, err := common.DecimalToPgNumeric(data.DiscountPercentage)
	if err != nil {
		s.logger.Error("failed to convert discount percentage to pg numeric", "error", err)
		return GetPromoCodeResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	promoCode, err := s.queries.CreatePromoCode(ctx, db.CreatePromoCodeParams{
		Code:               data.Code,
		DiscountPercentage: discountPercentage,
		StartDate:          data.StartDate,
		EndDate:            data.EndDate,
		MaxUses:            int32(data.MaxUses),
		MaxUsesPerUser:     int32(data.MaxUsesPerUser),
	})
	if err != nil {
		s.logger.Error("failed to create promo code", "error", err)
		return GetPromoCodeResponse{}, common.TranslateDBError(err)
	}

	return toGetPromoCodeResponse(promoCode)
}

func (s *Service) GetByCode(ctx context.Context, code string) (PreviewPromoCodeResponse, error) {
	promoCode, err := s.queries.GetPromoCodeByCode(ctx, code)
	if err != nil {
		return PreviewPromoCodeResponse{}, common.TranslateDBError(err)
	}

	now := time.Now().UTC()
	if now.Before(promoCode.StartDate) || now.After(promoCode.EndDate) {
		return PreviewPromoCodeResponse{}, fmt.Errorf("%w", httpx.NotFoundError)
	}

	used, err := s.queries.CountRedemptionsByPromoCodeID(ctx, promoCode.ID)
	if err != nil {
		s.logger.Error("failed to count promo code redemptions", "error", err, "promo_code_id", promoCode.ID)
		return PreviewPromoCodeResponse{}, common.TranslateDBError(err)
	}
	if used >= int64(promoCode.MaxUses) {
		return PreviewPromoCodeResponse{}, fmt.Errorf("%w", httpx.NotFoundError)
	}

	return toPreviewPromoCodeResponse(promoCode)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (GetPromoCodeResponse, error) {
	promoCode, err := s.queries.GetPromoCodeByID(ctx, id)
	if err != nil {
		return GetPromoCodeResponse{}, common.TranslateDBError(err)
	}
	return toGetPromoCodeResponse(promoCode)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.queries.DeletePromoCode(ctx, id); err != nil {
		return common.TranslateDBError(err)
	}
	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]GetPromoCodeResponse, error) {
	promoCodes, err := s.queries.GetAllPromoCodes(ctx)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}

	responses := make([]GetPromoCodeResponse, 0, len(promoCodes))
	for _, promoCode := range promoCodes {
		response, err := toGetPromoCodeResponse(promoCode)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, data UpdatePromoCodeRequest) (GetPromoCodeResponse, error) {
	if err := validatePromoCodeInput(data.DiscountPercentage, data.StartDate, data.EndDate, data.MaxUses, data.MaxUsesPerUser); err != nil {
		return GetPromoCodeResponse{}, err
	}

	discountPercentage, err := common.DecimalToPgNumeric(data.DiscountPercentage)
	if err != nil {
		s.logger.Error("failed to convert discount percentage to pg numeric", "error", err)
		return GetPromoCodeResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	promoCode, err := s.queries.UpdatePromoCode(ctx, db.UpdatePromoCodeParams{
		ID:                 id,
		DiscountPercentage: discountPercentage,
		StartDate:          data.StartDate,
		EndDate:            data.EndDate,
		MaxUses:            int32(data.MaxUses),
		MaxUsesPerUser:     int32(data.MaxUsesPerUser),
	})
	if err != nil {
		return GetPromoCodeResponse{}, common.TranslateDBError(err)
	}

	return toGetPromoCodeResponse(promoCode)
}

func validatePromoCodeInput(discount decimal.Decimal, startDate, endDate time.Time, maxUses, maxUsesPerUser int) error {
	if discount.IsNegative() || discount.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("%w: discount percentage must be between 0 and 100", httpx.BadRequestError)
	}
	if !endDate.After(startDate) {
		return fmt.Errorf("%w: end date must be after start date", httpx.BadRequestError)
	}
	if maxUses < 1 {
		return fmt.Errorf("%w: max uses must be greater than 0", httpx.BadRequestError)
	}
	if maxUsesPerUser < 1 || maxUsesPerUser > maxUses {
		return fmt.Errorf("%w: max uses per user must be between 1 and max uses", httpx.BadRequestError)
	}
	return nil
}

func discountFromPromo(promoCode db.PromoCode) (decimal.Decimal, error) {
	discount, err := common.PgNumericToDecimal(promoCode.DiscountPercentage)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w", httpx.InternalServerError)
	}
	return discount, nil
}
