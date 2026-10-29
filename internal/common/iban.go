package common

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"

	"cnmt/internal/common/httpx"
)

func NormalizeAndValidateIBAN(raw string) (string, error) {
	iban := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return unicode.ToUpper(r)
	}, strings.TrimSpace(raw))

	if iban == "" {
		return "", fmt.Errorf("%w: iban is required", httpx.BadRequestError)
	}
	if len(iban) < 15 || len(iban) > 34 {
		return "", fmt.Errorf("%w: iban is not valid", httpx.BadRequestError)
	}
	for i, r := range iban {
		if i < 2 {
			if r < 'A' || r > 'Z' {
				return "", fmt.Errorf("%w: iban is not valid", httpx.BadRequestError)
			}
			continue
		}
		if (r < '0' || r > '9') && (r < 'A' || r > 'Z') {
			return "", fmt.Errorf("%w: iban is not valid", httpx.BadRequestError)
		}
	}

	rearranged := iban[4:] + iban[:4]
	var expanded strings.Builder
	for _, r := range rearranged {
		switch {
		case r >= '0' && r <= '9':
			expanded.WriteByte(byte(r))
		case r >= 'A' && r <= 'Z':
			expanded.WriteString(fmt.Sprintf("%d", int(r-'A')+10))
		default:
			return "", fmt.Errorf("%w: iban is not valid", httpx.BadRequestError)
		}
	}

	n := new(big.Int)
	if _, ok := n.SetString(expanded.String(), 10); !ok {
		return "", fmt.Errorf("%w: iban is not valid", httpx.BadRequestError)
	}
	if new(big.Int).Mod(n, big.NewInt(97)).Int64() != 1 {
		return "", fmt.Errorf("%w: iban is not valid", httpx.BadRequestError)
	}
	return iban, nil
}
