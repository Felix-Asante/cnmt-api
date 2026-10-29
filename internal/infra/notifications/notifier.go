package notifications

import (
	"fmt"
	"log/slog"
	"strings"

	"cnmt/internal/common/env"
	"cnmt/internal/infra/notifications/email"
	"cnmt/internal/infra/notifications/whatsapp"
)

type Notifier struct {
	adminEmail string
	email      email.Sender
	whatsapp   *whatsapp.Client
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

	waClient, err := whatsapp.NewClient(whatsapp.Config{
		AccessToken:        env.GetString("WHATSAPP_ACCESS_TOKEN", ""),
		PhoneNumberID:      env.GetString("WHATSAPP_PHONE_NUMBER_ID", ""),
		ProcessingTemplate: env.GetString("WHATSAPP_TEMPLATE_NAME", ""),
		ThanksTemplate:     env.GetString("WHATSAPP_THANKS_TEMPLATE_NAME", ""),
		TemplateLang:       env.GetString("WHATSAPP_TEMPLATE_LANGUAGE", "en_US"),
		Logger:             logger,
	})
	if err != nil {
		return nil, fmt.Errorf("whatsapp: %w", err)
	}

	return &Notifier{
		adminEmail: adminEmail,
		email:      sender,
		whatsapp:   waClient,
	}, nil
}
