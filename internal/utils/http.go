package utils

import (
	"errors"
	"io/fs"
	"log"
	"net/http"

	"github.com/wsand02/heheserver/internal/templates"
)

// HttpLogErr logs the failure with request context (status, method, requested
// URL) and the underlying error, then renders a styled error page to the client.
func HttpLogErr(w http.ResponseWriter, r *http.Request, err error, msg string, code int) {
	log.Printf("[%d] %s %s — %s: %v", code, r.Method, r.URL.RequestURI(), msg, err)
	templates.RenderError(w, code, msg)
}

// StatusForErr maps a filesystem error to an HTTP status code.
func StatusForErr(err error) int {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, fs.ErrPermission):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
