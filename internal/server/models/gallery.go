package models

import (
	"path/filepath"
	"strings"
)

type GalleryItem struct {
	Filename string
	IsDir    bool
	Path     string
}

type GalleryContext struct {
	Items []GalleryItem
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

func (gi *GalleryItem) GetUrl() string {
	if gi.IsDir {
		return strings.Join([]string{"?path=", gi.Path, gi.Filename, "/"}, "")
	}
	return strings.Join([]string{"/fs", gi.Path, gi.Filename}, "")
}
