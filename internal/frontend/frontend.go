// Package frontend serves the committed production frontend; building Go never needs Node.
package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var files embed.FS

func Handler() http.Handler {
	sub, _ := fs.Sub(files, "dist")
	assets := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/" || r.URL.Path == "/post/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			b, _ := files.ReadFile("dist/index.html")
			w.Write(b)
			return
		}
		if len(r.URL.Path) > len("/assets/") && r.URL.Path[:len("/assets/")] == "/assets/" {
			assets.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}
