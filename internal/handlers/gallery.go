package handlers

import (
	"net/http"
	"strconv"
	"strings"

	iofs "io/fs"

	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/models"
	"github.com/wsand02/heheserver/internal/templates"
)

type GalleryContext struct {
	Items       []models.GalleryItem
	Resize      bool
	Path        string
	CurrentPage int
	MaxPage     int
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
	defer hf.Close()
	dirlis, err := hf.Readdir(-1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var divided [][]iofs.FileInfo
	for i := 0; i < len(dirlis); i += 12 {
		divided = append(divided, dirlis[i:min(i+12, len(dirlis))])
	}
	q := r.URL.Query().Get("p")
	if q == "" {
		q = "1"
	}
	pid, err := strconv.Atoi(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pid -= 1
	if pid > len(divided) || pid < 0 {
		http.Error(w, "p out of range", http.StatusBadRequest)
		return
	}
	var gc GalleryContext
	for _, item := range divided[pid] {
		gc.Items = append(gc.Items, models.GalleryItem{Filename: item.Name(), IsDir: item.IsDir(), Size: item.Size(), ModTime: item.ModTime(), Path: ctx})
	}
	gc.Resize = config.Resize
	gc.Path = ctx
	gc.CurrentPage = pid + 1
	gc.MaxPage = len(divided)

	templates.RenderTemplate(w, "list", &gc)
}

type PostContext struct {
	models.GalleryItem
}

func PostHandler(w http.ResponseWriter, r *http.Request, ctx string, hfs *fs.HeheFS, config *config.Config) {
	hf, err := hfs.Open(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError) // might leak path but for the intended deployment scenario, does not matter
		return
	}
	defer hf.Close()
	hfstat, err := hf.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if hfstat.IsDir() {
		http.Error(w, "This is a directory", http.StatusUnsupportedMediaType)
		return
	}
	templates.RenderTemplate(w, "post", &PostContext{models.GalleryItem{Filename: hfstat.Name(), IsDir: hfstat.IsDir(), Path: ctx, Size: hfstat.Size(), ModTime: hfstat.ModTime()}}) // oh well
}
