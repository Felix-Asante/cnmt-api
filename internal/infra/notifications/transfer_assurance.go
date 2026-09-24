package notifications

import (
	"context"
	"fmt"
	"log/slog"
)

type TransferAssurance struct {
	Reference   string
	SenderPhone string
}

func (n *Notifier) TransferAssurance(ctx context.Context, in TransferAssurance) error {
	if in.SenderPhone == "" {
		return fmt.Errorf("sender phone is required")
	}

	message := fmt.Sprintf(
		"Hi, we received your payment for transfer %s. It is being processed and we will update you when it is complete.",
		in.Reference,
	)

	// TODO: send `message` to in.SenderPhone via WhatsApp.
	slog.Info("transfer assurance ready",
		"phone", in.SenderPhone,
		"reference", in.Reference,
		"message", message,
	)
	return nil
}
