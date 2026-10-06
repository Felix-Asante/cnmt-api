package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"cnmt/internal/common/env"
)

// Meta WhatsApp webhook callbacks.
// Configure Callback URL in Meta as: GET+POST /api/v1/whatsapp/cb
func whatsappVerifyHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	expected := env.GetString("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "")
	if mode == "subscribe" && expected != "" && token == expected {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(challenge))
		return
	}

	http.Error(w, "forbidden", http.StatusForbidden)
}

func whatsappWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if appSecret := env.GetString("WHATSAPP_APP_SECRET", ""); appSecret != "" {
		if !validWhatsAppSignature(r.Header.Get("X-Hub-Signature-256"), body, appSecret) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	slog.Info("whatsapp webhook received", "bytes", len(body), "payload", string(body))
	w.WriteHeader(http.StatusOK)
}

func validWhatsAppSignature(header string, body []byte, appSecret string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	want, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), want)
}
