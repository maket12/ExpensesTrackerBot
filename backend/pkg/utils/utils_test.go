package utils

import (
	"testing"
)

func TestVPtr(t *testing.T) {
	valInt := 42
	ptrInt := VPtr(valInt)
	if ptrInt == nil || *ptrInt != 42 {
		t.Fatalf("expected 42, got %v", ptrInt)
	}

	valStr := "hello"
	ptrStr := VPtr(valStr)
	if ptrStr == nil || *ptrStr != "hello" {
		t.Fatalf("expected 'hello', got %v", ptrStr)
	}
}

func TestComparePtr(t *testing.T) {
	s1 := "test"
	s2 := "test"
	s3 := "diff"

	if !ComparePtr(&s1, &s2) {
		t.Errorf("expected &s1 and &s2 to be equal")
	}

	if ComparePtr(&s1, &s3) {
		t.Errorf("expected &s1 and &s3 to be unequal")
	}

	if !ComparePtr(nil, nil) {
		t.Errorf("expected nil and nil to be equal")
	}

	empty := ""
	if !ComparePtr(nil, &empty) {
		t.Errorf("expected nil and empty to be equal as per ComparePtr implementation")
	}
}

func TestAPIKeyGenerationAndVerification(t *testing.T) {
	raw, hash, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("unexpected error generating api key: %v", err)
	}
	if raw == "" || hash == "" {
		t.Fatalf("expected non-empty raw and hash")
	}

	if !CompareAPIKeyHash(raw, hash) {
		t.Errorf("expected raw key to match hash")
	}

	if CompareAPIKeyHash("invalid-raw-key", hash) {
		t.Errorf("expected invalid raw key to fail comparison")
	}
}
