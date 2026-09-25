package qr_test

import (
	"strings"
	"testing"

	"github.com/drug-verification/server/pkg/qr"
)

func TestGenerateURL(t *testing.T) {
	generator := qr.NewGenerator("https://verify.health.gov/check")
	serial := "MED-2026-XYZ123"

	url := generator.GenerateURL(serial)
	expected := "https://verify.health.gov/check/MED-2026-XYZ123"

	if url != expected {
		t.Errorf("expected URL %q, got %q", expected, url)
	}
}

func TestGeneratePNG(t *testing.T) {
	generator := qr.NewGenerator("https://verify.health.gov/check")
	serial := "TEST-SERIAL-456"

	png, err := generator.GeneratePNG(serial, 256)
	if err != nil {
		t.Fatalf("failed to generate PNG: %v", err)
	}

	if len(png) == 0 {
		t.Fatalf("expected non-empty PNG byte slice")
	}

	// Verify PNG header bytes: 0x89 'P' 'N' 'G'
	if len(png) < 8 || string(png[1:4]) != "PNG" {
		t.Errorf("expected PNG header magic bytes, got %v", png[:8])
	}

	_ = strings.Contains
}
