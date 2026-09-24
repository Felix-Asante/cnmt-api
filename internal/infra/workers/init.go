package workers

import (
	"fmt"
	"log/slog"

	"cnmt/internal/infra/db"
	"cnmt/internal/infra/notifications"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type workers struct {
	db       *pgxpool.Pool
	notifier *notifications.Notifier
	logger   *slog.Logger
}

func NewWorkers(db *pgxpool.Pool, notifier *notifications.Notifier, logger *slog.Logger) *workers {
	return &workers{db: db, notifier: notifier, logger: logger}
}

func (w *workers) Init() (*river.Client[pgx.Tx], error) {
	queries := db.New(w.db)
	registered := setupWorkers(queries, w.notifier, w.logger)

	client, err := river.NewClient(riverpgxv5.New(w.db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers:      registered,
		PeriodicJobs: []*river.PeriodicJob{newTransferAssurancePeriodicJob()},
	})
	if err != nil {
		return nil, fmt.Errorf("create river client: %w", err)
	}
	return client, nil
}

func setupWorkers(queries *db.Queries, notifier *notifications.Notifier, logger *slog.Logger) *river.Workers {
	registered := river.NewWorkers()
	river.AddWorker(registered, &NewTransferWorker{Notifier: notifier})
	river.AddWorker(registered, &TransferAssuranceWorker{
		Queries:  queries,
		Notifier: notifier,
		Logger:   logger,
	})
	return registered
}
