package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/table-order/config"
)

// TestUploadImage_NotConfigured verifies the handler returns 503 when R2 is not
// configured. Needs no database or R2 — runs anywhere.
func TestUploadImage_NotConfigured(t *testing.T) {
	config.R2Client = nil // ensure unconfigured

	r := setupRouter()
	r.POST("/api/merchant/upload", UploadImage)

	req, _ := http.NewRequest("POST", "/api/merchant/upload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when R2 unconfigured, got %d body: %s", w.Code, w.Body.String())
	}
}

var (
	pngMagic  = []byte("\x89PNG\r\n\x1a\n")
	jpegMagic = []byte("\xff\xd8\xff\xe0")
)

// Byte-exact round-trip on an input larger than the 512-byte sniff window —
// guards against the short-read/reassembly bug.
func TestValidateAndReadImage_AcceptsPNGByteExact(t *testing.T) {
	input := append(append([]byte{}, pngMagic...), bytes.Repeat([]byte{0xAB}, 1000)...)

	data, ct, ext, err := validateAndReadImage(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct != "image/png" || ext != ".png" {
		t.Errorf("expected image/png .png, got %s %s", ct, ext)
	}
	if !bytes.Equal(data, input) {
		t.Errorf("data not byte-exact: got %d bytes, want %d", len(data), len(input))
	}
}

func TestValidateAndReadImage_AcceptsJPEG(t *testing.T) {
	input := append(append([]byte{}, jpegMagic...), bytes.Repeat([]byte{0x00}, 600)...)
	_, ct, ext, err := validateAndReadImage(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct != "image/jpeg" || ext != ".jpg" {
		t.Errorf("expected image/jpeg .jpg, got %s %s", ct, ext)
	}
}

func TestValidateAndReadImage_RejectsNonImage(t *testing.T) {
	input := []byte("this is plainly not an image file at all")
	_, _, _, err := validateAndReadImage(bytes.NewReader(input), int64(len(input)))
	if err != errBadImageType {
		t.Errorf("expected errBadImageType, got %v", err)
	}
}

func TestValidateAndReadImage_RejectsOversizeDeclared(t *testing.T) {
	_, _, _, err := validateAndReadImage(bytes.NewReader(pngMagic), maxUploadSize+1)
	if err != errFileTooLarge {
		t.Errorf("expected errFileTooLarge for declared oversize, got %v", err)
	}
}

func TestValidateAndReadImage_RejectsOversizeContent(t *testing.T) {
	// declaredSize understates the real body — the bounded read must still reject.
	big := append(append([]byte{}, pngMagic...), bytes.Repeat([]byte{0x00}, maxUploadSize)...)
	_, _, _, err := validateAndReadImage(bytes.NewReader(big), 0)
	if err != errFileTooLarge {
		t.Errorf("expected errFileTooLarge for oversize content, got %v", err)
	}
}
