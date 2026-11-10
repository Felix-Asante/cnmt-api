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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db      *pgxpool.Pool
	queries *db.Queries
	logger  *slog.Logger
}

func NewService(pool *pgxpool.Pool, queries *db.Queries, logger *slog.Logger) *Service {
	return &Service{db: pool, queries: queries, logger: logger}
}

// ---- Admin ----

func (s *Service) Create(ctx context.Context, data CreatePromoCodeRequest) (GetPromoCodeResponse, error) {
	in := data.PromoCodeInput.withDefaults()
	if err := in.validate(); err != nil {
		return GetPromoCodeResponse{}, err
	}
	pct, err := common.DecimalToPgNumeric(in.DiscountPercentage)
	if err != nil {
		return GetPromoCodeResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	var resp GetPromoCodeResponse
	err = s.inTx(ctx, func(q *db.Queries) error {
		promo, err := q.CreatePromoCode(ctx, db.CreatePromoCodeParams{
			Code:               data.Code,
			DiscountPercentage: pct,
			StartDate:          in.StartDate,
			EndDate:            in.EndDate,
			MaxUses:            int32(in.MaxUses),
			MaxUsesPerUser:     int32(in.MaxUsesPerUser),
			AppliesToAllRoutes: *in.AppliesToAllRoutes,
			DiscountTarget:     in.DiscountTarget,
		})
		if err != nil {
			return common.TranslateDBError(err)
		}
		if err := s.saveRoutes(ctx, q, promo.ID, in.Routes); err != nil {
			return err
		}
		resp, err = s.toResponse(ctx, q, promo)
		return err
	})
	return resp, err
}

// Update replaces the promo's settings and route rules.
func (s *Service) Update(ctx context.Context, id uuid.UUID, data UpdatePromoCodeRequest) (GetPromoCodeResponse, error) {
	in := data.withDefaults()
	if err := in.validate(); err != nil {
		return GetPromoCodeResponse{}, err
	}
	pct, err := common.DecimalToPgNumeric(in.DiscountPercentage)
	if err != nil {
		return GetPromoCodeResponse{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	var resp GetPromoCodeResponse
	err = s.inTx(ctx, func(q *db.Queries) error {
		// Lock the row so a redemption can't land between the check and the update.
		if _, err := q.GetPromoCodeByIDForUpdate(ctx, id); err != nil {
			return common.TranslateDBError(err)
		}
		used, err := q.CountRedemptionsByPromoCodeID(ctx, id)
		if err != nil {
			return common.TranslateDBError(err)
		}
		if int64(in.MaxUses) < used {
			return badRequest(fmt.Sprintf("max uses cannot be lower than current usage (%d)", used))
		}

		promo, err := q.UpdatePromoCode(ctx, db.UpdatePromoCodeParams{
			ID:                 id,
			DiscountPercentage: pct,
			StartDate:          in.StartDate,
			EndDate:            in.EndDate,
			MaxUses:            int32(in.MaxUses),
			MaxUsesPerUser:     int32(in.MaxUsesPerUser),
			AppliesToAllRoutes: *in.AppliesToAllRoutes,
			DiscountTarget:     in.DiscountTarget,
		})
		if err != nil {
			return common.TranslateDBError(err)
		}
		if err := q.DeletePromoCodeRoutes(ctx, id); err != nil {
			return common.TranslateDBError(err)
		}
		if err := s.saveRoutes(ctx, q, id, in.Routes); err != nil {
			return err
		}
		resp, err = s.toResponse(ctx, q, promo)
		return err
	})
	return resp, err
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (GetPromoCodeResponse, error) {
	promo, err := s.queries.GetPromoCodeByID(ctx, id)
	if err != nil {
		return GetPromoCodeResponse{}, common.TranslateDBError(err)
	}
	return s.toResponse(ctx, s.queries, promo)
}

func (s *Service) GetAll(ctx context.Context) ([]GetPromoCodeResponse, error) {
	promos, err := s.queries.GetAllPromoCodes(ctx)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}

	ids := make([]uuid.UUID, len(promos))
	for i, p := range promos {
		ids[i] = p.ID
	}
	routes, err := s.queries.ListPromoCodeRoutes(ctx, ids)
	if err != nil {
		return nil, common.TranslateDBError(err)
	}
	routesByPromo := make(map[uuid.UUID][]db.ListPromoCodeRoutesRow)
	for _, r := range routes {
		routesByPromo[r.PromoCodeID] = append(routesByPromo[r.PromoCodeID], r)
	}

	responses := make([]GetPromoCodeResponse, 0, len(promos))
	for _, p := range promos {
		resp, err := toGetPromoCodeResponse(p, routesByPromo[p.ID])
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.queries.DeletePromoCode(ctx, id); err != nil {
		return common.TranslateDBError(err)
	}
	return nil
}

// ---- Public / transfers ----

// Preview checks a code before a transfer is created. Route and sender are
// optional; when given, the same checks as transfer creation are applied.
func (s *Service) Preview(ctx context.Context, code string, req PreviewPromoCodeRequest) (PreviewPromoCodeResponse, error) {
	promo, err := s.queries.GetPromoCodeByCode(ctx, code)
	if err != nil {
		return PreviewPromoCodeResponse{}, common.TranslateDBError(err)
	}

	var routeID *uuid.UUID
	if req.SourceCountryID > 0 && req.DestinationCountryID > 0 {
		route, err := s.queries.GetActiveRouteByCountries(ctx, db.GetActiveRouteByCountriesParams{
			SourceCountryID:      req.SourceCountryID,
			DestinationCountryID: req.DestinationCountryID,
		})
		if err != nil {
			return PreviewPromoCodeResponse{}, common.TranslateDBError(err)
		}
		routeID = &route.ID
	}

	sender := req.SenderPhone
	if sender != "" {
		if sender, err = common.NormalizePhone(sender); err != nil {
			return PreviewPromoCodeResponse{}, badRequest("sender phone is not valid")
		}
	}

	resolved, err := s.resolve(ctx, s.queries, promo, sender, routeID)
	if err != nil {
		return PreviewPromoCodeResponse{}, asHTTPError(err, httpx.NotFoundError)
	}
	return PreviewPromoCodeResponse{
		Code:               promo.Code,
		DiscountPercentage: resolved.Percentage,
		DiscountTarget:     resolved.Target,
		AppliesToAllRoutes: promo.AppliesToAllRoutes,
	}, nil
}

// Resolve locks the code and checks the sender can use it on the route.
// q must be bound to the transfer's transaction.
func (s *Service) Resolve(ctx context.Context, q *db.Queries, code, senderPhone string, routeID uuid.UUID) (Resolved, error) {
	promo, err := q.GetPromoCodeByCodeForUpdate(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return Resolved{}, badRequest(errNotValid.Error())
	}
	if err != nil {
		return Resolved{}, common.TranslateDBError(err)
	}

	resolved, err := s.resolve(ctx, q, promo, senderPhone, &routeID)
	if err != nil {
		return Resolved{}, asHTTPError(err, httpx.BadRequestError)
	}
	return resolved, nil
}

// Redeem records that a transfer used the promo, and what it did.
func (s *Service) Redeem(ctx context.Context, q *db.Queries, resolved Resolved, result Result, transferID uuid.UUID, senderPhone string) error {
	pct, err1 := common.DecimalToPgNumeric(resolved.Percentage)
	discount, err2 := common.DecimalToPgNumeric(result.DiscountAmount)
	bonus, err3 := common.DecimalToPgNumeric(result.BonusAmount)
	if err := errors.Join(err1, err2, err3); err != nil {
		return fmt.Errorf("%w", httpx.InternalServerError)
	}

	err := q.RedeemPromoCode(ctx, db.RedeemPromoCodeParams{
		PromoCodeID:        resolved.Promo.ID,
		TransferID:         transferID,
		SenderPhone:        senderPhone,
		DiscountPercentage: pct,
		DiscountTarget:     resolved.Target,
		DiscountAmount:     discount,
		BonusAmount:        bonus,
	})
	if err != nil {
		s.logger.Error("failed to redeem promo code", "error", err, "promo_code_id", resolved.Promo.ID, "transfer_id", transferID)
		return common.TranslateDBError(err)
	}
	return nil
}

// resolve runs the eligibility checks. senderPhone and routeID are optional.
// Returns an ineligibleError when the code can't be used.
func (s *Service) resolve(ctx context.Context, q *db.Queries, promo db.PromoCode, senderPhone string, routeID *uuid.UUID) (Resolved, error) {
	if !isActive(promo, time.Now().UTC()) {
		return Resolved{}, errNotValid
	}

	used, err := q.CountRedemptionsByPromoCodeID(ctx, promo.ID)
	if err != nil {
		return Resolved{}, common.TranslateDBError(err)
	}
	if used >= int64(promo.MaxUses) {
		return Resolved{}, errNotValid
	}

	if senderPhone != "" {
		usedBySender, err := q.CountRedemptionsByPromoCodeAndSender(ctx, db.CountRedemptionsByPromoCodeAndSenderParams{
			PromoCodeID: promo.ID,
			SenderPhone: senderPhone,
		})
		if err != nil {
			return Resolved{}, common.TranslateDBError(err)
		}
		if usedBySender >= int64(promo.MaxUsesPerUser) {
			return Resolved{}, errAlreadyUsed
		}
	}

	if routeID == nil {
		// No route to check against: describe the code's default behaviour.
		pct, err := common.PgNumericToDecimal(promo.DiscountPercentage)
		if err != nil {
			return Resolved{}, fmt.Errorf("%w", httpx.InternalServerError)
		}
		return Resolved{Promo: promo, Target: promo.DiscountTarget, Percentage: pct}, nil
	}

	var rule *db.PromoCodeRoute
	r, err := q.GetPromoCodeRouteRule(ctx, db.GetPromoCodeRouteRuleParams{PromoCodeID: promo.ID, RouteID: *routeID})
	switch {
	case err == nil:
		rule = &r
	case !errors.Is(err, pgx.ErrNoRows):
		return Resolved{}, common.TranslateDBError(err)
	}

	target, pct, err := ruleForRoute(promo, rule)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{Promo: promo, Target: target, Percentage: pct}, nil
}

// ---- Helpers ----

func (s *Service) saveRoutes(ctx context.Context, q *db.Queries, promoID uuid.UUID, routes []RouteRule) error {
	if len(routes) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, len(routes))
	for i, r := range routes {
		ids[i] = r.RouteID
	}
	alive, err := q.CountAliveRoutesByIDs(ctx, ids)
	if err != nil {
		return common.TranslateDBError(err)
	}
	if alive != int64(len(ids)) {
		return badRequest("one or more routes do not exist")
	}

	for _, r := range routes {
		var pct pgtype.Numeric
		if r.DiscountPercentage != nil {
			if pct, err = common.DecimalToPgNumeric(*r.DiscountPercentage); err != nil {
				return fmt.Errorf("%w", httpx.InternalServerError)
			}
		}
		if err := q.InsertPromoCodeRoute(ctx, db.InsertPromoCodeRouteParams{
			PromoCodeID:        promoID,
			RouteID:            r.RouteID,
			DiscountTarget:     r.DiscountTarget,
			DiscountPercentage: pct,
		}); err != nil {
			return common.TranslateDBError(err)
		}
	}
	return nil
}

func (s *Service) toResponse(ctx context.Context, q *db.Queries, promo db.PromoCode) (GetPromoCodeResponse, error) {
	routes, err := q.ListPromoCodeRoutes(ctx, []uuid.UUID{promo.ID})
	if err != nil {
		return GetPromoCodeResponse{}, common.TranslateDBError(err)
	}
	return toGetPromoCodeResponse(promo, routes)
}

func (s *Service) inTx(ctx context.Context, fn func(q *db.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return common.TranslateDBError(err)
	}
	defer tx.Rollback(ctx)

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}
	return common.TranslateDBError(tx.Commit(ctx))
}

// asHTTPError maps an ineligibleError to the given HTTP sentinel; other errors
// pass through unchanged.
func asHTTPError(err error, sentinel error) error {
	var ie ineligibleError
	if errors.As(err, &ie) {
		return fmt.Errorf("%w: %s", sentinel, ie)
	}
	return err
}
