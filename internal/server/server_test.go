package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	cfg := &config.Config{
		Directory: dir,
		Gallery:   gallery,
		Resize:    resize,
		Host:      "localhost",
		Port:      0, // will be overridden by httptest
		Split:     64,
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
