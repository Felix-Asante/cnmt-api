package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const graphAPIBase = "https://graph.facebook.com/v21.0"

type Client struct {
	accessToken         string
	phoneNumberID       string
	templateLang        string
	processingTemplate  string
	thanksTemplate      string
	http                *http.Client
	logger              *slog.Logger
}

type Config struct {
	AccessToken        string
	PhoneNumberID      string
	TemplateLang       string
	ProcessingTemplate string
	ThanksTemplate     string
	Logger             *slog.Logger
}

type TemplateMessage struct {
	To        string
	Name      string
	Reference string
}

func NewClient(cfg Config) (*Client, error) {
	accessToken := strings.TrimSpace(cfg.AccessToken)
	phoneNumberID := strings.TrimSpace(cfg.PhoneNumberID)
	processingTemplate := strings.TrimSpace(cfg.ProcessingTemplate)
	thanksTemplate := strings.TrimSpace(cfg.ThanksTemplate)
	templateLang := strings.TrimSpace(cfg.TemplateLang)
	if templateLang == "" {
		templateLang = "en_US"
	}

	if accessToken == "" {
		return nil, fmt.Errorf("WHATSAPP_ACCESS_TOKEN is required")
	}
	if phoneNumberID == "" {
		return nil, fmt.Errorf("WHATSAPP_PHONE_NUMBER_ID is required")
	}
	if processingTemplate == "" {
		return nil, fmt.Errorf("WHATSAPP_TEMPLATE_NAME is required")
	}
	if thanksTemplate == "" {
		return nil, fmt.Errorf("WHATSAPP_THANKS_TEMPLATE_NAME is required")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Client{
		accessToken:        accessToken,
		phoneNumberID:      phoneNumberID,
		templateLang:       templateLang,
		processingTemplate: processingTemplate,
		thanksTemplate:     thanksTemplate,
		http:               &http.Client{Timeout: 15 * time.Second},
		logger:             logger,
	}, nil
}

func (c *Client) SendTransferProcessingUpdate(ctx context.Context, in TemplateMessage) error {
	return c.sendNamedTemplate(ctx, c.processingTemplate, in)
}

func (c *Client) SendTransferCompletedThanks(ctx context.Context, in TemplateMessage) error {
	return c.sendNamedTemplate(ctx, c.thanksTemplate, in)
}

func (c *Client) sendNamedTemplate(ctx context.Context, templateName string, in TemplateMessage) error {
	to := normalizeWhatsAppTo(in.To)
	if to == "" {
		return fmt.Errorf("recipient phone is required")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "Customer"
	}
	ref := strings.TrimSpace(in.Reference)
	if ref == "" {
		return fmt.Errorf("transfer reference is required")
	}

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "template",
		"template": map[string]any{
			"name": templateName,
			"language": map[string]string{
				"code": c.templateLang,
			},
			"components": []map[string]any{
				{
					"type": "body",
					"parameters": []map[string]string{
						{"type": "text", "parameter_name": "name", "text": name},
						{"type": "text", "parameter_name": "reference", "text": ref},
					},
				},
				{
					"type":     "button",
					"sub_type": "url",
					"index":    "0",
					"parameters": []map[string]string{
						{"type": "text", "text": ref},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal whatsapp payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/%s/messages", graphAPIBase, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create whatsapp request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send whatsapp message: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Error("whatsapp send failed",
			"status", resp.StatusCode,
			"to", to,
			"reference", ref,
			"template", templateName,
			"response", string(respBody),
		)
		return fmt.Errorf("whatsapp API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	c.logger.Info("whatsapp template sent",
		"to", to,
		"reference", ref,
		"template", templateName,
	)
	return nil
}

func normalizeWhatsAppTo(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
