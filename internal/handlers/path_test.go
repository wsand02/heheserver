package handlers

import (
	"bytes"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/utils"
)

type forbiddenFS struct{ t *testing.T }

func (f forbiddenFS) Open(name string) (http.File, error) {
	f.t.Fatalf("invalid path reached filesystem: %q", name)
	return nil, nil
}

func TestThumbnailInvalidPaths(t *testing.T) {
	if err := cache.NewResizeCache(1); err != nil {
		t.Fatal(err)
	}
	if err := cache.NewVidThumbCache(1); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"/../private.png", "/sub/../../private.png", `/sub\private.png`, "/bad\x00.png"} {
		for name, handler := range map[string]func(http.ResponseWriter, *http.Request, string, *fs.HeheFS, *config.Config){"resize": ResizeHandler, "video": VidThumbHandler} {
			t.Run(name+url.QueryEscape(raw), func(t *testing.T) {
				hfs := &fs.HeheFS{Root: t.TempDir(), FileSystem: forbiddenFS{t}}
				// A cached value must never let a malformed path bypass validation.
				cache.GetResizeCache().Set(raw, cache.ResizeCacheItem{JPEG: []byte("cached")}, 1)
				cache.GetResizeCache().Wait()
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/"+name+"/?"+url.Values{"path": {raw}}.Encode(), nil)
				handler(w, r, raw, hfs, &config.Config{FFmpegExists: true})
				if w.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want 400", w.Code)
				}
			})
		}
	}
}

func TestThumbnailSpecialCharacterPaths(t *testing.T) {
	for _, ffmpeg := range []bool{false, true} {
		t.Run(map[bool]string{false: "fallback", true: "ffmpeg"}[ffmpeg], func(t *testing.T) {
			if ffmpeg && !utils.FFmpegExists() {
				t.Skip("ffmpeg not on PATH")
			}
			if err := cache.NewResizeCache(1); err != nil {
				t.Fatal(err)
			}
			if err := cache.NewVidThumbCache(1); err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			hfs := testHFS(t, root)
			imageName := "photo #+ å.png"
			writePNG(t, filepath.Join(root, imageName), 600, 400)
			cfg := &config.Config{FFmpegExists: ffmpeg}
			check := func(route, raw string, handler func(http.ResponseWriter, *http.Request, string, *fs.HeheFS, *config.Config)) {
				t.Helper()
				r := httptest.NewRequest("GET", route+"?"+url.Values{"path": {raw}}.Encode(), nil)
				w := httptest.NewRecorder()
				handler(w, r, r.URL.Query().Get("path"), hfs, cfg)
				if w.Code != http.StatusOK {
					t.Fatalf("status %d: %s", w.Code, w.Body.String())
				}
				if _, err := jpeg.Decode(bytes.NewReader(w.Body.Bytes())); err != nil {
					t.Fatal(err)
				}
			}
			check("/resize/", "/./"+imageName, ResizeHandler)
			cache.GetResizeCache().Wait()
			if _, ok := cache.GetResizeCache().Get("/" + imageName); !ok {
				t.Fatal("resize cache key was not normalized")
			}
			if ffmpeg {
				videoName := "clip #+ å.mp4"
				makeTestVideo(t, filepath.Join(root, videoName))
				check("/vidthumb/", "/./"+videoName, VidThumbHandler)
				cache.GetVidThumbCache().Wait()
				if _, ok := cache.GetVidThumbCache().Get("/" + videoName); !ok {
					t.Fatal("video cache key was not normalized")
				}
			}
		})
	}
}
