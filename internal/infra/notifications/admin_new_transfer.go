package notifications

import (
	"context"
	"fmt"

	"cnmt/internal/infra/notifications/email"
)

type AdminNewTransfer struct {
	SourceCountryName          string
	DestinationCountryName     string
	SourceCountryCurrency      string
	DestinationCountryCurrency string
	AmountPaid                 string
	AmountReceived             string
	Fee                        string
}

func (n *Notifier) AdminNewTransfer(ctx context.Context, in AdminNewTransfer) error {
	return n.email.Send(ctx, email.Message{
		To:      []string{n.adminEmail},
		Subject: fmt.Sprintf("New transfer from %s to %s", in.SourceCountryName, in.DestinationCountryName),
		Text: fmt.Sprintf(
			"Amount paid: %s %s\nAmount received: %s %s\nFee: %s %s",
			in.SourceCountryCurrency,in.AmountPaid,
			in.DestinationCountryCurrency,in.AmountReceived,
			in.SourceCountryCurrency,in.Fee,
		),
	})
}
