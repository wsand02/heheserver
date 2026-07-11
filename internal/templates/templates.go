package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/wsand02/heheserver/internal/version"
)

//go:embed *.html
var templatesFS embed.FS

//go:embed static/*.css
//go:embed static/*.js
var staticFS embed.FS

// StaticHandler serves the embedded static assets (glacialwisp CSS plus the
// masonry/imagesLoaded/infinite-scroll JS) under /static/.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err) // embedded FS is known at build time; this cannot fail
	}
	return http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
}

func sub(a, b int) int { return a - b }
func add(a, b int) int { return a + b }
func gt(a, b int) bool { return a > b }
func lt(a, b int) bool { return a < b }
func ge(a, b int) bool { return a >= b }
func le(a, b int) bool { return a <= b }

// seq returns a slice of ints from start to end inclusive
func seq(start, end int) []int {
	if start > end {
		return []int{}
	}
	s := make([]int, end-start+1)
	for i := start; i <= end; i++ {
		s[i-start] = i
	}
	return s
}

var templates = template.Must(template.New("").Funcs(template.FuncMap{"sub": sub,
	"add":     add,
	"gt":      gt,
	"lt":      lt,
	"ge":      ge,
	"le":      le,
	"seq":     seq,
	"version": version.GetVersion}).ParseFS(templatesFS, "*.html"))

func RenderTemplate(w http.ResponseWriter, tmpl string, ctx any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
