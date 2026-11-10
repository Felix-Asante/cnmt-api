package promocodes

import (
	"time"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RouteRule sets how a promo behaves on one route.
type RouteRule struct {
	RouteID        uuid.UUID              `json:"route_id" validate:"required"`
	DiscountTarget db.PromoDiscountTarget `json:"discount_target" validate:"required,oneof=SEND_AMOUNT RECIPIENT_AMOUNT"`
	// Optional; when nil the code's discount_percentage is used.
	DiscountPercentage *decimal.Decimal `json:"discount_percentage,omitempty"`
}

// PromoCodeInput is the editable part of a promo code. Updates replace all of
// it, including the route rules.
type PromoCodeInput struct {
	DiscountPercentage decimal.Decimal `json:"discount_percentage" validate:"required"`
	StartDate          time.Time       `json:"start_date" validate:"required"`
	EndDate            time.Time       `json:"end_date" validate:"required"`
	MaxUses            int             `json:"max_uses" validate:"required,min=1"`
	MaxUsesPerUser     int             `json:"max_uses_per_user" validate:"required,min=1"`
	// Defaults to SEND_AMOUNT.
	DiscountTarget db.PromoDiscountTarget `json:"discount_target" validate:"omitempty,oneof=SEND_AMOUNT RECIPIENT_AMOUNT"`
	// Defaults to true.
	AppliesToAllRoutes *bool       `json:"applies_to_all_routes"`
	Routes             []RouteRule `json:"routes" validate:"omitempty,max=500,dive"`
}

type CreatePromoCodeRequest struct {
	Code string `json:"code" validate:"required,min=1,max=64"`
	PromoCodeInput
}

type UpdatePromoCodeRequest = PromoCodeInput

type RouteRuleResponse struct {
	RouteRule
	SourceCountryName      string `json:"source_country_name"`
	DestinationCountryName string `json:"destination_country_name"`
}

type GetPromoCodeResponse struct {
	ID                 uuid.UUID              `json:"id"`
	Code               string                 `json:"code"`
	DiscountPercentage decimal.Decimal        `json:"discount_percentage"`
	DiscountTarget     db.PromoDiscountTarget `json:"discount_target"`
	AppliesToAllRoutes bool                   `json:"applies_to_all_routes"`
	Routes             []RouteRuleResponse    `json:"routes"`
	StartDate          time.Time              `json:"start_date"`
	EndDate            time.Time              `json:"end_date"`
	MaxUses            int                    `json:"max_uses"`
	MaxUsesPerUser     int                    `json:"max_uses_per_user"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// PreviewPromoCodeRequest comes from query params. Every field is optional:
// with a route the response shows that route's behaviour, and with a sender
// the per-user limit is checked too.
type PreviewPromoCodeRequest struct {
	SourceCountryID      int64
	DestinationCountryID int64
	SenderPhone          string
}

type PreviewPromoCodeResponse struct {
	Code               string                 `json:"code"`
	DiscountPercentage decimal.Decimal        `json:"discount_percentage"`
	DiscountTarget     db.PromoDiscountTarget `json:"discount_target"`
	AppliesToAllRoutes bool                   `json:"applies_to_all_routes"`
}

func toGetPromoCodeResponse(p db.PromoCode, routes []db.ListPromoCodeRoutesRow) (GetPromoCodeResponse, error) {
	pct, err := common.PgNumericToDecimal(p.DiscountPercentage)
	if err != nil {
		return GetPromoCodeResponse{}, err
	}

	rules := make([]RouteRuleResponse, 0, len(routes))
	for _, r := range routes {
		var override *decimal.Decimal
		if r.DiscountPercentage.Valid {
			d, err := common.PgNumericToDecimal(r.DiscountPercentage)
			if err != nil {
				return GetPromoCodeResponse{}, err
			}
			override = &d
		}
		rules = append(rules, RouteRuleResponse{
			RouteRule: RouteRule{
				RouteID:            r.RouteID,
				DiscountTarget:     r.DiscountTarget,
				DiscountPercentage: override,
			},
			SourceCountryName:      r.SourceCountryName,
			DestinationCountryName: r.DestinationCountryName,
		})
	}

	return GetPromoCodeResponse{
		ID:                 p.ID,
		Code:               p.Code,
		DiscountPercentage: pct,
		DiscountTarget:     p.DiscountTarget,
		AppliesToAllRoutes: p.AppliesToAllRoutes,
		Routes:             rules,
		StartDate:          p.StartDate,
		EndDate:            p.EndDate,
		MaxUses:            int(p.MaxUses),
		MaxUsesPerUser:     int(p.MaxUsesPerUser),
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}, nil
}
