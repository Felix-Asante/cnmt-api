package common

import (
	"strings"
	"testing"
)

func TestGenerateTransferReference(t *testing.T) {
	ref := GenerateTransferReference("GH", "LR")

	if !strings.HasPrefix(ref, "TRANS-GH-LR-") {
		t.Errorf("expected reference to start with TRANS-GH-LR-, got %s", ref)
	}

	// Should be uppercase and trim spaces
	refLower := GenerateTransferReference(" gh ", "lr ")
	if !strings.HasPrefix(refLower, "TRANS-GH-LR-") {
		t.Errorf("expected trimmed uppercase reference, got %s", refLower)
	}
}

func TestGenerateTransferReferenceUniqueness(t *testing.T) {
	const count = 1000
	seen := make(map[string]struct{}, count)

	for i := 0; i < count; i++ {
		ref := GenerateTransferReference("GH", "LR")
		if _, exists := seen[ref]; exists {
			t.Fatalf("duplicate reference generated: %s on iteration %d", ref, i)
		}
		seen[ref] = struct{}{}
	}
}

func TestGenerateReference(t *testing.T) {
	ref := GenerateReference()
	if !strings.HasPrefix(ref, "CNMT-") {
		t.Errorf("expected legacy reference to start with CNMT-, got %s", ref)
	}
}
