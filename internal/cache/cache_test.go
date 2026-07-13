package cache

import (
	"image"
	"slices"
	"testing"
)

func TestSizeToNCMB(t *testing.T) {
	bytes, nc := sizeToNCMB(16)
	if bytes != 16*1024*1024 {
		t.Errorf("bytes = %d, want %d", bytes, 16*1024*1024)
	}
	if nc != 16*10000 {
		t.Errorf("numCounters = %d, want %d", nc, 16*10000)
	}
}

func TestIgnoreCacheRoundtrip(t *testing.T) {
	if err := NewIgnoreCache(1); err != nil {
		t.Fatal(err)
	}
	lines := []string{"*.tmp", "!keep.tmp"}
	c := GetIgnoreCache()
	c.Set("key", lines, 1)
	c.Wait() // ristretto applies Set asynchronously

	got, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cached ignore value, got miss")
	}
	if !slices.Equal(got, lines) {
		t.Error("cached ignore value differs from stored value")
	}
}

func TestResizeCacheRoundtrip(t *testing.T) {
	if err := NewResizeCache(1); err != nil {
		t.Fatal(err)
	}
	item := ResizeCacheItem{Image: image.NewRGBA(image.Rect(0, 0, 2, 2)), Transparent: true}
	c := GetResizeCache()
	c.Set("key", item, 4)
	c.Wait()

	got, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cached resize item, got miss")
	}
	if !got.Transparent || got.Image == nil {
		t.Errorf("cached resize item not preserved: %+v", got)
	}
}

func TestVidThumbCacheRoundtrip(t *testing.T) {
	if err := NewVidThumbCache(1); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	c := GetVidThumbCache()
	c.Set("key", img, 4)
	c.Wait()

	got, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cached thumbnail, got miss")
	}
	if got == nil {
		t.Error("cached thumbnail is nil")
	}
}
