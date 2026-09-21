package promocodes

import (
	"context"
	"errors"
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


func (s *Service) ApplyToFee(ctx context.Context, q *db.Queries, code, senderPhone string, fee decimal.Decimal) (db.PromoCode, decimal.Decimal, error) {
	promo, err := q.GetPromoCodeByCodeForUpdate(ctx, code)
	if err != nil {
		translated := common.TranslateDBError(err)
		if errors.Is(translated, httpx.NotFoundError) {
			return db.PromoCode{}, decimal.Zero, fmt.Errorf("%w: promo code is not valid", httpx.BadRequestError)
		}
		return db.PromoCode{}, decimal.Zero, translated
	}

	now := time.Now().UTC()
	if now.Before(promo.StartDate) || now.After(promo.EndDate) {
		return db.PromoCode{}, decimal.Zero, fmt.Errorf("%w: promo code is not valid", httpx.BadRequestError)
	}

	used, err := q.CountRedemptionsByPromoCodeID(ctx, promo.ID)
	if err != nil {
		s.logger.Error("failed to count promo code redemptions", "error", err, "promo_code_id", promo.ID)
		return db.PromoCode{}, decimal.Zero, common.TranslateDBError(err)
	}
	if used >= int64(promo.MaxUses) {
		return db.PromoCode{}, decimal.Zero, fmt.Errorf("%w: promo code is not valid", httpx.BadRequestError)
	}

	usedBySender, err := q.CountRedemptionsByPromoCodeAndSender(ctx, db.CountRedemptionsByPromoCodeAndSenderParams{
		PromoCodeID: promo.ID,
		SenderPhone: senderPhone,
	})
	if err != nil {
		s.logger.Error("failed to count promo code redemptions for sender", "error", err, "promo_code_id", promo.ID)
		return db.PromoCode{}, decimal.Zero, common.TranslateDBError(err)
	}
	if usedBySender >= int64(promo.MaxUsesPerUser) {
		return db.PromoCode{}, decimal.Zero, fmt.Errorf("%w: promo code has already been used", httpx.BadRequestError)
	}

	pct, err := common.PgNumericToDecimal(promo.DiscountPercentage)
	if err != nil {
		s.logger.Error("failed to convert promo discount percentage", "error", err, "promo_code_id", promo.ID)
		return db.PromoCode{}, decimal.Zero, fmt.Errorf("%w", httpx.InternalServerError)
	}

	discount := common.RoundMoney(fee.Mul(pct).Div(decimal.NewFromInt(100)))
	discounted := common.RoundMoney(fee.Sub(discount))
	if discounted.IsNegative() {
		discounted = decimal.Zero
	}

	return promo, discounted, nil
}


func (s *Service) RedeemForTransfer(ctx context.Context, q *db.Queries, promo db.PromoCode, transferID uuid.UUID, senderPhone string) error {
	_, err := q.RedeemPromoCode(ctx, db.RedeemPromoCodeParams{
		PromoCodeID:        promo.ID,
		TransferID:         transferID,
		SenderPhone:        senderPhone,
		DiscountPercentage: promo.DiscountPercentage,
	})
	if err != nil {
		s.logger.Error("failed to redeem promo code", "error", err, "promo_code_id", promo.ID, "transfer_id", transferID)
		return common.TranslateDBError(err)
	}
	return nil
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
