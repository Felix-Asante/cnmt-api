package notifications

import (
	"context"
	"fmt"
	"strings"

	"cnmt/internal/infra/notifications/whatsapp"
)

type TransferWhatsAppNotice struct {
	Reference   string
	SenderPhone string
	SenderName  string
}

func (n *Notifier) TransferAssurance(ctx context.Context, in TransferWhatsAppNotice) error {
	return n.sendTransferWhatsApp(ctx, in, n.whatsapp.SendTransferProcessingUpdate)
}

func (n *Notifier) TransferCompletedThanks(ctx context.Context, in TransferWhatsAppNotice) error {
	return n.sendTransferWhatsApp(ctx, in, n.whatsapp.SendTransferCompletedThanks)
}

func (n *Notifier) sendTransferWhatsApp(
	ctx context.Context,
	in TransferWhatsAppNotice,
	send func(context.Context, whatsapp.TemplateMessage) error,
) error {
	if n.whatsapp == nil {
		return fmt.Errorf("whatsapp is not configured")
	}
	if strings.TrimSpace(in.SenderPhone) == "" {
		return fmt.Errorf("sender phone is required")
	}
	if strings.TrimSpace(in.Reference) == "" {
		return fmt.Errorf("transfer reference is required")
	}

	return send(ctx, whatsapp.TemplateMessage{
		To:        in.SenderPhone,
		Name:      in.SenderName,
		Reference: in.Reference,
	})
}
