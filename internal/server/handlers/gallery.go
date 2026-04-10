package handlers

import (
	"net/http"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/config"
	"github.com/wsand02/heheserver/internal/server/models"
	"github.com/wsand02/heheserver/internal/server/templates"
)

type GalleryContext struct {
	Items  []models.GalleryItem
	Resize bool
}

func GalleryHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
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
		gc.Items = append(gc.Items, models.GalleryItem{Filename: dirlis[i].Name(), IsDir: dirlis[i].IsDir(), Path: ctx})
	}
	gc.Resize = config.Resize

	templates.RenderTemplate(w, "list", gc)
}
