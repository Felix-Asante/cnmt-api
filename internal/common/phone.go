package common

import (
	"fmt"
	"strings"

	"cnmt/internal/common/httpx"

	"github.com/nyaruka/phonenumbers"
)

// NormalizePhone parses raw as an international phone number, checks it is valid,
// and returns it in E.164 for storage. Pass numbers with a country code (e.g. +233...).
func NormalizePhone(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("%w: phone number is required", httpx.BadRequestError)
	}

	num, err := phonenumbers.Parse(raw, "ZZ")
	if err != nil {
		return "", fmt.Errorf("%w: phone number is not valid", httpx.BadRequestError)
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("%w: phone number is not valid", httpx.BadRequestError)
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}
