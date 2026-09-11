package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/resize"
	"github.com/wsand02/heheserver/internal/utils"
)

func ResizeHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, cfg *config.Config) {

	cImg, ok := cache.GetResizeCache().Get(ctx)
	w.Header().Add("Cache-Control", "private, max-age=86400")
	if ok {
		if cImg.Transparent {
			err := png.Encode(w, cImg.Image)
			if err != nil {
				utils.HttpLogErr(w, r, err, "encode cached png", http.StatusInternalServerError)
				return
			}
			return
		}
		if len(cImg.JPEG) > 0 {
			if _, err := w.Write(cImg.JPEG); err != nil {
				utils.HttpLogErr(w, r, err, "write cached jpeg", http.StatusInternalServerError)
			}
			return
		}
		err := jpeg.Encode(w, cImg.Image, nil)
		if err != nil {
			utils.HttpLogErr(w, r, err, "encode cached jpeg", http.StatusInternalServerError)
			return
		}
		return
	}

	hf, err := hfs.Open(ctx)
	if err != nil {
		utils.HttpLogErr(w, r, err, "open file", utils.StatusForErr(err))
		return
	}
	defer hf.Close()

	var transparent bool
	head := make([]byte, 512)
	hf.Read(head)
	hf.Seek(0, io.SeekStart)
	if !filetype.IsImage(head) {
		utils.HttpLogErr(w, r, fmt.Errorf("not an image"), "not an image", http.StatusUnsupportedMediaType)
		return
	}

	var dst image.Image

	if cfg.FFmpegExists {
		fullName := filepath.Join(hfs.Root, ctx)
		dst, err = resize.ResizeImage(fullName)
	} else {
		dst, err = resize.ResizeImageFallback(hf)
	}
	if err != nil {
		utils.HttpLogErr(w, r, err, "resize image", http.StatusInternalServerError)
		return
	}

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, dst, nil); err != nil {
		utils.HttpLogErr(w, r, err, "encode jpeg", http.StatusInternalServerError)
		return
	}
	encodedJPEG := encoded.Bytes()
	cache.GetResizeCache().Set(ctx, cache.ResizeCacheItem{
		Image:       dst,
		Transparent: transparent,
		JPEG:        encodedJPEG,
	}, utils.GetCost(dst)+int64(len(encodedJPEG)))
	if _, err := w.Write(encodedJPEG); err != nil {
		utils.HttpLogErr(w, r, err, "write jpeg", http.StatusInternalServerError)
	}
}
