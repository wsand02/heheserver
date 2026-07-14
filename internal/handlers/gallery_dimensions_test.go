package handlers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/models"
)

// writePNG (a real PNG of the given size) is defined in resize_test.go.

// TestProbeImageDimensions verifies the header probe fills Width/Height for a
// real image, leaves non-images and undecodable files at zero, and that a
// second pass is served from the dimension cache.
func TestProbeImageDimensions(t *testing.T) {
	if err := cache.NewDimensionCache(); err != nil {
		t.Fatal(err)
	}
	// hfs.Open consults the ignore cache; it must be initialized.
	if err := cache.NewIgnoreCache(16); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "pic.png"), 640, 360)
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A .svg is classified as an image but has no decodable raster header.
	if err := os.WriteFile(filepath.Join(dir, "vec.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	hfs, ok := fs.Dir(dir).(*fs.HeheFS)
	if !ok {
		t.Fatal("fs.Dir did not return *fs.HeheFS")
	}

	items := []models.GalleryItem{
		{Filename: "pic.png", Path: "/"},
		{Filename: "note.txt", Path: "/"},
		{Filename: "vec.svg", Path: "/"},
	}
	probeImageDimensions(hfs, items)

	if items[0].Width != 640 || items[0].Height != 360 {
		t.Errorf("pic.png: got %dx%d, want 640x360", items[0].Width, items[0].Height)
	}
	if items[1].Width != 0 || items[1].Height != 0 {
		t.Errorf("note.txt (non-image): expected zero dimensions, got %dx%d", items[1].Width, items[1].Height)
	}
	if items[2].Width != 0 || items[2].Height != 0 {
		t.Errorf("vec.svg (undecodable): expected zero dimensions, got %dx%d", items[2].Width, items[2].Height)
	}

	// The dimensions must now be cached under the file's path (ristretto writes
	// are async, so flush the buffer before asserting).
	cache.GetDimensionCache().Wait()
	if pt, ok := cache.GetDimensionCache().Get("/pic.png"); !ok || pt.X != 640 || pt.Y != 360 {
		t.Errorf("expected cached 640x360 for /pic.png, got %v (ok=%v)", pt, ok)
	}
}
