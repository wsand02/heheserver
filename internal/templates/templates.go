package templates

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
)

//go:embed *.html
var templatesFS embed.FS

func sub(a, b int) int { return a - b }
func add(a, b int) int { return a + b }
func gt(a, b int) bool { return a > b }
func lt(a, b int) bool { return a < b }
func ge(a, b int) bool { return a >= b }
func le(a, b int) bool { return a <= b }
func pageURL(path string, page int) string {
	q := url.Values{}
	if path != "" {
		q.Set("path", path)
	}
	if page > 1 { // optionally omit p=1 for cleaner URL
		q.Set("p", fmt.Sprintf("%d", page))
	}
	if len(q) == 0 {
		return "?"
	}
	return "?" + q.Encode()
}

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
	"pageURL": pageURL,
	"seq":     seq}).ParseFS(templatesFS, "*.html"))

func RenderTemplate(w http.ResponseWriter, tmpl string, ctx any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
