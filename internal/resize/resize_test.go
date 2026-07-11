package resize

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wsand02/heheserver/internal/utils"
)

// makePNG returns a w x h PNG encoded into a buffer.
func makePNG(t *testing.T, w, h int) *bytes.Buffer {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestResizeImageFallback(t *testing.T) {
	buf := makePNG(t, 600, 400)
	dst, err := ResizeImageFallback(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.Bounds().Dx() != width {
		t.Errorf("width = %d, want %d", dst.Bounds().Dx(), width)
	}
	// 600x400 has ratio 1.5, so target height = width / 1.5 = 200.
	if dst.Bounds().Dy() != 200 {
		t.Errorf("height = %d, want 200", dst.Bounds().Dy())
	}
}

func TestResizeImageFallbackDecodeError(t *testing.T) {
	if _, err := ResizeImageFallback(strings.NewReader("not an image")); err == nil {
		t.Error("expected decode error for non-image input, got nil")
	}
}

func TestResizeImageFFmpeg(t *testing.T) {
	if !utils.FFmpegExists() {
		t.Skip("ffmpeg not on PATH")
	}
	// Write a valid PNG to disk for ffmpeg to read. Only feed a decodable image:
	// ResizeImage log.Fatals on a decode error.
	path := filepath.Join(t.TempDir(), "src.png")
	if err := os.WriteFile(path, makePNG(t, 600, 400).Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	img, err := ResizeImage(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Fatal("expected a resized image, got nil")
	}
	if img.Bounds().Dx() != width {
		t.Errorf("width = %d, want %d", img.Bounds().Dx(), width)
	}
}
