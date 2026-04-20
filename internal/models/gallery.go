package models

import (
	"path/filepath"
	"strings"
	"time"
)

type GalleryItem struct {
	Filename string
	IsDir    bool
	Path     string
	Size     int64
	ModTime  time.Time
}

func (gi *GalleryItem) SizeMB() float64 {
	return float64(gi.Size) / 1000000
}

func (gi *GalleryItem) IsImage() bool {
	ext := strings.ToLower(filepath.Ext(gi.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".svg":
		return true
	default:
		return false
	}
}

func (gi *GalleryItem) IsVideo() bool {
	ext := strings.ToLower(filepath.Ext(gi.Filename))
	switch ext {
	case ".mov", ".mp4", ".m4v", ".webm":
		return true
	default:
		return false
	}
}
func (gi *GalleryItem) IsAudio() bool {
	ext := strings.ToLower(filepath.Ext(gi.Filename))
	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a":
		return true
	default:
		return false
	}
}
func (gi *GalleryItem) IsResizable() bool {
	ext := strings.ToLower(filepath.Ext(gi.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true

	default:
		return false
	}
}

func (gi *GalleryItem) GetUrl() string {
	if gi.IsDir {
		return strings.Join([]string{"?path=", gi.Path, gi.Filename, "/"}, "")
	}
	return strings.Join([]string{"/fs", gi.Path, gi.Filename}, "")
}

func (gi *GalleryItem) GetPath() string {
	return strings.Join([]string{"/fs", gi.Path}, "")
}

func (gi *GalleryItem) GetPostLink() string {
	return strings.Join([]string{"/post/", "?path=", gi.Path, gi.Filename}, "")
}

func (gi *GalleryItem) GetResized() string {
	return strings.Join([]string{"/resize/", "?path=", gi.Path, gi.Filename}, "")
}
