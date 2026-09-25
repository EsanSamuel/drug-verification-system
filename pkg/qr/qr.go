// Package qr provides QR code generation for drug verification.
package qr

import (
	"fmt"
	"net/url"

	qrcode "github.com/skip2/go-qrcode"
)

// Generator creates QR codes for drug unit verification.
type Generator struct {
	baseURL string
}

// NewGenerator creates a new QR generator with the given verification base URL.
func NewGenerator(baseURL string) *Generator {
	return &Generator{baseURL: baseURL}
}

// GenerateURL returns the full verification URL for a serial number.
func (g *Generator) GenerateURL(serialNumber string) string {
	return fmt.Sprintf("%s/%s", g.baseURL, url.PathEscape(serialNumber))
}

// GeneratePNG returns a PNG-encoded QR code image for the given serial number.
// The size parameter specifies the image width/height in pixels.
func (g *Generator) GeneratePNG(serialNumber string, size int) ([]byte, error) {
	verifyURL := g.GenerateURL(serialNumber)
	png, err := qrcode.Encode(verifyURL, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("qr: encode: %w", err)
	}
	return png, nil
}
