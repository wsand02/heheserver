package templates

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
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
		// Rendering may have already written a partial body (and a 200 header),
		// so we can't reliably swap in a styled error page here — just log it.
		log.Printf("render template %q: %v", tmpl, err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

type errorContext struct {
	Code    int
	Status  string
	Message string
}

// RenderError writes a glacialwisp-themed error page with the given status code
// and detail message.
func RenderError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	ctx := errorContext{Code: code, Status: http.StatusText(code), Message: msg}
	if err := templates.ExecuteTemplate(w, "error.html", ctx); err != nil {
		log.Printf("render error page (%d): %v", code, err)
	}
}
