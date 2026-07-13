package handlers

import (
	"fmt"
	"image/jpeg"
	"net/http"
	"path"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
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
		err = jpeg.Encode(w, vtb, nil)
		if err != nil {
			utils.HttpLogErr(w, r, err, "encode cached thumbnail", http.StatusInternalServerError)
			return
		}
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
	img, err := vidthumb.GenerateThumb(fullName)
	if err != nil {
		utils.HttpLogErr(w, r, err, "generate thumbnail", http.StatusInternalServerError)
		return
	}
	cache.GetVidThumbCache().Set(ctx, img, utils.GetCost(img))
	err = jpeg.Encode(w, img, nil)
	if err != nil {
		utils.HttpLogErr(w, r, err, "encode thumbnail", http.StatusInternalServerError)
		return
	}
}
