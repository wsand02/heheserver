package templates

import (
	"embed"
	"text/template"
)

//go:embed *.html
var templates embed.FS

func LoadTemplate(pattern string) (*template.Template, error) {
	return template.ParseFS(templates, pattern)
}
