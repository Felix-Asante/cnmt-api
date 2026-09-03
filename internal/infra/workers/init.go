package workers

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type workers struct {
	db *pgxpool.Pool
}

func NewWorkers(db *pgxpool.Pool) *workers {
	return &workers{db: db}
}

func (w *workers) Init() (*river.Client[pgx.Tx], error) {
	workers := setupWorkers()
	client, err := river.NewClient(riverpgxv5.New(w.db), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers: workers,
	})
	return client, err
}

func setupWorkers() *river.Workers {
	workers := river.NewWorkers()
	river.AddWorker(workers, &NewTransferWorker{})

	return workers
}