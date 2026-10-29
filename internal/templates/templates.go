// Package templates renders small standalone error pages for file and thumbnail requests.
package templates

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed error.html
var templateFiles embed.FS
var templates = template.Must(template.ParseFS(templateFiles, "error.html"))

type errorContext struct {
	Code    int
	Status  string
	Message string
}

// RenderError writes a standalone error page with the given status code
// and detail message.
func RenderError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	ctx := errorContext{Code: code, Status: http.StatusText(code), Message: msg}
	if err := templates.ExecuteTemplate(w, "error.html", ctx); err != nil {
		log.Printf("render error page (%d): %v", code, err)
	}
}
