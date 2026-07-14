package models

import (
	"net/url"
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
	// Width/Height are the source image's pixel dimensions, filled in for image
	// items so the grid can reserve the right aspect ratio before the thumbnail
	// loads (avoiding masonry reflow). Zero when unknown (non-image, or an
	// undecodable format like svg).
	Width  int
	Height int
}

// SizeMB returns the size as Megabytes NOT MEBIBYTESDSDKFJK MAYBE MY:......
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
// TypeCategory classifies the item into a single filter bucket:
// "dir", "image", "video", "audio", or "other". It is the single source of
// truth for type-based filtering, built on the Is* predicates above.
func (gi *GalleryItem) TypeCategory() string {
	switch {
	case gi.IsDir:
		return "dir"
	case gi.IsImage():
		return "image"
	case gi.IsVideo():
		return "video"
	case gi.IsAudio():
		return "audio"
	default:
		return "other"
	}
}

// GalleryFilter narrows a directory listing. An empty field means "no
// constraint" for that dimension; active constraints combine with AND.
type GalleryFilter struct {
	Types map[string]bool // TypeCategory values to keep; empty = all
	Query string          // lowercased filename substring; "" = no filter
	Exts  map[string]bool // normalized ".png" extension keys; empty = all
}

// Active reports whether any filter dimension is set.
func (f GalleryFilter) Active() bool {
	return len(f.Types) > 0 || f.Query != "" || len(f.Exts) > 0
}

// Matches reports whether gi satisfies every active filter dimension.
func (f GalleryFilter) Matches(gi *GalleryItem) bool {
	if len(f.Types) > 0 && !f.Types[gi.TypeCategory()] {
		return false
	}
	if f.Query != "" && !strings.Contains(strings.ToLower(gi.Filename), f.Query) {
		return false
	}
	if len(f.Exts) > 0 && !f.Exts[strings.ToLower(filepath.Ext(gi.Filename))] {
		return false
	}
	return true
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

// escapeQueryPath percent-encodes each path segment for use as a ?path= query
// value, leaving the separating slashes intact so the path stays readable.
func escapeQueryPath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = url.QueryEscape(s)
	}
	return strings.Join(segs, "/")
}

// escapeURLPath percent-encodes each path segment for use in a /fs/... URL path,
// leaving the separating slashes intact.
func escapeURLPath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return strings.Join(segs, "/")
}

func (gi *GalleryItem) GetUrl() string {
	if gi.IsDir {
		return strings.Join([]string{"?path=", escapeQueryPath(gi.Path + gi.Filename), "/"}, "")
	}
	return strings.Join([]string{"/fs", escapeURLPath(gi.Path + gi.Filename)}, "")
}

func (gi *GalleryItem) GetPath() string {
	return strings.Join([]string{"/fs", escapeURLPath(gi.Path)}, "")
}

func (gi *GalleryItem) GetPostLink() string {
	return strings.Join([]string{"/post/", "?path=", escapeQueryPath(gi.Path + gi.Filename)}, "")
}

func (gi *GalleryItem) GetResized() string {
	return strings.Join([]string{"/resize/", "?path=", escapeQueryPath(gi.Path + gi.Filename)}, "")
}

func (gi *GalleryItem) GetVidThumb() string {
	return strings.Join([]string{"/vidthumb/", "?path=", escapeQueryPath(gi.Path + gi.Filename)}, "")
}
