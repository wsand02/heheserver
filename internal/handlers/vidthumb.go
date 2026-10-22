package handlers

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"net/http"
	"path"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/thumbgen"
	"github.com/wsand02/heheserver/internal/utils"
	"github.com/wsand02/heheserver/internal/vidthumb"
)

func VidThumbHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		utils.HttpLogErr(w, r, err, "open file", utils.StatusForErr(err))
		return
	}
	defer hf.Close()

	w.Header().Add("Cache-Control", "private, max-age=86400")
	vtb, ok := cache.GetVidThumbCache().Get(ctx)
	if ok {
		writeVideoThumbnail(w, r, vtb.JPEG)
		return
	}
	head := make([]byte, 512)
	hf.Read(head)
	if !filetype.IsVideo(head) {
		utils.HttpLogErr(w, r, fmt.Errorf("not a video"), "not a video", http.StatusUnsupportedMediaType)
		return
	}

	// taken from fs.Open()
	path := path.Clean("/" + ctx)[1:]
	if path == "" {
		path = "."
	}
	path, err = filepath.Localize(path)
	if err != nil {
		utils.HttpLogErr(w, r, err, "localize path", http.StatusInternalServerError)
		return
	}
	dir := string(hfs.Root)
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, path)
	generated, err := thumbgen.Do("video:"+ctx, func() (any, error) {
		if cached, ok := cache.GetVidThumbCache().Get(ctx); ok {
			return cached, nil
		}
		img, generateErr := vidthumb.GenerateThumb(fullName)
		if generateErr != nil {
			return nil, generateErr
		}
		var encoded bytes.Buffer
		if encodeErr := jpeg.Encode(&encoded, img, nil); encodeErr != nil {
			return nil, encodeErr
		}
		item := cache.VidThumbCacheItem{Image: img, JPEG: encoded.Bytes()}
		cache.GetVidThumbCache().Set(ctx, item, utils.GetCost(img)+int64(len(item.JPEG)))
		return item, nil
	})
	if err != nil {
		utils.HttpLogErr(w, r, err, "generate thumbnail", http.StatusInternalServerError)
		return
	}
	item := generated.(cache.VidThumbCacheItem)
	writeVideoThumbnail(w, r, item.JPEG)
}

func writeVideoThumbnail(w http.ResponseWriter, r *http.Request, jpegBytes []byte) {
	if _, err := w.Write(jpegBytes); err != nil {
		utils.HttpLogWriteErr(r, err, "write thumbnail", http.StatusInternalServerError)
	}
}
