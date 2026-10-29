package common

import "testing"

func TestNormalizeAndValidateIBANValid(t *testing.T) {
	got, err := NormalizeAndValidateIBAN("DE89 3704 0044 0532 0130 00")
	if err != nil {
		t.Fatalf("expected valid IBAN, got %v", err)
	}
	if got != "DE89370400440532013000" {
		t.Fatalf("unexpected normalized IBAN: %s", got)
	}
}

func TestNormalizeAndValidateIBANRejectsInvalid(t *testing.T) {
	for _, raw := range []string{"", "DE89", "DE89370400440532013001"} {
		if _, err := NormalizeAndValidateIBAN(raw); err == nil {
			t.Fatalf("expected rejection for %q", raw)
		}
	}
}
