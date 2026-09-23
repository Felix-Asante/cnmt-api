package common

import "testing"

func TestNormalizePhoneValidInternational(t *testing.T) {
	got, err := NormalizePhone("+233274233889")
	if err != nil {
		t.Fatalf("expected valid Ghana number, got %v", err)
	}
	if got != "+233274233889" {
		t.Fatalf("expected +233274233889, got %s", got)
	}

	got, err = NormalizePhone("+23276123456")
	if err != nil {
		t.Fatalf("expected valid SL number, got %v", err)
	}
	if got != "+23276123456" {
		t.Fatalf("expected +23276123456, got %s", got)
	}

	got, err = NormalizePhone("+231770123456")
	if err != nil {
		t.Fatalf("expected valid LR number, got %v", err)
	}
	if got != "+231770123456" {
		t.Fatalf("expected +231770123456, got %s", got)
	}
}

func TestNormalizePhoneRejectsInvalid(t *testing.T) {
	if _, err := NormalizePhone(""); err == nil {
		t.Fatal("expected empty phone to be rejected")
	}
	if _, err := NormalizePhone("76123456"); err == nil {
		t.Fatal("expected national number without country code to be rejected")
	}
	if _, err := NormalizePhone("+233123"); err == nil {
		t.Fatal("expected short number to be rejected")
	}
	if _, err := NormalizePhone("not-a-phone"); err == nil {
		t.Fatal("expected garbage to be rejected")
	}
}
