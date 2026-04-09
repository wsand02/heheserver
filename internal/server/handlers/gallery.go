package handlers

import (
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/server/templates"
)

type Item struct {
	URL string
}

type Context struct {
	Items []Item
}

func BuildContext(hfs fs.HeheFS) *Context {
	hf, err := hfs.Open(hfs.Root)
	if err != nil {
		log.Fatal(err)
	}
	files, err := hf.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}
	var items []Item
	for i := 0; i < len(files); i++ {
		newitem := Item{URL: files[i].Name()}
		items = append(items, newitem)
	}
	return &Context{
		Items: items,
	}
}

func GalleryHandler(w http.ResponseWriter, r *http.Request, ctx any) {
	ctx, ok := ctx.(*Context)
	if !ok {
		http.Error(w, "Invalid context", http.StatusInternalServerError)
		return
	}
	templates.RenderTemplate(w, "list", ctx)
}
