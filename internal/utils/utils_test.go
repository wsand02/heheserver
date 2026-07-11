package utils

import (
	"errors"
	"image"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCost(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	if got := GetCost(img); got != 300*200 {
		t.Errorf("GetCost() = %d, want %d", got, 300*200)
	}
}

func TestHttpLogErr(t *testing.T) {
	w := httptest.NewRecorder()
	HttpLogErr(w, errors.New("boom"), "something failed", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if body := w.Body.String(); body != "something failed\n" {
		t.Errorf("body = %q, want %q", body, "something failed\n")
	}
}

func TestFFmpegExists(t *testing.T) {
	// Result depends on the host; just assert it does not panic and returns a bool.
	_ = FFmpegExists()
}
