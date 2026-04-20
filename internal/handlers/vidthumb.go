package handlers

import (
	"fmt"
	"image/jpeg"
	"net/http"
	"path"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/vidthumb"
)

func VidThumbHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer hf.Close()
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
	jpeg.Encode(w, img, nil)
}
