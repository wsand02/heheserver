package handlers

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
	"github.com/wsand02/heheserver/internal/config"
	"github.com/wsand02/heheserver/internal/fs"
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
