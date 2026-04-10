package handlers

import (
	"net/http"
	"strings"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/config"
	"github.com/wsand02/heheserver/internal/server/models"
	"github.com/wsand02/heheserver/internal/server/templates"
)

type GalleryContext struct {
	Items  []models.GalleryItem
	Resize bool
	Path   string
}

type Breadcrumb struct {
}

func (gc *GalleryContext) GetBreadcrumbs() []string {
	return strings.Split(gc.Path, "/")

}

func (gc *GalleryContext) BreadcrumbToUrl(i int) string {
	crumbs := gc.GetBreadcrumbs()[1 : i+1]
	pcrumbs := []string{"?path="}
	pcrumbs = append(pcrumbs, crumbs...)
	url := strings.Join(pcrumbs, "/")
	urla := strings.Join([]string{url, "/"}, "")
	return urla
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
	gc.Path = ctx

	templates.RenderTemplate(w, "list", &gc)
}

type PostContext struct {
	models.GalleryItem
}

func PostHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hfstat, err := hf.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if hfstat.IsDir() {
		http.Error(w, "This is a directory", http.StatusUnsupportedMediaType)
		return
	}
	templates.RenderTemplate(w, "post", &PostContext{models.GalleryItem{Filename: hfstat.Name(), IsDir: hfstat.IsDir(), Path: ctx}}) // oh well
}
