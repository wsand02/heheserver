package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
)

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 64, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
}

func benchmarkJPEGImage(b testing.TB) (*image.RGBA, []byte) {
	b.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 300; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 64, 255})
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, nil); err != nil {
		b.Fatal(err)
	}
	return img, encoded.Bytes()
}

func BenchmarkResizeCacheHitReencode(b *testing.B) {
	img, _ := benchmarkJPEGImage(b)
	var output bytes.Buffer
	output.Grow(16 * 1024)
	b.ReportAllocs()
	for b.Loop() {
		output.Reset()
		if err := jpeg.Encode(&output, img, nil); err != nil {
			b.Fatal(err)
		}
	}
	if output.Len() == 0 {
		b.Fatal("expected encoded bytes")
	}
}

func BenchmarkResizeCacheHitEncoded(b *testing.B) {
	_, encoded := benchmarkJPEGImage(b)
	var output bytes.Buffer
	output.Grow(len(encoded))
	b.ReportAllocs()
	for b.Loop() {
		output.Reset()
		if _, err := output.Write(encoded); err != nil {
			b.Fatal(err)
		}
	}
	if output.Len() == 0 {
		b.Fatal("expected cached bytes")
	}
}

func TestResizeHandlerFallback(t *testing.T) {
	if err := cache.NewResizeCache(1); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "pic.png"), 600, 400)
	hfs := testHFS(t, dir) // also initializes the ignore cache
	cfg := &config.Config{FFmpegExists: false}

	// First request: generated via the pure-Go fallback and cached.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/resize/?path=/pic.png", nil)
	ResizeHandler(w, r, "/pic.png", hfs, cfg)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if _, err := jpeg.Decode(bytes.NewReader(w.Body.Bytes())); err != nil {
		t.Fatalf("response body is not a decodable JPEG: %v", err)
	}
	firstBody := append([]byte(nil), w.Body.Bytes()...)
	cache.GetResizeCache().Wait()
	cached, ok := cache.GetResizeCache().Get("/pic.png")
	if !ok || !bytes.Equal(cached.JPEG, firstBody) {
		t.Fatal("expected encoded JPEG in resize cache")
	}

	// Second request: served from cache.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/resize/?path=/pic.png", nil)
	ResizeHandler(w, r, "/pic.png", hfs, cfg)
	if w.Code != 200 {
		t.Fatalf("cached status = %d, want 200", w.Code)
	}
	if cc := w.Header().Get("Cache-Control"); cc == "" {
		t.Error("expected Cache-Control header on cached response")
	}
	if _, err := jpeg.Decode(bytes.NewReader(w.Body.Bytes())); err != nil {
		t.Fatalf("cached body is not a decodable JPEG: %v", err)
	}
	if !bytes.Equal(w.Body.Bytes(), firstBody) {
		t.Fatal("cached response body differs from the generated JPEG")
	}
}

func TestResizeHandlerNotAnImage(t *testing.T) {
	if err := cache.NewResizeCache(1); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("just some text, not an image at all"), 0644); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cfg := &config.Config{FFmpegExists: false}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/resize/?path=/note.txt", nil)
	ResizeHandler(w, r, "/note.txt", hfs, cfg)
	if w.Code != 415 {
		t.Errorf("status = %d, want 415", w.Code)
	}
}
