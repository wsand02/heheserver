package handlers

import (
	"net/http"

	"github.com/wsand02/heheserver/internal/server/templates"
)

func GalleryHandler(w http.ResponseWriter, r *http.Request) {
	templates.RenderTemplate(w, "gallery")
}
