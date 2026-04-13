package handlers

import (
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/resize"
	"github.com/wsand02/heheserver/internal/server/config"
)

type resizeCacheItem struct {
	image.Image
	transparent bool
}

// since the thumbnails are so small we can just cache them in memory
var (
	resizeCache = map[string]*resizeCacheItem{}
	cacheMu     sync.RWMutex
)

func ResizeHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer hf.Close()
	hfstat, err := hf.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cacheMu.RLock()
	cImg, ok := resizeCache[ctx]
	cacheMu.RUnlock()
	w.Header().Add("Cache-Control", "private, max-age=86400")
	if ok {
		if cImg.transparent {
			err = png.Encode(w, cImg.Image)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		err = jpeg.Encode(w, cImg.Image, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	var src image.Image
	var transparent bool
	switch strings.ToLower(filepath.Ext(hfstat.Name())) {
	case ".jpg", ".jpeg":
		src, err = jpeg.Decode(hf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transparent = false

	case ".png":
		src, err = png.Decode(hf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transparent = true
	default:
		http.Error(w, "Invalid file", http.StatusUnsupportedMediaType)
		return
	}
	if src == nil {
		http.Error(w, "Invalid file", http.StatusUnsupportedMediaType)
		return
	}
	dst := resize.ResizeImage(src)
	if dst == nil {
		http.Error(w, "Invalid file", http.StatusUnsupportedMediaType)
		return
	}
	cacheMu.Lock()
	resizeCache[ctx] = &resizeCacheItem{
		dst,
		transparent,
	}
	cacheMu.Unlock()
	if transparent {
		err = png.Encode(w, dst)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	err = jpeg.Encode(w, dst, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
