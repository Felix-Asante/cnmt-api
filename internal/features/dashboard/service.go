package dashboard

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"cnmt/internal/common"
	"cnmt/internal/common/httpx"
	"cnmt/internal/infra/db"

	"github.com/shopspring/decimal"
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger
}

func NewService(queries *db.Queries, logger *slog.Logger) *Service {
	return &Service{queries: queries, logger: logger}
}

func (s *Service) Get(ctx context.Context, period Period) (Response, error) {
	expiringBefore := time.Now().UTC().Add(expiringWindow)

	statusRows, err := s.queries.DashboardStatusCounts(ctx, db.DashboardStatusCountsParams{
		FromTs: period.From,
		ToTs:   period.To,
	})
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	moneyRows, err := s.queries.DashboardMoneyTotals(ctx, db.DashboardMoneyTotalsParams{
		FromTs: period.From,
		ToTs:   period.To,
	})
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	volumeRows, err := s.queries.DashboardDailyVolume(ctx, db.DashboardDailyVolumeParams{
		FromTs: period.From,
		ToTs:   period.To,
	})
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	actionCounts, err := s.queries.DashboardActionRequiredCounts(ctx, expiringBefore)
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	actionTransfers, err := s.queries.DashboardActionRequiredTransfers(ctx, db.DashboardActionRequiredTransfersParams{
		ExpiringBefore: expiringBefore,
		RowLimit:       defaultRecentLimit,
	})
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	recentTransfers, err := s.queries.DashboardRecentTransfers(ctx, db.DashboardRecentTransfersParams{
		FromTs:   period.From,
		ToTs:     period.To,
		RowLimit: defaultRecentLimit,
	})
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	topRoutes, err := s.queries.DashboardTopRoutes(ctx, db.DashboardTopRoutesParams{
		FromTs:   period.From,
		ToTs:     period.To,
		RowLimit: defaultTopRoutes,
	})
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	activity, err := s.queries.DashboardRecentActivity(ctx, defaultActivityLimit)
	if err != nil {
		return Response{}, common.TranslateDBError(err)
	}

	overview, err := buildOverview(statusRows, moneyRows)
	if err != nil {
		s.logger.Error("failed to build dashboard overview", "error", err)
		return Response{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	volume, err := mapVolume(volumeRows)
	if err != nil {
		s.logger.Error("failed to map dashboard volume", "error", err)
		return Response{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	actionItems, err := mapActionRequiredTransfers(actionTransfers)
	if err != nil {
		s.logger.Error("failed to map action-required transfers", "error", err)
		return Response{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	recent, err := mapRecentTransfers(recentTransfers)
	if err != nil {
		s.logger.Error("failed to map recent transfers", "error", err)
		return Response{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	routes, err := mapTopRoutes(topRoutes)
	if err != nil {
		s.logger.Error("failed to map top routes", "error", err)
		return Response{}, fmt.Errorf("%w", httpx.InternalServerError)
	}

	return Response{
		Period:   period,
		Overview: overview,
		ActionRequired: ActionRequired{
			PaymentVerificationCount: actionCounts.PaymentVerificationCount,
			ProcessingCount:          actionCounts.ProcessingCount,
			ExpiringCount:            actionCounts.ExpiringCount,
			Transfers:                actionItems,
		},
		Volume:             volume,
		StatusDistribution: mapStatusDistribution(statusRows),
		RecentTransfers:    recent,
		TopRoutes:          routes,
		RecentActivity:     mapRecentActivity(activity),
	}, nil
}

func buildOverview(
	statusRows []db.DashboardStatusCountsRow,
	moneyRows []db.DashboardMoneyTotalsRow,
) (Overview, error) {
	overview := Overview{
		TotalTransferVolume: []MoneyByCurrency{},
		TotalFees:           []MoneyByCurrency{},
	}

	byStatus := make(map[db.TransferStatus]int64, len(statusRows))
	for _, row := range statusRows {
		byStatus[row.Status] = row.Count
		overview.TotalTransfers += row.Count
	}

	overview.PendingPayment = byStatus[db.TransferStatusPENDINGPAYMENT]
	overview.PaymentVerification = byStatus[db.TransferStatusPAYMENTRECEIVED] + byStatus[db.TransferStatusVERIFYING]
	overview.Processing = byStatus[db.TransferStatusPROCESSING]
	overview.Completed = byStatus[db.TransferStatusCOMPLETED]
	overview.Cancelled = byStatus[db.TransferStatusCANCELLED]
	overview.Failed = byStatus[db.TransferStatusFAILED]

	volumeByCurrency := map[string]decimal.Decimal{}
	feesByCurrency := map[string]decimal.Decimal{}
	for _, row := range moneyRows {
		sent, err := common.PgNumericToDecimal(row.TotalAmountSent)
		if err != nil {
			return Overview{}, err
		}
		fees, err := common.PgNumericToDecimal(row.TotalFees)
		if err != nil {
			return Overview{}, err
		}
		volumeByCurrency[row.SourceCurrencyCode] = volumeByCurrency[row.SourceCurrencyCode].Add(sent)
		feesByCurrency[row.DestinationCurrencyCode] = feesByCurrency[row.DestinationCurrencyCode].Add(fees)
	}

	overview.TotalTransferVolume = moneyMapToSlice(volumeByCurrency)
	overview.TotalFees = moneyMapToSlice(feesByCurrency)
	return overview, nil
}

func moneyMapToSlice(m map[string]decimal.Decimal) []MoneyByCurrency {
	out := make([]MoneyByCurrency, 0, len(m))
	for currency, amount := range m {
		out = append(out, MoneyByCurrency{Currency: currency, Amount: amount})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Currency < out[j].Currency })
	return out
}

func mapStatusDistribution(rows []db.DashboardStatusCountsRow) []StatusCount {
	out := make([]StatusCount, 0, len(rows))
	for _, row := range rows {
		out = append(out, StatusCount{Status: row.Status, Count: row.Count})
	}
	return out
}

func mapVolume(rows []db.DashboardDailyVolumeRow) ([]VolumePoint, error) {
	out := make([]VolumePoint, 0, len(rows))
	for _, row := range rows {
		if !row.Day.Valid {
			continue
		}
		volume, err := common.PgNumericToDecimal(row.Volume)
		if err != nil {
			return nil, err
		}
		out = append(out, VolumePoint{
			Date:          row.Day.Time.Format("2006-01-02"),
			Currency:      row.CurrencyCode,
			TransferCount: row.TransferCount,
			Volume:        volume,
		})
	}
	return out, nil
}

func mapActionRequiredTransfers(rows []db.DashboardActionRequiredTransfersRow) ([]TransferSummary, error) {
	out := make([]TransferSummary, 0, len(rows))
	for _, row := range rows {
		amount, err := common.PgNumericToDecimal(row.AmountSent)
		if err != nil {
			return nil, err
		}
		expiresAt := row.ExpiresAt
		out = append(out, TransferSummary{
			ID:                   row.ID,
			Reference:            row.Reference,
			Status:               row.Status,
			SourceCountry:        row.SourceCountryName,
			DestinationCountry:   row.DestinationCountryName,
			AmountSent:           amount,
			CurrencyCode:         row.SourceCurrencyCode,
			CurrencySymbol:       row.SourceCurrencySymbol,
			SenderPhone:          row.SenderPhone,
			ReceivingAccountName: row.ReceivingAccountName,
			ReceivingMethod:      row.ReceivingMethod,
			ExpiresAt:            &expiresAt,
			CreatedAt:            row.CreatedAt,
		})
	}
	return out, nil
}

func mapRecentTransfers(rows []db.DashboardRecentTransfersRow) ([]TransferSummary, error) {
	out := make([]TransferSummary, 0, len(rows))
	for _, row := range rows {
		amount, err := common.PgNumericToDecimal(row.AmountSent)
		if err != nil {
			return nil, err
		}
		out = append(out, TransferSummary{
			ID:                   row.ID,
			Reference:            row.Reference,
			Status:               row.Status,
			SourceCountry:        row.SourceCountryName,
			DestinationCountry:   row.DestinationCountryName,
			AmountSent:           amount,
			CurrencyCode:         row.SourceCurrencyCode,
			CurrencySymbol:       row.SourceCurrencySymbol,
			SenderPhone:          row.SenderPhone,
			ReceivingAccountName: row.ReceivingAccountName,
			ReceivingMethod:      row.ReceivingMethod,
			CreatedAt:            row.CreatedAt,
		})
	}
	return out, nil
}

func mapTopRoutes(rows []db.DashboardTopRoutesRow) ([]TopRoute, error) {
	out := make([]TopRoute, 0, len(rows))
	for _, row := range rows {
		volume, err := common.PgNumericToDecimal(row.TransferVolume)
		if err != nil {
			return nil, err
		}
		out = append(out, TopRoute{
			RouteID:             row.RouteID,
			SourceCountry:       row.SourceCountryName,
			SourceISOCode:       row.SourceIsoCode,
			SourceFlag:          row.SourceFlag,
			SourceCurrency:      row.SourceCurrencyCode,
			DestinationCountry:  row.DestinationCountryName,
			DestinationISOCode:  row.DestinationIsoCode,
			DestinationFlag:     row.DestinationFlag,
			DestinationCurrency: row.DestinationCurrencyCode,
			TransferCount:       row.TransferCount,
			TransferVolume:      volume,
		})
	}
	return out, nil
}

func mapRecentActivity(rows []db.DashboardRecentActivityRow) []ActivityItem {
	out := make([]ActivityItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, ActivityItem{
			Status:    row.Status,
			Reference: row.Reference,
			Actor:     row.Actor,
			Note:      row.Note,
			CreatedAt: row.CreatedAt,
		})
	}
	return out
}
