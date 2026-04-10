package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/templates"
)

type GalleryContext struct {
	Items []GalleryItem
}

var imgExt = []string{".jpg", ".jpeg", ".png", ".webp", ".svg"}
var vidExt = []string{".mov", ".mp4", ".m4v"}
var audExt = []string{".mp3", ".wav", ".ogg", ".m4a"}

type GalleryItem struct {
	Filename string
	IsDir    bool
	Path     string
}

func (gi *GalleryItem) IsImage() bool {
	for _, ie := range imgExt {
		if filepath.Ext(strings.ToLower(gi.Filename)) == ie {
			fmt.Println(filepath.Ext(strings.ToLower(gi.Filename)))
			return true
		}
	}
	return false
}

func (gi *GalleryItem) IsVideo() bool {
	for _, ie := range vidExt {
		if filepath.Ext(strings.ToLower(gi.Filename)) == ie {
			return true
		}
	}
	return false
}
func (gi *GalleryItem) IsAudio() bool {
	for _, ie := range audExt {
		if filepath.Ext(strings.ToLower(gi.Filename)) == ie {
			return true
		}
	}
	return false
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
