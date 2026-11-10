package promocodes

import (
	"fmt"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var hundred = decimal.NewFromInt(100)

// ineligibleError is a reason a sender can't use a code. Callers decide the
// HTTP status (404 on preview, 400 on transfer creation).
type ineligibleError string

func (e ineligibleError) Error() string { return string(e) }

const (
	errNotValid         ineligibleError = "promo code is not valid"
	errAlreadyUsed      ineligibleError = "promo code has already been used"
	errNotValidForRoute ineligibleError = "promo code is not valid for this route"
)

// Resolved is a code the sender is allowed to use, with the behaviour that
// applies on the transfer's route.
type Resolved struct {
	Promo      db.PromoCode
	Target     db.PromoDiscountTarget
	Percentage decimal.Decimal
}

// Result is a transfer quote after the promo has been applied.
type Result struct {
	AmountPaid     decimal.Decimal
	AmountReceived decimal.Decimal
	DiscountAmount decimal.Decimal // source currency (SEND_AMOUNT)
	BonusAmount    decimal.Decimal // destination currency (RECIPIENT_AMOUNT)
}

// Apply works out the promo's effect on a quote. The discount value is the
// same either way; only who gets it changes:
//   - SEND_AMOUNT: the sender pays less.
//   - RECIPIENT_AMOUNT: the recipient gets the discount, converted at the rate.
func (r Resolved) Apply(amountPaid, rate, amountReceived decimal.Decimal) Result {
	discount := common.RoundMoney(amountPaid.Mul(r.Percentage).Div(hundred))

	if r.Target == db.PromoDiscountTargetRECIPIENTAMOUNT {
		bonus := common.RoundMoney(discount.Mul(rate))
		return Result{
			AmountPaid:     amountPaid,
			AmountReceived: amountReceived.Add(bonus),
			BonusAmount:    bonus,
		}
	}
	return Result{
		AmountPaid:     amountPaid.Sub(discount),
		AmountReceived: amountReceived,
		DiscountAmount: discount,
	}
}

// ruleForRoute returns how a promo behaves on a route. rule is the
// promo_code_routes row for that route, or nil if there isn't one.
func ruleForRoute(promo db.PromoCode, rule *db.PromoCodeRoute) (db.PromoDiscountTarget, decimal.Decimal, error) {
	if rule == nil && !promo.AppliesToAllRoutes {
		return "", decimal.Zero, errNotValidForRoute
	}

	target, pct := promo.DiscountTarget, promo.DiscountPercentage
	if rule != nil {
		target = rule.DiscountTarget
		if rule.DiscountPercentage.Valid {
			pct = rule.DiscountPercentage
		}
	}

	d, err := common.PgNumericToDecimal(pct)
	if err != nil {
		return "", decimal.Zero, fmt.Errorf("%w", httpx.InternalServerError)
	}
	return target, d, nil
}

// withDefaults fills in optional fields.
func (in PromoCodeInput) withDefaults() PromoCodeInput {
	if in.DiscountTarget == "" {
		in.DiscountTarget = db.PromoDiscountTargetSENDAMOUNT
	}
	if in.AppliesToAllRoutes == nil {
		all := true
		in.AppliesToAllRoutes = &all
	}
	return in
}

// validate checks business rules not covered by struct tags. Call it on the
// result of withDefaults. Route existence is checked in saveRoutes.
func (in PromoCodeInput) validate() error {
	if err := validatePercentage(in.DiscountPercentage, "discount_percentage"); err != nil {
		return err
	}
	if !in.EndDate.After(in.StartDate) {
		return badRequest("end date must be after start date")
	}
	if in.MaxUsesPerUser > in.MaxUses {
		return badRequest("max uses per user cannot be greater than max uses")
	}
	if !*in.AppliesToAllRoutes && len(in.Routes) == 0 {
		return badRequest("at least one route is required when the promo code does not apply to all routes")
	}

	seen := make(map[uuid.UUID]bool, len(in.Routes))
	for _, r := range in.Routes {
		if seen[r.RouteID] {
			return badRequest("each route can only be listed once")
		}
		seen[r.RouteID] = true
		if r.DiscountPercentage != nil {
			if err := validatePercentage(*r.DiscountPercentage, "route discount_percentage"); err != nil {
				return err
			}
		}
	}
	return nil
}

func validatePercentage(pct decimal.Decimal, field string) error {
	if !pct.IsPositive() || pct.GreaterThan(hundred) {
		return badRequest(field + " must be greater than 0 and at most 100")
	}
	return nil
}

func isActive(p db.PromoCode, now time.Time) bool {
	return !now.Before(p.StartDate) && !now.After(p.EndDate)
}

func badRequest(msg string) error {
	return fmt.Errorf("%w: %s", httpx.BadRequestError, msg)
}
