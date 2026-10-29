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
	transferCompletedThanksInterval = 5 * time.Minute
	transferCompletedThanksBatch    = 100
)

type TransferCompletedThanksArgs struct{}

func (TransferCompletedThanksArgs) Kind() string { return "transfer_completed_thanks" }

type TransferCompletedThanksWorker struct {
	river.WorkerDefaults[TransferCompletedThanksArgs]
	Queries  *db.Queries
	Notifier *notifications.Notifier
	Logger   *slog.Logger
}

func (w *TransferCompletedThanksWorker) Work(ctx context.Context, _ *river.Job[TransferCompletedThanksArgs]) error {
	items, err := w.Queries.ListTransfersNeedingCompletionThanks(ctx, transferCompletedThanksBatch)
	if err != nil {
		w.Logger.Error("failed to list transfers needing completion thanks", "error", err)
		return err
	}
	if len(items) == 0 {
		return nil
	}

	w.Logger.Info("sending transfer completion thanks", "count", len(items))

	var firstErr error
	for _, item := range items {
		if err := w.notifyOne(ctx, item); err != nil {
			w.Logger.Error("failed to send transfer completion thanks",
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

func (w *TransferCompletedThanksWorker) notifyOne(ctx context.Context, item db.ListTransfersNeedingCompletionThanksRow) error {
	if err := w.Notifier.TransferCompletedThanks(ctx, notifications.TransferWhatsAppNotice{
		Reference:   item.Reference,
		SenderPhone: item.SenderPhone,
		SenderName:  "Customer",
	}); err != nil {
		return err
	}

	rows, err := w.Queries.MarkTransferCompletionThanksSent(ctx, item.ID)
	if err != nil {
		return err
	}
	if rows == 0 {
		w.Logger.Info("transfer completion thanks already marked sent", "transfer_id", item.ID, "reference", item.Reference)
	}
	return nil
}

func newTransferCompletedThanksPeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(transferCompletedThanksInterval),
		func() (river.JobArgs, *river.InsertOpts) {
			return TransferCompletedThanksArgs{}, nil
		},
		&river.PeriodicJobOpts{RunOnStart: false},
	)
}
