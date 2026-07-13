package utils

import (
	"errors"
	"image"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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
	r := httptest.NewRequest(http.MethodGet, "/post/?path=/x.txt", nil)
	HttpLogErr(w, r, errors.New("boom"), "something failed", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if body := w.Body.String(); !strings.Contains(body, "something failed") {
		t.Errorf("body missing message %q: %s", "something failed", body)
	}
}

func TestFFmpegExists(t *testing.T) {
	// Result depends on the host; just assert it does not panic and returns a bool.
	_ = FFmpegExists()
}

func TestGetLocalIP(t *testing.T) {
	ip, err := GetLocalIP()
	if err != nil {
		// No network available in this environment; nothing more to assert.
		return
	}
	if net.ParseIP(ip) == nil {
		t.Errorf("GetLocalIP() = %q, not a valid IP", ip)
	}
}
