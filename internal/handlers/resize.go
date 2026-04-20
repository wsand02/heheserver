package handlers

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"

	"github.com/h2non/filetype"
	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/resize"
	"github.com/wsand02/heheserver/internal/utils"
)

// since the thumbnails are so small we can just cache them in memory
var resizeCache *cache.ResizeCache

func ResizeHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer hf.Close()
	if resizeCache == nil {
		rC, err := cache.NewResizeCache()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		resizeCache = rC
	}
	cImg, ok := resizeCache.Get(ctx)
	w.Header().Add("Cache-Control", "private, max-age=86400")
	if ok {
		if cImg.Transparent {
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
	head := make([]byte, 512)
	hf.Read(head)
	hf.Seek(0, io.SeekStart)
	if filetype.IsMIME(head, "image/jpeg") {
		src, err = jpeg.Decode(hf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transparent = false
	} else if filetype.IsMIME(head, "image/png") {
		src, err = png.Decode(hf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transparent = true
	} else {
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
	// fmt.Print(ctx)
	// fmt.Print(": ")
	// fmt.Println(cost)

	resizeCache.Set(ctx, cache.ResizeCacheItem{
		Image:       dst,
		Transparent: transparent,
	}, utils.GetCost(dst))
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
