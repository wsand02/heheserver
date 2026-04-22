package handlers

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/h2non/filetype"
	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/resize"
	"github.com/wsand02/heheserver/internal/utils"
)

var once sync.Once

// since the thumbnails are so small we can just cache them in memory
var resizeCache *cache.ResizeCache

func ResizeHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	once.Do(func() {
		resizeCache, _ = cache.NewResizeCache()
	})

	cImg, ok := resizeCache.Get(ctx)
	w.Header().Add("Cache-Control", "private, max-age=86400")
	if ok {
		if cImg.Transparent {
			err := png.Encode(w, cImg.Image)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		err := jpeg.Encode(w, cImg.Image, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer hf.Close()

	var transparent bool
	head := make([]byte, 512)
	hf.Read(head)
	if !filetype.IsImage(head) {
		http.Error(w, "Not an image", http.StatusUnsupportedMediaType)
		return
	}

	fullName := filepath.Join(hfs.Root, ctx)
	fmt.Println(fullName)
	dst, err := resize.ResizeImage(fullName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resizeCache.Set(ctx, cache.ResizeCacheItem{
		Image:       dst,
		Transparent: transparent,
	}, utils.GetCost(dst))
	err = jpeg.Encode(w, dst, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
