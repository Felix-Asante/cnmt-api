package workers

import (
	"context"

	"cnmt/internal/infra/notifications"

	"github.com/riverqueue/river"
)

type NewTransferArgs struct {
	Reference                  string `json:"reference"`
	SourceCountryName          string `json:"source_country_name"`
	DestinationCountryName     string `json:"destination_country_name"`
	SourceCountryCurrency      string `json:"source_country_currency"`
	DestinationCountryCurrency string `json:"destination_country_currency"`
	AmountPaid                 string `json:"amount_paid"`
	AmountReceived             string `json:"amount_received"`
	Fee                        string `json:"fee"`
}

func (NewTransferArgs) Kind() string { return "new_transfer" }

type NewTransferWorker struct {
	river.WorkerDefaults[NewTransferArgs]
	Notifier *notifications.Notifier
}

func (w *NewTransferWorker) Work(ctx context.Context, job *river.Job[NewTransferArgs]) error {
	if err := w.Notifier.AdminNewTransfer(ctx, notifications.AdminNewTransfer{
		SourceCountryName:          job.Args.SourceCountryName,
		DestinationCountryName:     job.Args.DestinationCountryName,
		SourceCountryCurrency:      job.Args.SourceCountryCurrency,
		DestinationCountryCurrency: job.Args.DestinationCountryCurrency,
		AmountPaid:                 job.Args.AmountPaid,
		AmountReceived:             job.Args.AmountReceived,
		Fee:                        job.Args.Fee,
	}); err != nil {
		return err
	}
	return nil
}
