package utils

import (
	"strings"
	"testing"
)

func TestMaskSensitiveJSON_NestedArray(t *testing.T) {
	raw := `{"items":[{"token":"abc"}],"password":"secret"}`
	got := MaskSensitiveJSON(raw)
	if strings.Contains(got, "abc") {
		t.Fatalf("nested token not masked: %s", got)
	}
	if !strings.Contains(got, maskedValue) {
		t.Fatalf("expected masked value in %s", got)
	}
}

func TestMaskEmail(t *testing.T) {
	got := MaskEmail("alice@example.com")
	want := "a***@example.com"
	if got != want {
		t.Fatalf("MaskEmail = %q, want %q", got, want)
	}
}