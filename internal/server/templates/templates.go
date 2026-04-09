package templates

import (
	"embed"
	"net/http"
	"text/template"
)

//go:embed *.html
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "*.html"))

func RenderTemplate(w http.ResponseWriter, tmpl string, ctx any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
