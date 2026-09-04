package email

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

func NewSender(apiKey, from string, logger *slog.Logger) (Sender, error) {
	apiKey = strings.TrimSpace(apiKey)
	from = strings.TrimSpace(from)
	if apiKey == "" {
		return nil, fmt.Errorf("EMAIL_API_KEY is required")
	}
	if from == "" {
		return nil, fmt.Errorf("EMAIL_FROM is required")
	}
	return newResendSender(apiKey, from, logger), nil
}
