package workers

import (
	"context"

	"github.com/riverqueue/river"
)

type NewTransferArgs struct {
	Reference string `json:"reference"`
	SourceCountryName string `json:"source_country_name"`
	DestinationCountryName string `json:"destination_country_name"`
	Amount string `json:"amount"`
	Currency string `json:"currency"`
}


// name used to identify the job in the queue
func (NewTransferArgs) Kind() string { return "new_transfer"}

// define job worker

type NewTransferWorker struct {
	river.WorkerDefaults[NewTransferArgs]
}

func (w *NewTransferWorker) Work(ctx context.Context, job *river.Job[NewTransferArgs]) error {
	if job.FinalizedAt != nil {
		return nil
	}

	
	return nil
}