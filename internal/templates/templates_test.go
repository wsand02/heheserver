package templates

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderError(t *testing.T) {
	rec := httptest.NewRecorder()
	RenderError(rec, http.StatusNotFound, "open file")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{"404", http.StatusText(http.StatusNotFound), "open file", "Back to gallery"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q\n%s", want, body)
		}
	}
}

func TestErrorEscapesMessage(t *testing.T) {
	w := httptest.NewRecorder()
	RenderError(w, 400, "<script>alert(1)</script>")
	if strings.Contains(w.Body.String(), "<script>") {
		t.Fatal("unescaped message")
	}
}
