package handlers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
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

func ResizeHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, cfg *config.Config) {
	once.Do(func() {
		resizeCache, _ = cache.NewResizeCache()
	})

	cImg, ok := resizeCache.Get(ctx)
	w.Header().Add("Cache-Control", "private, max-age=86400")
	if ok {
		if cImg.Transparent {
			err := png.Encode(w, cImg.Image)
			if err != nil {
				utils.HttpLogErr(w, err, "Error encoding cached png", http.StatusInternalServerError)
				return
			}
			return
		}
		err := jpeg.Encode(w, cImg.Image, nil)
		if err != nil {
			utils.HttpLogErr(w, err, "Error encoding cached jpeg", http.StatusInternalServerError)
			return
		}
		return
	}

	hf, err := hfs.Open(ctx)
	if err != nil {
		utils.HttpLogErr(w, err, "Error opening file", http.StatusInternalServerError)
		return
	}
	defer hf.Close()

	var transparent bool
	head := make([]byte, 512)
	hf.Read(head)
	hf.Seek(0, io.SeekStart)
	if !filetype.IsImage(head) {
		utils.HttpLogErr(w, fmt.Errorf("Not an image"), "Not an image", http.StatusUnsupportedMediaType)
		return
	}

	var dst image.Image

	if cfg.FFmpegExists {
		fullName := filepath.Join(hfs.Root, ctx)
		fmt.Println(fullName)
		dst, err = resize.ResizeImage(fullName)
	} else {
		log.Println("Resizing using fallback")
		dst, err = resize.ResizeImageFallback(hf)
	}
	if err != nil {
		utils.HttpLogErr(w, err, "Error resizing image", http.StatusInternalServerError)
		return
	}

	resizeCache.Set(ctx, cache.ResizeCacheItem{
		Image:       dst,
		Transparent: transparent,
	}, utils.GetCost(dst))
	err = jpeg.Encode(w, dst, nil)
	if err != nil {
		utils.HttpLogErr(w, err, "Error encoding jpeg", http.StatusInternalServerError)
		return
	}
}
