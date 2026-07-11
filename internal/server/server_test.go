package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wsand02/heheserver/internal/config"
)

func setupTestDir(t *testing.T) string {
	tmpDir := t.TempDir() // automatically cleaned up at end of test

	// Create sample files
	err := os.WriteFile(filepath.Join(tmpDir, "hello.txt"), []byte("world"), 0644)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = os.WriteFile(filepath.Join(tmpDir, "image.jpg"), []byte("fake image data"), 0644)
	if err != nil {
		t.Fatal(err.Error())
	}

	return tmpDir
}

func testServer(t *testing.T, dir string, gallery, resize bool) *httptest.Server {
	cfg, err := config.NewConfig(0, 64, gallery, resize, dir, "localhost", 1, 1000, 1000)
	if err != nil {
		t.Fatal(err.Error())
	}
	srv := NewServer(cfg)
	ts := httptest.NewServer(srv.mux) // or srv.Start() but httptest is easier
	t.Cleanup(ts.Close)
	return ts
}

func TestFileServerMode(t *testing.T) {
	dir := setupTestDir(t)

	ts := testServer(t, dir, false, false)
	client := ts.Client()

	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		t.Fatal("Status not OK")
	}

	resp, err = client.Get(ts.URL + "/hello.txt")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "world" {
		t.Fatal("body != world")
	}
}

func TestGalleryServerMode(t *testing.T) {
	dir := setupTestDir(t)

	ts := testServer(t, dir, true, false)
	client := ts.Client()

	resp, err := client.Get(ts.URL + "/?path=/")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		t.Fatal("Status not OK")
	}

	resp, err = client.Get(ts.URL + "/fs/image.jpg")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		t.Fatal("Status not OK")
	}

	resp, err = client.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		t.Fatal("Status not OK")
	}

	resp, err = client.Get(ts.URL + "/post/?path=/image.jpg")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		t.Fatal("Status not OK")
	}
}

// TestGalleryStaticAssets covers issue #16: the gallery must serve its CSS from the
// embedded /static/ route so styling works with no internet access.
func TestGalleryStaticAssets(t *testing.T) {
	dir := setupTestDir(t)

	ts := testServer(t, dir, true, false)
	client := ts.Client()

	for _, name := range []string{"glacialwisp.min.css", "glacialwisp-icons.min.css"} {
		resp, err := client.Get(ts.URL + "/static/" + name)
		if err != nil {
			t.Fatal(err.Error())
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /static/%s => %d, want 200", name, resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
			t.Fatalf("GET /static/%s content-type = %q, want text/css", name, ct)
		}
		if len(body) == 0 {
			t.Fatalf("GET /static/%s returned empty body", name)
		}
	}
}

// TestGallerySpecialCharFilenames reproduces issue #13: files whose names contain
// emoji or URL-significant characters must be reachable through the gallery.
func TestGallerySpecialCharFilenames(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"😀.jpg", "c++.png", "a b.png"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644); err != nil {
			t.Fatal(err.Error())
		}
	}

	ts := testServer(t, dir, true, false)
	client := ts.Client()

	// The gallery page must not emit raw special chars into the links.
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err.Error())
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	for _, raw := range []string{"path=/c++.png", "path=/a b.png"} {
		if strings.Contains(string(body), raw) {
			t.Fatalf("gallery emitted unescaped link containing %q", raw)
		}
	}

	// Each generated link must resolve to a real file (200), not 404.
	cases := []string{
		"/post/?path=/%F0%9F%98%80.jpg",
		"/fs/%F0%9F%98%80.jpg",
		"/post/?path=/c%2B%2B.png",
		"/fs/c%2B%2B.png",
		"/post/?path=/a%20b.png",
		"/fs/a%20b.png",
	}
	for _, u := range cases {
		resp, err := client.Get(ts.URL + u)
		if err != nil {
			t.Fatal(err.Error())
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s => %d, want 200", u, resp.StatusCode)
		}
	}
}
