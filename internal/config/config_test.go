package config

import (
	"path/filepath"
	"testing"
)

func TestNewConfigValid(t *testing.T) {
	dir := t.TempDir()
	cfg, err := NewConfig(3400, 64, true, false, dir, "0.0.0.0", 16, 1000, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 3400 || cfg.Split != 64 || !cfg.Gallery || cfg.Resize {
		t.Errorf("config fields not populated as expected: %+v", cfg)
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
		if _, err := NewConfig(3400, split, false, false, dir, "0.0.0.0", 16, 1000, 1000); err == nil {
			t.Errorf("split %d: expected error, got nil", split)
		}
	}
}

func TestNewConfigMissingDirectory(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if _, err := NewConfig(3400, 64, false, false, missing, "0.0.0.0", 16, 1000, 1000); err == nil {
		t.Error("expected error for missing directory, got nil")
	}
}

func TestGetAddress(t *testing.T) {
	c := &Config{Host: "127.0.0.1", Port: 8080}
	if got := c.GetAddress(); got != "127.0.0.1:8080" {
		t.Errorf("GetAddress() = %q, want %q", got, "127.0.0.1:8080")
	}
}
