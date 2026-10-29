package notifications

import (
	"context"
	"fmt"
	"strings"

	"cnmt/internal/infra/notifications/whatsapp"
)

type TransferAssurance struct {
	Reference   string
	SenderPhone string
	SenderName  string
}

func (n *Notifier) TransferAssurance(ctx context.Context, in TransferAssurance) error {
	if n.whatsapp == nil {
		return fmt.Errorf("whatsapp is not configured")
	}
	if strings.TrimSpace(in.SenderPhone) == "" {
		return fmt.Errorf("sender phone is required")
	}
	if strings.TrimSpace(in.Reference) == "" {
		return fmt.Errorf("transfer reference is required")
	}

	return n.whatsapp.SendTransferProcessingUpdate(ctx, whatsapp.TransferProcessingUpdate{
		To:        in.SenderPhone,
		Name:      in.SenderName,
		Reference: in.Reference,
	})
}
