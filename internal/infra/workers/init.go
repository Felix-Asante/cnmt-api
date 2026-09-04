package workers

import (
	"cnmt/internal/infra/notifications"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type workers struct {
	db       *pgxpool.Pool
	notifier *notifications.Notifier
}

func NewWorkers(db *pgxpool.Pool, notifier *notifications.Notifier) *workers {
	return &workers{db: db, notifier: notifier}
}

func (w *workers) Init() (*river.Client[pgx.Tx], error) {
	workers := setupWorkers(w.notifier)
	client, err := river.NewClient(riverpgxv5.New(w.db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
	})
	return client, err
}

func setupWorkers(notifier *notifications.Notifier) *river.Workers {
	registered := river.NewWorkers()
	river.AddWorker(registered, &NewTransferWorker{Notifier: notifier})
	return registered
}
