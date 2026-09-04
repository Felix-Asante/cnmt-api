package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v3"
)

type resendSender struct {
	client *resend.Client
	from   string
	logger *slog.Logger
}

func newResendSender(apiKey, from string, logger *slog.Logger) Sender {
	return &resendSender{
		client: resend.NewClient(apiKey),
		from:   from,
		logger: logger,
	}
}

func (s *resendSender) Send(ctx context.Context, msg Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("email recipient is required")
	}
	if msg.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if msg.Text == "" && msg.HTML == "" {
		return fmt.Errorf("email body is required")
	}

	resp, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.from,
		To:      msg.To,
		Subject: msg.Subject,
		Text:    msg.Text,
		Html:    msg.HTML,
	})
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	if resp == nil || resp.Id == "" {
		return fmt.Errorf("send email: empty response from provider")
	}

	s.logger.Info("email sent", "id", resp.Id, "to", msg.To, "subject", msg.Subject)
	return nil
}
