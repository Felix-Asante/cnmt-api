package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const graphAPIBase = "https://graph.facebook.com/v21.0"

type Client struct {
	accessToken   string
	phoneNumberID string
	templateName  string
	templateLang  string
	http          *http.Client
	logger        *slog.Logger
}

type Config struct {
	AccessToken   string
	PhoneNumberID string
	TemplateName  string
	TemplateLang  string
	Logger        *slog.Logger
}

func NewClient(cfg Config) (*Client, error) {
	accessToken := strings.TrimSpace(cfg.AccessToken)
	phoneNumberID := strings.TrimSpace(cfg.PhoneNumberID)
	templateName := strings.TrimSpace(cfg.TemplateName)
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
	if templateName == "" {
		return nil, fmt.Errorf("WHATSAPP_TEMPLATE_NAME is required")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Client{
		accessToken:   accessToken,
		phoneNumberID: phoneNumberID,
		templateName:  templateName,
		templateLang:  templateLang,
		http:          &http.Client{Timeout: 15 * time.Second},
		logger:        logger,
	}, nil
}

type TransferProcessingUpdate struct {
	To        string
	Name      string
	Reference string
}

func (c *Client) SendTransferProcessingUpdate(ctx context.Context, in TransferProcessingUpdate) error {
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
			"name": c.templateName,
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
						{"type": "text", "parameter_name": "reference", "text": url.QueryEscape(ref)},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal whatsapp payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", graphAPIBase, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
			"response", string(respBody),
		)
		return fmt.Errorf("whatsapp API returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	c.logger.Info("whatsapp template sent",
		"to", to,
		"reference", ref,
		"template", c.templateName,
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
