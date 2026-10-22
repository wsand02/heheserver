package handlers

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/utils"
)

type failingResponseWriter struct {
	header           http.Header
	writeHeaderCalls int
	writeCalls       int
}

func (w *failingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingResponseWriter) WriteHeader(int) {
	w.writeHeaderCalls++
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	w.writeCalls++
	return 0, errors.New("client disconnected")
}

func makeTestVideo(t *testing.T, path string) {
	t.Helper()
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", "testsrc=duration=2:size=64x64:rate=10",
		"-t", "2", "-an", "-c:v", "mpeg4", "-pix_fmt", "yuv420p", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to generate test video: %v\n%s", err, out)
	}
}

func TestVidThumbHandler(t *testing.T) {
	if !utils.FFmpegExists() {
		t.Skip("ffmpeg not on PATH")
	}
	if err := cache.NewVidThumbCache(1); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	makeTestVideo(t, filepath.Join(dir, "clip.mp4"))
	hfs := testHFS(t, dir) // also initializes the ignore cache
	cfg := &config.Config{FFmpegExists: true}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/vidthumb/?path=/clip.mp4", nil)
	VidThumbHandler(w, r, "/clip.mp4", hfs, cfg)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if _, err := jpeg.Decode(bytes.NewReader(w.Body.Bytes())); err != nil {
		t.Fatalf("response body is not a decodable JPEG: %v", err)
	}
}

func TestVidThumbHandlerNotAVideo(t *testing.T) {
	if !utils.FFmpegExists() {
		t.Skip("ffmpeg not on PATH")
	}
	if err := cache.NewVidThumbCache(1); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("not a video"), 0644); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cfg := &config.Config{FFmpegExists: true}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/vidthumb/?path=/note.txt", nil)
	VidThumbHandler(w, r, "/note.txt", hfs, cfg)
	if w.Code != 415 {
		t.Errorf("status = %d, want 415", w.Code)
	}
}

func TestVidThumbHandlerWriteFailureDoesNotRenderError(t *testing.T) {
	if err := cache.NewVidThumbCache(1); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "clip.mp4"), []byte("cached video"), 0644); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cache.GetVidThumbCache().Set("/clip.mp4", image.NewRGBA(image.Rect(0, 0, 2, 2)), 4)
	cache.GetVidThumbCache().Wait()
	w := &failingResponseWriter{}
	r := httptest.NewRequest("GET", "/vidthumb/?path=/clip.mp4", nil)

	VidThumbHandler(w, r, "/clip.mp4", hfs, &config.Config{})

	if w.writeCalls != 1 {
		t.Errorf("response writes = %d, want 1", w.writeCalls)
	}
	if w.writeHeaderCalls != 0 {
		t.Errorf("explicit status writes = %d, want 0 after client disconnect", w.writeHeaderCalls)
	}
}
