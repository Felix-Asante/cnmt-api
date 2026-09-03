package workers

import (
	"context"
	"fmt"

	"cnmt/internal/common/env"

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

// name used to identify the job in the queue
func (NewTransferArgs) Kind() string { return "new_transfer" }

// define job worker

type NewTransferWorker struct {
	river.WorkerDefaults[NewTransferArgs]
}

func (w *NewTransferWorker) Work(ctx context.Context, job *river.Job[NewTransferArgs]) error {
	return sendAdminNewTransferEmail(ctx, SendAdminNewTransferEmailArgs{
		SourceCountryName:          job.Args.SourceCountryName,
		DestinationCountryName:     job.Args.DestinationCountryName,
		SourceCountryCurrency:      job.Args.SourceCountryCurrency,
		DestinationCountryCurrency: job.Args.DestinationCountryCurrency,
		AmountPaid:                 job.Args.AmountPaid,
		AmountReceived:             job.Args.AmountReceived,
		Fee:                        job.Args.Fee,
	})
}

type SendAdminNewTransferEmailArgs struct {
	SourceCountryName          string
	DestinationCountryName string
	SourceCountryCurrency      string
	DestinationCountryCurrency string
	AmountPaid                 string
	AmountReceived             string
	Fee                        string
}

func sendAdminNewTransferEmail(ctx context.Context, args SendAdminNewTransferEmailArgs) error {
	_ = ctx
	email := env.GetString("ADMIN_EMAIL", "")
	if email == "" {
		return fmt.Errorf("ADMIN_EMAIL is not set")
	}
	subject := fmt.Sprintf("New transfer from %s to %s", args.SourceCountryName, args.DestinationCountryName)
	body := fmt.Sprintf("Amount paid: %s %s\nAmount received: %s %s\nFee: %s %s", args.AmountPaid, args.SourceCountryCurrency, args.AmountReceived, args.DestinationCountryCurrency, args.Fee, args.SourceCountryCurrency)

	fmt.Printf("Sending email to %s with subject %s and body %s", email, subject, body)
	return nil
}