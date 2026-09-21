package promocodes

import (
	"time"

	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreatePromoCodeRequest struct {
	Code               string          `json:"code" validate:"required,min=1,max=64"`
	DiscountPercentage decimal.Decimal `json:"discount_percentage" validate:"required"`
	StartDate          time.Time       `json:"start_date" validate:"required"`
	EndDate            time.Time       `json:"end_date" validate:"required"`
	MaxUses            int             `json:"max_uses" validate:"required,min=1"`
	MaxUsesPerUser     int             `json:"max_uses_per_user" validate:"required,min=1"`
}

type UpdatePromoCodeRequest struct {
	DiscountPercentage decimal.Decimal `json:"discount_percentage" validate:"required"`
	StartDate          time.Time       `json:"start_date" validate:"required"`
	EndDate            time.Time       `json:"end_date" validate:"required"`
	MaxUses            int             `json:"max_uses" validate:"required,min=1"`
	MaxUsesPerUser     int             `json:"max_uses_per_user" validate:"required,min=1"`
}

type GetPromoCodeResponse struct {
	ID                 uuid.UUID       `json:"id"`
	Code               string          `json:"code"`
	DiscountPercentage decimal.Decimal `json:"discount_percentage"`
	StartDate          time.Time       `json:"start_date"`
	EndDate            time.Time       `json:"end_date"`
	MaxUses            int             `json:"max_uses"`
	MaxUsesPerUser     int             `json:"max_uses_per_user"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type PreviewPromoCodeResponse struct {
	Code               string          `json:"code"`
	DiscountPercentage decimal.Decimal `json:"discount_percentage"`
}

func toGetPromoCodeResponse(promoCode db.PromoCode) (GetPromoCodeResponse, error) {
	discount, err := discountFromPromo(promoCode)
	if err != nil {
		return GetPromoCodeResponse{}, err
	}
	return GetPromoCodeResponse{
		ID:                 promoCode.ID,
		Code:               promoCode.Code,
		DiscountPercentage: discount,
		StartDate:          promoCode.StartDate,
		EndDate:            promoCode.EndDate,
		MaxUses:            int(promoCode.MaxUses),
		MaxUsesPerUser:     int(promoCode.MaxUsesPerUser),
		CreatedAt:          promoCode.CreatedAt,
		UpdatedAt:          promoCode.UpdatedAt,
	}, nil
}

func toPreviewPromoCodeResponse(promoCode db.PromoCode) (PreviewPromoCodeResponse, error) {
	discount, err := discountFromPromo(promoCode)
	if err != nil {
		return PreviewPromoCodeResponse{}, err
	}
	return PreviewPromoCodeResponse{
		Code:               promoCode.Code,
		DiscountPercentage: discount,
	}, nil
}
