package password

import "testing"

func TestHashAndCompare(t *testing.T) {
	encoded, err := Hash("secure-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !Compare(encoded, "secure-password") {
		t.Fatal("expected password to match")
	}

	if Compare(encoded, "wrong-password") {
		t.Fatal("expected password mismatch")
	}
}

func TestCompareRejectsInvalidHash(t *testing.T) {
	if Compare("not-a-valid-hash", "password") {
		t.Fatal("expected invalid hash to fail comparison")
	}
}
