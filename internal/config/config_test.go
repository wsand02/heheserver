package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfigValid(t *testing.T) {
	dir := t.TempDir()
	cfg, err := NewConfig(3400, 64, true, false, false, dir, "0.0.0.0", 16, 1000, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 3400 || cfg.Split != 64 || !cfg.Gallery || cfg.Resize {
		t.Errorf("config fields not populated as expected: %+v", cfg)
	}
	if cfg.NoInfiniteScroll {
		t.Errorf("NoInfiniteScroll should default to false when not set: %+v", cfg)
	}
	if cfg.Directory != dir || cfg.Host != "0.0.0.0" {
		t.Errorf("directory/host not populated: %+v", cfg)
	}
	if cfg.IgnoreCacheSize != 16 || cfg.ResizeCacheSize != 1000 || cfg.VidThumbCacheSize != 1000 {
		t.Errorf("cache sizes not populated: %+v", cfg)
	}
}

func TestNewConfigInvalidSplit(t *testing.T) {
	dir := t.TempDir()
	for _, split := range []int{0, -1} {
		if _, err := NewConfig(3400, split, false, false, false, dir, "0.0.0.0", 16, 1000, 1000); err == nil {
			t.Errorf("split %d: expected error, got nil", split)
		}
	}
}

func TestNewConfigMissingDirectory(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := NewConfig(3400, 64, false, false, false, missing, "0.0.0.0", 16, 1000, 1000); err == nil {
		t.Error("expected error for missing directory, got nil")
	}
}

func TestNewConfigNoInfiniteScroll(t *testing.T) {
	dir := t.TempDir()
	cfg, err := NewConfig(3400, 64, true, false, true, dir, "0.0.0.0", 16, 1000, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.NoInfiniteScroll {
		t.Errorf("NoInfiniteScroll not threaded into Config: %+v", cfg)
	}
}

func TestGetAddress(t *testing.T) {
	c := &Config{Host: "127.0.0.1", Port: 8080}
	if got := c.GetAddress(); got != "127.0.0.1:8080" {
		t.Errorf("GetAddress() = %q, want %q", got, "127.0.0.1:8080")
	}
}

func TestGetDisplayURLsLoopback(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "localhost", "::1"} {
		c := &Config{Host: host, Port: 8080}
		urls := c.GetDisplayURLs()
		if len(urls) != 1 {
			t.Fatalf("host %q: got %d urls, want 1: %v", host, len(urls), urls)
		}
		if urls[0] != "http://localhost:8080" {
			t.Errorf("host %q: urls[0] = %q, want %q", host, urls[0], "http://localhost:8080")
		}
	}
}

func TestGetDisplayURLsBindAll(t *testing.T) {
	c := &Config{Host: "0.0.0.0", Port: 8080}
	urls := c.GetDisplayURLs()
	if len(urls) < 1 || urls[0] != "http://localhost:8080" {
		t.Fatalf("urls[0] = %v, want first entry http://localhost:8080", urls)
	}
	if len(urls) == 2 {
		if !strings.HasPrefix(urls[1], "http://") || !strings.HasSuffix(urls[1], ":8080") {
			t.Errorf("urls[1] = %q, want format http://<ip>:8080", urls[1])
		}
	} else if len(urls) != 1 {
		t.Fatalf("got %d urls, want 1 or 2: %v", len(urls), urls)
	}
}

func TestGetDisplayURLsExplicitHost(t *testing.T) {
	c := &Config{Host: "192.168.1.5", Port: 8080}
	urls := c.GetDisplayURLs()
	want := []string{"http://localhost:8080", "http://192.168.1.5:8080"}
	if fmt.Sprint(urls) != fmt.Sprint(want) {
		t.Errorf("GetDisplayURLs() = %v, want %v", urls, want)
	}
}
