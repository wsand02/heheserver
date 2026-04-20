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
	"github.com/wsand02/heheserver/internal/vidthumb"
)

var vidThumbCache *cache.VidThumbCache

func VidThumbHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer hf.Close()
	if vidThumbCache == nil {
		vTC, err := cache.NewVidThumbCache()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		vidThumbCache = vTC
	}
	vtb, ok := vidThumbCache.Get(ctx)
	if ok {
		err = jpeg.Encode(w, vtb, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	head := make([]byte, 512)
	hf.Read(head)
	if !filetype.IsVideo(head) {
		http.Error(w, "Not a video", http.StatusUnsupportedMediaType)
		return
	}

	hfstat, _ := hf.Stat()
	fmt.Println(hfstat.Name())
	// a := path.Clean(path.Clean("/" + ctx)[1:])
	// a, _ = filepath.Localize(a)

	// taken from fs.Open()
	path := path.Clean("/" + ctx)[1:]
	if path == "" {
		path = "."
	}
	path, err = filepath.Localize(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dir := string(hfs.Root)
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, path)
	img, err := vidthumb.GenerateThumb(fullName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vidThumbCache.Set(ctx, img, int64(img.Bounds().Dx()*img.Bounds().Dy()))
	err = jpeg.Encode(w, img, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
