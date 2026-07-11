package handlers

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
	"github.com/wsand02/heheserver/internal/models"
)

func TestGetBreadcrumbs(t *testing.T) {
	cases := []struct {
		path string
		want []string
	}{
		{"/", nil},
		{"/foo/bar/", []string{"foo", "bar"}},
		{"/a", []string{"a"}},
	}
	for _, c := range cases {
		gc := &GalleryContext{Path: c.path}
		if got := gc.GetBreadcrumbs(); !reflect.DeepEqual(got, c.want) {
			t.Errorf("GetBreadcrumbs(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestBreadcrumbToUrl(t *testing.T) {
	gc := &GalleryContext{Path: "/foo/bar/"}
	if got := gc.BreadcrumbToUrl(0); got != "?path=/foo/" {
		t.Errorf("BreadcrumbToUrl(0) = %q, want %q", got, "?path=/foo/")
	}
	if got := gc.BreadcrumbToUrl(1); got != "?path=/foo/bar/" {
		t.Errorf("BreadcrumbToUrl(1) = %q, want %q", got, "?path=/foo/bar/")
	}

	// Special characters in a segment must be query-escaped.
	special := &GalleryContext{Path: "/a b/"}
	if got := special.BreadcrumbToUrl(0); got != "?path=/a+b/" {
		t.Errorf("BreadcrumbToUrl(0) = %q, want %q", got, "?path=/a+b/")
	}
}

// testHFS builds a HeheFS rooted at dir. The ignore cache must be initialized
// because hfs.Open consults it.
func testHFS(t *testing.T, dir string) *fs.HeheFS {
	t.Helper()
	if err := cache.NewIgnoreCache(1); err != nil {
		t.Fatal(err)
	}
	hfs, ok := fs.Dir(dir).(*fs.HeheFS)
	if !ok {
		t.Fatal("fs.Dir did not return *HeheFS")
	}
	return hfs
}

func TestGalleryHandler(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cfg := &config.Config{Split: 64}

	cases := []struct {
		name       string
		target     string
		wantStatus int
	}{
		{"valid", "/?path=/", 200},
		{"bad page (not a number)", "/?path=/&p=abc", 400},
		{"page out of range", "/?path=/&p=999", 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", c.target, nil)
			GalleryHandler(w, r, "/", hfs, cfg)
			if w.Code != c.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, c.wantStatus, w.Body.String())
			}
		})
	}
}

func TestGalleryHandlerEmptyDir(t *testing.T) {
	hfs := testHFS(t, t.TempDir())
	cfg := &config.Config{Split: 64}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?path=/", nil)
	GalleryHandler(w, r, "/", hfs, cfg)
	if w.Code != 200 {
		t.Errorf("empty dir: status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
}

// mixedDir creates a temp dir with one file of each relevant type and returns it.
func mixedDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"cat.png", "dog.jpg", "clip.mp4", "song.ogg", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// itemNames renders the gallery for target and returns which of the known
// filenames appear in the output body.
func galleryBody(t *testing.T, hfs *fs.HeheFS, cfg *config.Config, target string) (int, string) {
	t.Helper()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", target, nil)
	GalleryHandler(w, r, "/", hfs, cfg)
	return w.Code, w.Body.String()
}

func TestGalleryHandlerFilter(t *testing.T) {
	hfs := testHFS(t, mixedDir(t))
	cfg := &config.Config{Split: 64}

	cases := []struct {
		name       string
		target     string
		wantIn     []string
		wantNotIn  []string
	}{
		{
			name:      "type image",
			target:    "/?path=/&type=image",
			wantIn:    []string{"cat.png", "dog.jpg"},
			wantNotIn: []string{"clip.mp4", "song.ogg", "notes.txt", ">sub<"},
		},
		{
			name:      "type video and audio",
			target:    "/?path=/&type=video&type=audio",
			wantIn:    []string{"clip.mp4", "song.ogg"},
			wantNotIn: []string{"cat.png", "notes.txt"},
		},
		{
			name:      "query substring",
			target:    "/?path=/&q=cat",
			wantIn:    []string{"cat.png"},
			wantNotIn: []string{"dog.jpg", "clip.mp4"},
		},
		{
			name:      "extension",
			target:    "/?path=/&ext=mp4",
			wantIn:    []string{"clip.mp4"},
			wantNotIn: []string{"cat.png", "song.ogg"},
		},
		{
			name:      "AND image + query",
			target:    "/?path=/&type=image&q=dog",
			wantIn:    []string{"dog.jpg"},
			wantNotIn: []string{"cat.png", "clip.mp4"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, body := galleryBody(t, hfs, cfg, c.target)
			if code != 200 {
				t.Fatalf("status = %d, want 200 (body: %s)", code, body)
			}
			for _, w := range c.wantIn {
				if !strings.Contains(body, w) {
					t.Errorf("expected %q in body", w)
				}
			}
			for _, w := range c.wantNotIn {
				if strings.Contains(body, w) {
					t.Errorf("did not expect %q in body", w)
				}
			}
		})
	}
}

// TestGalleryHandlerFilterEmptyResult guards the fixed fallback bug: a filter
// that matches nothing must render an empty gallery, NOT the whole directory.
func TestGalleryHandlerFilterEmptyResult(t *testing.T) {
	dir := t.TempDir()
	// image-only directory
	if err := os.WriteFile(filepath.Join(dir, "cat.png"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cfg := &config.Config{Split: 64}

	code, body := galleryBody(t, hfs, cfg, "/?path=/&type=video")
	if code != 200 {
		t.Fatalf("status = %d, want 200 (body: %s)", code, body)
	}
	if strings.Contains(body, "cat.png") {
		t.Errorf("empty filter result fell back to showing unfiltered files (found cat.png)")
	}
}

// TestGalleryPageURLKeepsFilter checks pagination links carry the active filter.
func TestGalleryPageURLKeepsFilter(t *testing.T) {
	gc := &GalleryContext{
		Path:        "/pics/",
		FilterQuery: "cat",
		FilterExt:   "png",
		Filter:      models.GalleryFilter{Types: map[string]bool{"image": true}},
	}
	got := gc.PageURL(2)
	for _, want := range []string{"path=%2Fpics%2F", "p=2", "type=image", "q=cat", "ext=png"} {
		if !strings.Contains(got, want) {
			t.Errorf("PageURL(2) = %q, missing %q", got, want)
		}
	}
}

func TestPostContextGalleryURL(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/folder/file.png", "?path=%2Ffolder%2F"},
		{"/file.png", "?path=%2F"},
		{"/a/b/c.mp4", "?path=%2Fa%2Fb%2F"},
	}
	for _, c := range cases {
		pc := &PostContext{models.GalleryItem{Path: c.path}}
		if got := pc.GalleryURL(); got != c.want {
			t.Errorf("GalleryURL(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestPostHandler(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	hfs := testHFS(t, dir)
	cfg := &config.Config{Split: 64}

	// A file path renders successfully.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/post/?path=/hello.txt", nil)
	PostHandler(w, r, "/hello.txt", hfs, cfg)
	if w.Code != 200 {
		t.Errorf("file: status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	// A directory path is rejected.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/post/?path=/", nil)
	PostHandler(w, r, "/", hfs, cfg)
	if w.Code != 400 {
		t.Errorf("directory: status = %d, want 400", w.Code)
	}
}
