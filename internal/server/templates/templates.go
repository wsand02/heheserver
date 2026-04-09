package templates

import (
	"embed"
	"net/http"
	"text/template"
)

//go:embed *.html
var templatesFS embed.FS

var templates = template.Must(template.ParseFS(templatesFS, "*.html"))

func RenderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
