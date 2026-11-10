package promocodes

import (
	"errors"
	"testing"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func num(t *testing.T, s string) pgtype.Numeric {
	t.Helper()
	n, err := common.DecimalToPgNumeric(dec(s))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func promo(t *testing.T, pct string, allRoutes bool, target db.PromoDiscountTarget) db.PromoCode {
	return db.PromoCode{DiscountPercentage: num(t, pct), AppliesToAllRoutes: allRoutes, DiscountTarget: target}
}

func TestRuleForRoute(t *testing.T) {
	send, recv := db.PromoDiscountTargetSENDAMOUNT, db.PromoDiscountTargetRECIPIENTAMOUNT

	tests := []struct {
		name       string
		promo      db.PromoCode
		rule       *db.PromoCodeRoute
		wantTarget db.PromoDiscountTarget
		wantPct    string
		wantErr    error
	}{
		{
			name:       "all routes, no rule -> code default",
			promo:      promo(t, "10", true, send),
			wantTarget: send, wantPct: "10",
		},
		{
			name:       "all routes, recipient default",
			promo:      promo(t, "10", true, recv),
			wantTarget: recv, wantPct: "10",
		},
		{
			name:       "all routes, route override switches target",
			promo:      promo(t, "10", true, send),
			rule:       &db.PromoCodeRoute{DiscountTarget: recv},
			wantTarget: recv, wantPct: "10",
		},
		{
			name:       "route override percentage",
			promo:      promo(t, "10", true, send),
			rule:       &db.PromoCodeRoute{DiscountTarget: recv, DiscountPercentage: num(t, "5")},
			wantTarget: recv, wantPct: "5",
		},
		{
			name:       "route-restricted, listed route",
			promo:      promo(t, "10", false, send),
			rule:       &db.PromoCodeRoute{DiscountTarget: send},
			wantTarget: send, wantPct: "10",
		},
		{
			name:    "route-restricted, unlisted route",
			promo:   promo(t, "10", false, send),
			wantErr: errNotValidForRoute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, pct, err := ruleForRoute(tt.promo, tt.rule)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if target != tt.wantTarget || !pct.Equal(dec(tt.wantPct)) {
				t.Fatalf("got (%s, %s), want (%s, %s)", target, pct, tt.wantTarget, tt.wantPct)
			}
		})
	}
}

func TestApplySendAmount(t *testing.T) {
	r := Resolved{Target: db.PromoDiscountTargetSENDAMOUNT, Percentage: dec("10")}
	// 100 sent at rate 15 with a fee of 20 -> 1480 received.
	got := r.Apply(dec("100"), dec("15"), dec("1480"))
	want := Result{AmountPaid: dec("90"), AmountReceived: dec("1480"), DiscountAmount: dec("10")}
	assertResult(t, got, want)
}

func TestApplyRecipientAmount(t *testing.T) {
	r := Resolved{Target: db.PromoDiscountTargetRECIPIENTAMOUNT, Percentage: dec("10")}
	// The 10 discount is converted at rate 15 and given to the recipient.
	got := r.Apply(dec("100"), dec("15"), dec("1480"))
	want := Result{AmountPaid: dec("100"), AmountReceived: dec("1630"), BonusAmount: dec("150")}
	assertResult(t, got, want)
}

func TestApplyFullDiscount(t *testing.T) {
	r := Resolved{Target: db.PromoDiscountTargetSENDAMOUNT, Percentage: dec("100")}
	got := r.Apply(dec("50"), dec("2"), dec("90"))
	assertResult(t, got, Result{AmountPaid: dec("0"), AmountReceived: dec("90"), DiscountAmount: dec("50")})
}

func assertResult(t *testing.T, got, want Result) {
	t.Helper()
	if !got.AmountPaid.Equal(want.AmountPaid) ||
		!got.AmountReceived.Equal(want.AmountReceived) ||
		!got.DiscountAmount.Equal(want.DiscountAmount) ||
		!got.BonusAmount.Equal(want.BonusAmount) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestValidate(t *testing.T) {
	id := uuid.New()
	five, zero := dec("5"), dec("0")
	send := db.PromoDiscountTargetSENDAMOUNT
	restricted := false

	base := PromoCodeInput{
		DiscountPercentage: dec("10"),
		StartDate:          time.Now(),
		EndDate:            time.Now().Add(time.Hour),
		MaxUses:            10,
		MaxUsesPerUser:     1,
	}
	with := func(f func(*PromoCodeInput)) PromoCodeInput {
		in := base
		f(&in)
		return in.withDefaults()
	}

	cases := []struct {
		name    string
		in      PromoCodeInput
		wantErr bool
	}{
		{"defaults: all routes, no rules", with(func(*PromoCodeInput) {}), false},
		{"zero percentage", with(func(in *PromoCodeInput) { in.DiscountPercentage = zero }), true},
		{"end before start", with(func(in *PromoCodeInput) { in.EndDate = in.StartDate }), true},
		{"per user above max", with(func(in *PromoCodeInput) { in.MaxUsesPerUser = 11 }), true},
		{"restricted needs routes", with(func(in *PromoCodeInput) { in.AppliesToAllRoutes = &restricted }), true},
		{"restricted with route", with(func(in *PromoCodeInput) {
			in.AppliesToAllRoutes = &restricted
			in.Routes = []RouteRule{{RouteID: id, DiscountTarget: send}}
		}), false},
		{"duplicate route", with(func(in *PromoCodeInput) {
			in.Routes = []RouteRule{{RouteID: id, DiscountTarget: send}, {RouteID: id, DiscountTarget: send}}
		}), true},
		{"route override ok", with(func(in *PromoCodeInput) {
			in.Routes = []RouteRule{{RouteID: id, DiscountTarget: send, DiscountPercentage: &five}}
		}), false},
		{"route override zero", with(func(in *PromoCodeInput) {
			in.Routes = []RouteRule{{RouteID: id, DiscountTarget: send, DiscountPercentage: &zero}}
		}), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.in.validate()
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}

func TestWithDefaults(t *testing.T) {
	in := PromoCodeInput{}.withDefaults()
	if in.DiscountTarget != db.PromoDiscountTargetSENDAMOUNT || in.AppliesToAllRoutes == nil || !*in.AppliesToAllRoutes {
		t.Fatalf("unexpected defaults: %+v", in)
	}
}
