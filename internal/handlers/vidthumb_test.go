package handlers

import (
	"bytes"
	"image/jpeg"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/utils"
)

func makeTestVideo(t *testing.T, path string) {
	t.Helper()
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", "testsrc=duration=2:size=64x64:rate=10",
		"-pix_fmt", "yuv420p", path)
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
