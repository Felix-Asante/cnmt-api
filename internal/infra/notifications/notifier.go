package notifications

import (
	"fmt"
	"log/slog"
	"strings"

	"cnmt/internal/common/env"
	"cnmt/internal/infra/notifications/email"
)

type Notifier struct {
	adminEmail string
	email      email.Sender
}

func NewFromEnv(logger *slog.Logger) (*Notifier, error) {
	sender, err := email.NewSender(
		env.GetString("EMAIL_API_KEY", ""),
		env.GetString("EMAIL_FROM", ""),
		logger,
	)
	if err != nil {
		return nil, err
	}

	adminEmail := strings.TrimSpace(env.GetString("ADMIN_EMAIL", ""))
	if adminEmail == "" {
		return nil, fmt.Errorf("ADMIN_EMAIL is required")
	}

	return &Notifier{
		adminEmail: adminEmail,
		email:      sender,
	}, nil
}
