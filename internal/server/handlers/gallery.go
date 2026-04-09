package handlers

import (
	"net/http"
	"strings"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/templates"
)

type GalleryContext struct {
	Items []GalleryItem
}

type GalleryItem struct {
	Filename string
	IsDir    bool
	Path     string
}

func (gi *GalleryItem) GetUrl() string {
	if gi.IsDir {
		return strings.Join([]string{"?path=", gi.Path, gi.Filename, "/"}, "")
	}
	return strings.Join([]string{"/fs", gi.Path, gi.Filename}, "")
}

func GalleryHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dirlis, err := hf.Readdir(-1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var gc GalleryContext
	for i := 0; i < len(dirlis); i++ {
		gc.Items = append(gc.Items, GalleryItem{Filename: dirlis[i].Name(), IsDir: dirlis[i].IsDir(), Path: ctx})
	}

	templates.RenderTemplate(w, "list", gc)
}
