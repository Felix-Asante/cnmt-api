package workers

import (
	"context"
	"log/slog"
	"time"

	"cnmt/internal/infra/db"
	"cnmt/internal/infra/notifications"

	"github.com/riverqueue/river"
)

const (
	transferAssuranceAge      = 30 * time.Minute
	transferAssuranceInterval = 5 * time.Minute
	transferAssuranceBatch    = 100
)

type TransferAssuranceArgs struct{}

func (TransferAssuranceArgs) Kind() string { return "transfer_assurance" }

type TransferAssuranceWorker struct {
	river.WorkerDefaults[TransferAssuranceArgs]
	Queries  *db.Queries
	Notifier *notifications.Notifier
	Logger   *slog.Logger
}

func (w *TransferAssuranceWorker) Work(ctx context.Context, _ *river.Job[TransferAssuranceArgs]) error {
	items, err := w.Queries.ListTransfersNeedingAssurance(ctx, db.ListTransfersNeedingAssuranceParams{
		OlderThan: time.Now().UTC().Add(-transferAssuranceAge),
		RowLimit:  transferAssuranceBatch,
	})
	if err != nil {
		w.Logger.Error("failed to list transfers needing assurance", "error", err)
		return err
	}
	if len(items) == 0 {
		return nil
	}

	w.Logger.Info("sending transfer assurance notifications", "count", len(items))

	var firstErr error
	for _, item := range items {
		if err := w.notifyOne(ctx, item); err != nil {
			w.Logger.Error("failed to send transfer assurance",
				"error", err,
				"transfer_id", item.ID,
				"reference", item.Reference,
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (w *TransferAssuranceWorker) notifyOne(ctx context.Context, item db.ListTransfersNeedingAssuranceRow) error {
	if err := w.Notifier.TransferAssurance(ctx, notifications.TransferAssurance{
		Reference:   item.Reference,
		SenderPhone: item.SenderPhone,
	}); err != nil {
		return err
	}

	rows, err := w.Queries.MarkTransferAssuranceSent(ctx, item.ID)
	if err != nil {
		return err
	}
	if rows == 0 {
		w.Logger.Info("transfer assurance already marked sent", "transfer_id", item.ID, "reference", item.Reference)
	}
	return nil
}

func newTransferAssurancePeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(transferAssuranceInterval),
		func() (river.JobArgs, *river.InsertOpts) {
			return TransferAssuranceArgs{}, nil
		},
		&river.PeriodicJobOpts{RunOnStart: false},
	)
}
